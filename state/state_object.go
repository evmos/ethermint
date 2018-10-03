package state

import (
	"bytes"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	_ ethstate.StateObject = (*stateObject)(nil)

	emptyCodeHash = crypto.Keccak256(nil)
)

type (
	// stateObject represents an Ethereum account which is being modified.
	//
	// The usage pattern is as follows:
	// First you need to obtain a state object.
	// Account values can be accessed and modified through the object.
	// Finally, call CommitTrie to write the modified storage trie into a database.
	stateObject struct {
		address ethcmn.Address
		stateDB *CommitStateDB
		account *types.Account

		// DB error.
		// State objects are used by the consensus core and VM which are
		// unable to deal with database-level errors. Any error that occurs
		// during a database read is memoized here and will eventually be returned
		// by StateDB.Commit.
		dbErr error

		code types.Code // contract bytecode, which gets set when code is loaded

		originStorage types.Storage // Storage cache of original entries to dedup rewrites
		dirtyStorage  types.Storage // Storage entries that need to be flushed to disk

		// cache flags
		//
		// When an object is marked suicided it will be delete from the trie during
		// the "update" phase of the state transition.
		dirtyCode bool // true if the code was updated
		suicided  bool
		deleted   bool
	}
)

func newObject(db *CommitStateDB, accProto auth.Account) *stateObject {
	acc, ok := accProto.(*types.Account)
	if !ok {
		panic(fmt.Sprintf("invalid account type for state object: %T", acc))
	}

	if acc.CodeHash == nil {
		acc.CodeHash = emptyCodeHash
	}

	return &stateObject{
		stateDB:       db,
		account:       acc,
		address:       ethcmn.BytesToAddress(acc.Address.Bytes()),
		originStorage: make(types.Storage),
		dirtyStorage:  make(types.Storage),
	}
}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetState updates a value in account storage. Note, the key must be prefixed
// with the address.
func (so *stateObject) SetState(db ethstate.Database, key, value ethcmn.Hash) {
	// if the new value is the same as old, don't set
	prev := so.GetState(db, key)
	if prev == value {
		return
	}

	// since the new value is different, update and journal the change
	so.stateDB.journal.append(storageChange{
		account:  &so.address,
		key:      key,
		prevalue: prev,
	})

	so.setState(key, value)
}

func (so *stateObject) setState(key, value ethcmn.Hash) {
	so.dirtyStorage[key] = value
}

// SetCode sets the state object's code.
func (so *stateObject) SetCode(codeHash ethcmn.Hash, code []byte) {
	prevcode := so.Code(nil)

	so.stateDB.journal.append(codeChange{
		account:  &so.address,
		prevhash: so.CodeHash(),
		prevcode: prevcode,
	})

	so.setCode(codeHash, code)
}

func (so *stateObject) setCode(codeHash ethcmn.Hash, code []byte) {
	so.code = code
	so.account.CodeHash = codeHash.Bytes()
	so.dirtyCode = true
}

// AddBalance adds an amount to a state object's balance. It is used to add
// funds to the destination account of a transfer.
func (so *stateObject) AddBalance(amount *big.Int) {
	amt := sdk.NewIntFromBigInt(amount)

	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amt.Sign() == 0 {
		if so.empty() {
			so.touch()
		}

		return
	}

	newBalance := so.account.Balance().Add(amt)
	so.SetBalance(newBalance.BigInt())
}

// SubBalance removes an amount from the stateObject's balance. It is used to
// remove funds from the origin account of a transfer.
func (so *stateObject) SubBalance(amount *big.Int) {
	amt := sdk.NewIntFromBigInt(amount)

	if amt.Sign() == 0 {
		return
	}

	newBalance := so.account.Balance().Sub(amt)
	so.SetBalance(newBalance.BigInt())
}

// SetBalance sets the state object's balance.
func (so *stateObject) SetBalance(amount *big.Int) {
	amt := sdk.NewIntFromBigInt(amount)

	so.stateDB.journal.append(balanceChange{
		account: &so.address,
		prev:    so.account.Balance(),
	})

	so.setBalance(amt)
}

func (so *stateObject) setBalance(amount sdk.Int) {
	so.account.SetBalance(amount)
}

// SetNonce sets the state object's nonce (sequence number).
func (so *stateObject) SetNonce(nonce uint64) {
	so.stateDB.journal.append(nonceChange{
		account: &so.address,
		prev:    so.account.Sequence,
	})

	so.setNonce(int64(nonce))
}

func (so *stateObject) setNonce(nonce int64) {
	so.account.Sequence = nonce
}

// setError remembers the first non-nil error it is called with.
func (so *stateObject) setError(err error) {
	if so.dbErr == nil {
		so.dbErr = err
	}
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// Address returns the address of the state object.
func (so stateObject) Address() ethcmn.Address {
	return so.address
}

// Balance returns the state object's current balance.
func (so *stateObject) Balance() *big.Int {
	return so.account.Balance().BigInt()
}

// CodeHash returns the state object's code hash.
func (so *stateObject) CodeHash() []byte {
	return so.account.CodeHash
}

// Nonce returns the state object's current nonce (sequence number).
func (so *stateObject) Nonce() uint64 {
	return uint64(so.account.Sequence)
}

// Code returns the contract code associated with this object, if any.
func (so *stateObject) Code(_ ethstate.Database) []byte {
	if so.code != nil {
		return so.code
	}

	if bytes.Equal(so.CodeHash(), emptyCodeHash) {
		return nil
	}

	store := so.stateDB.ctx.KVStore(so.stateDB.codeKey)
	code := store.Get(so.CodeHash())

	if len(code) == 0 {
		so.setError(fmt.Errorf("failed to get code hash %x for address: %x", so.CodeHash(), so.Address()))
	}

	so.code = code
	return code
}

// GetState retrieves a value from the account storage trie. Note, the key must
// be prefixed with the address.
func (so *stateObject) GetState(_ ethstate.Database, key ethcmn.Hash) ethcmn.Hash {
	// if we have a dirty value for this state entry, return it
	value, dirty := so.dirtyStorage[key]
	if dirty {
		return value
	}

	// otherwise return the entry's original value
	return so.getCommittedState(so.stateDB.ctx, key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
// Note, the must be prefixed with the address.
func (so *stateObject) getCommittedState(ctx sdk.Context, key ethcmn.Hash) ethcmn.Hash {
	// if we have the original value cached, return that
	value, cached := so.originStorage[key]
	if cached {
		return value
	}

	// otherwise load the value from the KVStore
	store := ctx.KVStore(so.stateDB.storageKey)
	rawValue := store.Get(key.Bytes())

	if len(rawValue) > 0 {
		value.SetBytes(rawValue)
	}

	so.originStorage[key] = value
	return value
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// ReturnGas returns the gas back to the origin. Used by the Virtual machine or
// Closures. It performs a no-op.
func (so *stateObject) ReturnGas(gas *big.Int) {}

// empty returns whether the account is considered empty.
func (so *stateObject) empty() bool {
	return so.account.Sequence == 0 &&
		so.account.Balance().Sign() == 0 &&
		bytes.Equal(so.account.CodeHash, emptyCodeHash)
}

func (so *stateObject) touch() {
	so.stateDB.journal.append(touchChange{
		account: &so.address,
	})

	if so.address == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		so.stateDB.journal.dirty(so.address)
	}
}

// prefixStorageKey prefixes a storage key with the state object's address.
func (so stateObject) prefixStorageKey(key []byte) []byte {
	prefix := so.Address().Bytes()
	compositeKey := make([]byte, len(prefix)+len(key))

	copy(compositeKey, prefix)
	copy(compositeKey[len(prefix):], key)

	return compositeKey
}

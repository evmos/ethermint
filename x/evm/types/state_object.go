package types

import (
	"bytes"
	"fmt"
	"io"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	_ StateObject = (*stateObject)(nil)

	emptyCodeHash = ethcrypto.Keccak256(nil)
)

type (
	// StateObject interface for interacting with state object
	StateObject interface {
		GetCommittedState(db ethstate.Database, key ethcmn.Hash) ethcmn.Hash
		GetState(db ethstate.Database, key ethcmn.Hash) ethcmn.Hash
		SetState(db ethstate.Database, key, value ethcmn.Hash)

		Code(db ethstate.Database) []byte
		SetCode(codeHash ethcmn.Hash, code []byte)
		CodeHash() []byte

		AddBalance(amount *big.Int)
		SubBalance(amount *big.Int)
		SetBalance(amount *big.Int)

		Balance() *big.Int
		ReturnGas(gas *big.Int)
		Address() ethcmn.Address

		SetNonce(nonce uint64)
		Nonce() uint64
	}
	// stateObject represents an Ethereum account which is being modified.
	//
	// The usage pattern is as follows:
	// First you need to obtain a state object.
	// Account values can be accessed and modified through the object.
	// Finally, call CommitTrie to write the modified storage trie into a database.
	stateObject struct {
		code types.Code // contract bytecode, which gets set when code is loaded
		// DB error.
		// State objects are used by the consensus core and VM which are
		// unable to deal with database-level errors. Any error that occurs
		// during a database read is memoized here and will eventually be returned
		// by StateDB.Commit.
		dbErr         error
		stateDB       *CommitStateDB
		account       *types.Account
		originStorage types.Storage // Storage cache of original entries to dedup rewrites
		dirtyStorage  types.Storage // Storage entries that need to be flushed to disk
		address       ethcmn.Address
		// cache flags
		//
		// When an object is marked suicided it will be delete from the trie during
		// the "update" phase of the state transition.
		dirtyCode bool // true if the code was updated
		suicided  bool
		deleted   bool
	}
)

func newStateObject(db *CommitStateDB, accProto authexported.Account) *stateObject {
	ethermintAccount, ok := accProto.(*types.Account)
	if !ok {
		panic(fmt.Sprintf("invalid account type for state object: %T", accProto))
	}

	// set empty code hash
	if ethermintAccount.CodeHash == nil {
		ethermintAccount.CodeHash = emptyCodeHash
	}

	return &stateObject{
		stateDB:       db,
		account:       ethermintAccount,
		address:       ethcmn.BytesToAddress(ethermintAccount.GetAddress().Bytes()),
		originStorage: make(types.Storage),
		dirtyStorage:  make(types.Storage),
	}
}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetState updates a value in account storage. Note, the key will be prefixed
// with the address of the state object.
func (so *stateObject) SetState(db ethstate.Database, key, value ethcmn.Hash) {
	// if the new value is the same as old, don't set
	prev := so.GetState(db, key)
	if prev == value {
		return
	}

	prefixKey := so.GetStorageByAddressKey(key.Bytes())

	// since the new value is different, update and journal the change
	so.stateDB.journal.append(storageChange{
		account:   &so.address,
		key:       prefixKey,
		prevValue: prev,
	})

	so.setState(prefixKey, value)
}

func (so *stateObject) setState(key, value ethcmn.Hash) {
	so.dirtyStorage[key] = value
}

// SetCode sets the state object's code.
func (so *stateObject) SetCode(codeHash ethcmn.Hash, code []byte) {
	prevCode := so.Code(nil)

	so.stateDB.journal.append(codeChange{
		account:  &so.address,
		prevHash: so.CodeHash(),
		prevCode: prevCode,
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

// SetNonce sets the state object's nonce (i.e sequence number of the account).
func (so *stateObject) SetNonce(nonce uint64) {
	so.stateDB.journal.append(nonceChange{
		account: &so.address,
		prev:    so.account.Sequence,
	})

	so.setNonce(nonce)
}

func (so *stateObject) setNonce(nonce uint64) {
	if so.account == nil {
		panic("state object account is empty")
	}
	so.account.Sequence = nonce
}

// setError remembers the first non-nil error it is called with.
func (so *stateObject) setError(err error) {
	if so.dbErr == nil {
		so.dbErr = err
	}
}

func (so *stateObject) markSuicided() {
	so.suicided = true
}

// commitState commits all dirty storage to a KVStore.
func (so *stateObject) commitState() {
	ctx := so.stateDB.ctx
	store := ctx.KVStore(so.stateDB.storeKey)

	for key, value := range so.dirtyStorage {
		delete(so.dirtyStorage, key)

		// skip no-op changes, persist actual changes
		if value == so.originStorage[key] {
			continue
		}

		so.originStorage[key] = value

		// delete empty values
		if (value == ethcmn.Hash{}) {
			store.Delete(key.Bytes())
			continue
		}

		store.Set(key.Bytes(), value.Bytes())
	}

	// TODO: Set the account (storage) root (but we probably don't need this)
}

// commitCode persists the state object's code to the KVStore.
func (so *stateObject) commitCode() {
	ctx := so.stateDB.ctx
	store := ctx.KVStore(so.stateDB.codeKey)
	store.Set(so.CodeHash(), so.code)
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
	balance := so.account.Balance().BigInt()
	if balance == nil {
		return zeroBalance
	}
	return balance
}

// CodeHash returns the state object's code hash.
func (so *stateObject) CodeHash() []byte {
	if so.account == nil || len(so.account.CodeHash) == 0 {
		return emptyCodeHash
	}
	return so.account.CodeHash
}

// Nonce returns the state object's current nonce (sequence number).
func (so *stateObject) Nonce() uint64 {
	if so.account == nil {
		return 0
	}
	return so.account.Sequence
}

// Code returns the contract code associated with this object, if any.
func (so *stateObject) Code(_ ethstate.Database) []byte {
	if len(so.code) > 0 {
		return so.code
	}

	if bytes.Equal(so.CodeHash(), emptyCodeHash) {
		return nil
	}

	ctx := so.stateDB.ctx
	store := ctx.KVStore(so.stateDB.codeKey)
	code := store.Get(so.CodeHash())

	if len(code) == 0 {
		so.setError(fmt.Errorf("failed to get code hash %x for address %s", so.CodeHash(), so.Address().String()))
	}

	return code
}

// GetState retrieves a value from the account storage trie. Note, the key will
// be prefixed with the address of the state object.
func (so *stateObject) GetState(db ethstate.Database, key ethcmn.Hash) ethcmn.Hash {
	prefixKey := so.GetStorageByAddressKey(key.Bytes())

	// if we have a dirty value for this state entry, return it
	value, dirty := so.dirtyStorage[prefixKey]
	if dirty {
		return value
	}

	// otherwise return the entry's original value
	return so.GetCommittedState(db, key)
}

// GetCommittedState retrieves a value from the committed account storage trie.
// Note, the key will be prefixed with the address of the state object.
func (so *stateObject) GetCommittedState(_ ethstate.Database, key ethcmn.Hash) ethcmn.Hash {
	prefixKey := so.GetStorageByAddressKey(key.Bytes())

	// if we have the original value cached, return that
	value, cached := so.originStorage[prefixKey]
	if cached {
		return value
	}

	// otherwise load the value from the KVStore
	ctx := so.stateDB.ctx
	store := ctx.KVStore(so.stateDB.storeKey)
	rawValue := store.Get(prefixKey.Bytes())

	if len(rawValue) > 0 {
		value.SetBytes(rawValue)
	}

	so.originStorage[prefixKey] = value
	return value
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// ReturnGas returns the gas back to the origin. Used by the Virtual machine or
// Closures. It performs a no-op.
func (so *stateObject) ReturnGas(gas *big.Int) {}

func (so *stateObject) deepCopy(db *CommitStateDB) *stateObject {
	newStateObj := newStateObject(db, so.account)

	newStateObj.code = so.code
	newStateObj.dirtyStorage = so.dirtyStorage.Copy()
	newStateObj.originStorage = so.originStorage.Copy()
	newStateObj.suicided = so.suicided
	newStateObj.dirtyCode = so.dirtyCode
	newStateObj.deleted = so.deleted

	return newStateObj
}

// empty returns whether the account is considered empty.
func (so *stateObject) empty() bool {
	return so.account == nil ||
		(so.account != nil &&
			so.account.Sequence == 0 &&
			so.account.Balance().Sign() == 0 &&
			bytes.Equal(so.account.CodeHash, emptyCodeHash))
}

// EncodeRLP implements rlp.Encoder.
func (so *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, so.account)
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

// GetStorageByAddressKey returns a hash of the composite key for a state
// object's storage prefixed with it's address.
func (so stateObject) GetStorageByAddressKey(key []byte) ethcmn.Hash {
	prefix := so.Address().Bytes()
	compositeKey := make([]byte, len(prefix)+len(key))

	copy(compositeKey, prefix)
	copy(compositeKey[len(prefix):], key)

	return ethcrypto.Keccak256Hash(compositeKey)
}

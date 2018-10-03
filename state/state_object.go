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

	// // Account is the Ethereum consensus representation of accounts.
	// // These objects are stored in the main account trie.
	// Account struct {
	// 	Nonce    uint64
	// 	Balance  *big.Int
	// 	Root     ethcmn.Hash // merkle root of the storage trie
	// 	CodeHash []byte
	// }
)

func newObject(db *CommitStateDB, accProto auth.Account) *stateObject {
	// if acc.Balance == nil {
	// 	data.Balance = new(big.Int)
	// }
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

// Address returns the address of the state object.
func (so stateObject) Address() ethcmn.Address {
	return so.address
}

// GetState retrieves a value from the account storage trie.
func (so *stateObject) GetState(_ Database, key ethcmn.Hash) ethcmn.Hash {
	// if we have a dirty value for this state entry, return it
	value, dirty := so.dirtyStorage[key]
	if dirty {
		return value
	}

	// otherwise return the entry's original value
	return so.getCommittedState(key)
}

// SetState updates a value in account storage.
func (so *stateObject) SetState(db Database, key, value ethcmn.Hash) {
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

func (so *stateObject) SetBalance(amount *big.Int) {
	amt := sdk.NewIntFromBigInt(amount)

	so.stateDB.journal.append(balanceChange{
		account: &so.address,
		prev:    so.account.Balance(),
	})

	so.setBalance(amt)
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (so *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}

	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))
}

// func (so *stateObject) Balance() *big.Int {

// }

// func (so *stateObject) ReturnGas(gas *big.Int) {

// }

// func (so *stateObject) Address() ethcmn.Address {

// }

// func (so *stateObject) SetCode(codeHash ethcmn.Hash, code []byte) {

// }

// func (so *stateObject) SetNonce(nonce uint64) {

// }

// func (so *stateObject) Nonce() uint64 {

// }

// func (so *stateObject) Code(db Database) []byte {

// }

// func (so *stateObject) CodeHash() []byte {

// }

func (so *stateObject) setBalance(amount sdk.Int) {
	so.account.SetBalance(amount)
}

// GetCommittedState retrieves a value from the committed account storage trie.
func (so *stateObject) getCommittedState(key ethcmn.Hash) ethcmn.Hash {
	// if we have the original value cached, return that
	value, cached := so.originStorage[key]
	if cached {
		return value
	}

	// otherwise load the value from the KVStore
	store := so.stateDB.ctx.KVStore(so.stateDB.storageKey)
	rawValue := store.Get(key.Bytes())

	if len(rawValue) > 0 {
		value.SetBytes(rawValue)
	}

	so.originStorage[key] = value
	return value
}

func (so *stateObject) setState(key, value ethcmn.Hash) {
	so.dirtyStorage[key] = value
}

// setError remembers the first non-nil error it is called with.
func (so *stateObject) setError(err error) {
	if so.dbErr == nil {
		so.dbErr = err
	}
}

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

package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.StateDB = &StateDB{}

// StateDB implements vm.StateDB interface
type StateDB struct {
	keeper StateDBKeeper

	// Manage the initial context and cache context stack for accessing the store,
	// emit events and log info.
	// It is kept as a field to make is accessible by the StateDb
	// functions. Created in `ApplyMessage` on the fly.
	ctxStack ContextStack

	// error from previous state operation
	stateErr error
}

// NewStateDB construct a new StateDB
func NewStateDB(ctx sdk.Context, k StateDBKeeper) *StateDB {
	return &StateDB{
		keeper:   k,
		ctxStack: NewContextStack(ctx),
		stateErr: nil,
	}
}

// ----------------------------------------------------------------------------
// Account
// ----------------------------------------------------------------------------

// CreateAccount creates a new EthAccount instance from the provided address and
// sets the value to store. If an account with the given address already exists,
// this function also resets any preexisting code and storage associated with that
// address.
func (db *StateDB) CreateAccount(addr common.Address) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.CreateAccount(ctx, addr)
}

// ----------------------------------------------------------------------------
// Balance
// ----------------------------------------------------------------------------

// AddBalance adds the given amount to the address balance coin by minting new
// coins and transferring them to the address. The coin denomination is obtained
// from the module parameters.
func (db *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.stateErr = db.keeper.AddBalance(ctx, addr, amount)
}

// SubBalance subtracts the given amount from the address balance by transferring the
// coins to an escrow account and then burning them. The coin denomination is obtained
// from the module parameters. This function performs a no-op if the amount is negative
// or the user doesn't have enough funds for the transfer.
func (db *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.stateErr = db.keeper.SubBalance(ctx, addr, amount)
}

// GetBalance returns the EVM denomination balance of the provided address. The
// denomination is obtained from the module parameters.
func (db *StateDB) GetBalance(addr common.Address) *big.Int {
	if db.HasStateError() {
		return big.NewInt(0)
	}

	ctx := db.CurrentContext()
	return db.keeper.GetBalance(ctx, addr)
}

// ----------------------------------------------------------------------------
// Nonce
// ----------------------------------------------------------------------------

// GetNonce retrieves the account with the given address and returns the tx
// sequence (i.e nonce). The function performs a no-op if the account is not found.
func (db *StateDB) GetNonce(addr common.Address) uint64 {
	if db.HasStateError() {
		return 0
	}

	ctx := db.CurrentContext()
	return db.keeper.GetNonce(ctx, addr)
}

// SetNonce sets the given nonce as the sequence of the address' account. If the
// account doesn't exist, a new one will be created from the address.
func (db *StateDB) SetNonce(addr common.Address, nonce uint64) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.stateErr = db.keeper.SetNonce(ctx, addr, nonce)
}

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// GetCodeHash fetches the account from the store and returns its code hash. If the account doesn't
// exist or is not an EthAccount type, GetCodeHash returns the empty code hash value.
func (db *StateDB) GetCodeHash(addr common.Address) common.Hash {
	if db.HasStateError() {
		return common.Hash{}
	}

	ctx := db.CurrentContext()
	return db.keeper.GetCodeHash(ctx, addr)
}

// GetCode returns the code byte array associated with the given address.
// If the code hash from the account is empty, this function returns nil.
func (db *StateDB) GetCode(addr common.Address) []byte {
	if db.HasStateError() {
		return nil
	}

	ctx := db.CurrentContext()
	return db.keeper.GetCode(ctx, addr)
}

// SetCode stores the code byte array to the application KVStore and sets the
// code hash to the given account. The code is deleted from the store if it is empty.
func (db *StateDB) SetCode(addr common.Address, code []byte) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.stateErr = db.keeper.SetCode(ctx, addr, code)
}

// GetCodeSize returns the size of the contract code associated with this object,
// or zero if none.
func (db *StateDB) GetCodeSize(addr common.Address) int {
	if db.HasStateError() {
		return 0
	}

	ctx := db.CurrentContext()
	return db.keeper.GetCodeSize(ctx, addr)
}

// ----------------------------------------------------------------------------
// Refund
// ----------------------------------------------------------------------------

// NOTE: gas refunded needs to be tracked and stored in a separate variable in
// order to add it subtract/add it from/to the gas used value after the EVM
// execution has finalized. The refund value is cleared on every transaction and
// at the end of every block.

// AddRefund adds the given amount of gas to the refund transient value.
func (db *StateDB) AddRefund(gas uint64) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.AddRefund(ctx, gas)
}

// SubRefund subtracts the given amount of gas from the transient refund value. This function
// will panic if gas amount is greater than the stored refund.
func (db *StateDB) SubRefund(gas uint64) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.SubRefund(ctx, gas)
}

// GetRefund returns the amount of gas available for return after the tx execution
// finalizes. This value is reset to 0 on every transaction.
func (db *StateDB) GetRefund() uint64 {
	if db.HasStateError() {
		return 0
	}

	ctx := db.CurrentContext()
	return db.keeper.GetRefund(ctx)
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// GetCommittedState returns the value set in store for the given key hash. If the key is not registered
// this function returns the empty hash.
func (db *StateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	if db.HasStateError() {
		return common.Hash{}
	}

	return db.keeper.GetState(db.ctxStack.initialCtx, addr, hash)
}

// GetState returns the committed state for the given key hash, as all changes are committed directly
// to the KVStore.
func (db *StateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	if db.HasStateError() {
		return common.Hash{}
	}

	ctx := db.CurrentContext()
	return db.keeper.GetState(ctx, addr, hash)
}

// SetState sets the given hashes (key, value) to the KVStore. If the value hash is empty, this
// function deletes the key from the store.
func (db *StateDB) SetState(addr common.Address, key, value common.Hash) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.SetState(ctx, addr, key, value)
}

// ----------------------------------------------------------------------------
// Suicide
// ----------------------------------------------------------------------------

// Suicide marks the given account as suicided and clears the account balance of
// the EVM tokens.
func (db *StateDB) Suicide(addr common.Address) (result bool) {
	if db.HasStateError() {
		return false
	}

	ctx := db.CurrentContext()
	result, db.stateErr = db.keeper.Suicide(ctx, addr)
	return
}

// HasSuicided queries the transient store to check if the account has been marked as suicided in the
// current block. Accounts that are suicided will be returned as non-nil during queries and "cleared"
// after the block has been committed.
func (db *StateDB) HasSuicided(addr common.Address) bool {
	if db.HasStateError() {
		return false
	}

	ctx := db.CurrentContext()
	return db.keeper.HasSuicided(ctx, addr)
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist returns true if the given account exists in store or if it has been
// marked as suicided in the transient store.
func (db *StateDB) Exist(addr common.Address) bool {
	if db.HasStateError() {
		return false
	}

	ctx := db.CurrentContext()
	return db.keeper.Exist(ctx, addr)
}

// Empty returns true if the address meets the following conditions:
// 	- nonce is 0
// 	- balance amount for evm denom is 0
// 	- account code hash is empty
//
// Non-ethereum accounts are considered not empty
func (db *StateDB) Empty(addr common.Address) bool {
	if db.HasStateError() {
		return false
	}

	ctx := db.CurrentContext()
	return db.keeper.Empty(ctx, addr)
}

// ----------------------------------------------------------------------------
// Access List
// ----------------------------------------------------------------------------

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// 	- Add sender to access list (2929)
// 	- Add destination to access list (2929)
// 	- Add precompiles to access list (2929)
// 	- Add the contents of the optional tx access list (2930)
//
// This method should only be called if Yolov3/Berlin/2929+2930 is applicable at the current number.
func (db *StateDB) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.PrepareAccessList(ctx, sender, dest, precompiles, txAccesses)
}

// AddressInAccessList returns true if the address is registered on the transient store.
func (db *StateDB) AddressInAccessList(addr common.Address) bool {
	if db.HasStateError() {
		return false
	}

	ctx := db.CurrentContext()
	return db.keeper.AddressInAccessList(ctx, addr)
}

// SlotInAccessList checks if the address and the slots are registered in the transient store
func (db *StateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk, slotOk bool) {
	if db.HasStateError() {
		return false, false
	}

	ctx := db.CurrentContext()
	return db.keeper.SlotInAccessList(ctx, addr, slot)
}

// AddAddressToAccessList adds the given address to the access list. If the address is already
// in the access list, this function performs a no-op.
func (db *StateDB) AddAddressToAccessList(addr common.Address) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.AddAddressToAccessList(ctx, addr)
}

// AddSlotToAccessList adds the given (address, slot) to the access list. If the address and slot are
// already in the access list, this function performs a no-op.
func (db *StateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.AddSlotToAccessList(ctx, addr, slot)
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot return the index in the cached context stack
func (db *StateDB) Snapshot() int {
	if db.HasStateError() {
		return 0
	}

	return db.ctxStack.Snapshot()
}

// RevertToSnapshot pop all the cached contexts after(including) the snapshot
func (db *StateDB) RevertToSnapshot(target int) {
	if db.HasStateError() {
		return
	}

	db.ctxStack.RevertToSnapshot(target)
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog appends the given ethereum Log to the list of Logs associated with the transaction hash kept in the current
// context. This function also fills in the tx hash, block hash, tx index and log index fields before setting the log
// to store.
func (db *StateDB) AddLog(log *ethtypes.Log) {
	if db.HasStateError() {
		return
	}

	ctx := db.CurrentContext()
	db.keeper.AddLog(ctx, log)
}

// ----------------------------------------------------------------------------
// Trie
// ----------------------------------------------------------------------------

// AddPreimage performs a no-op since the EnablePreimageRecording flag is disabled
// on the vm.Config during state transitions. No store trie preimages are written
// to the database.
func (db *StateDB) AddPreimage(_ common.Hash, _ []byte) {}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// ForEachStorage uses the store iterator to iterate over all the state keys and perform a callback
// function on each of them.
func (db *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	if db.HasStateError() {
		return db.stateErr
	}

	ctx := db.CurrentContext()
	return db.keeper.ForEachStorage(ctx, addr, cb)
}

// HasStateError return the previous error for any state operations
func (db *StateDB) HasStateError() bool {
	return db.stateErr != nil
}

// ClearStateError reset the previous state operation error to nil
func (db *StateDB) ClearStateError() {
	db.stateErr = nil
}

// Commit the intermediate contexts
func (db *StateDB) Commit() {
	db.ctxStack.Commit()
}

// CurrentContext return the current context
func (db *StateDB) CurrentContext() sdk.Context {
	return db.ctxStack.CurrentContext()
}

// Dirty returns if there's uncommitted state
func (db StateDB) Dirty() bool {
	return !db.ctxStack.IsEmpty()
}

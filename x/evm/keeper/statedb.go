package keeper

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

var _ vm.StateDB = &Keeper{}

// ----------------------------------------------------------------------------
// Account
// ----------------------------------------------------------------------------

// CreateAccount calls CommitStateDB.CreateAccount using the passed in context
func (k *Keeper) CreateAccount(addr common.Address) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Balance
// ----------------------------------------------------------------------------

// AddBalance calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(addr common.Address, amount *big.Int) {
	// TODO: implement
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(addr common.Address, amount *big.Int) {
	// TODO: implement
}

// GetBalance calls CommitStateDB.GetBalance using the passed in context
func (k *Keeper) GetBalance(addr common.Address) *big.Int {
	// TODO: implement

	return nil
}

// ----------------------------------------------------------------------------
// Nonce
// ----------------------------------------------------------------------------

// GetNonce calls CommitStateDB.GetNonce using the passed in context
func (k *Keeper) GetNonce(addr common.Address) uint64 {
	// TODO: implement

	return 0
}

// SetNonce calls CommitStateDB.SetNonce using the passed in context
func (k *Keeper) SetNonce(addr common.Address, nonce uint64) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Code
// ----------------------------------------------------------------------------

// GetCodeHash calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(addr common.Address) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// GetCode calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(addr common.Address) []byte {
	// TODO: implement
	return nil
}

// SetCode calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(addr common.Address, code []byte) {
	// TODO: implement
}

// GetCodeSize calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(addr common.Address) int {
	// TODO: implement
	return 0
}

// ----------------------------------------------------------------------------
// Refund
// ----------------------------------------------------------------------------

// AddRefund calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(gas uint64) {
	// TODO: implement
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(gas uint64) {
	// TODO: implement
}

// GetRefund calls CommitStateDB.GetRefund using the passed in context
func (k *Keeper) GetRefund() uint64 {
	// TODO: implement
	return 0
}

// ----------------------------------------------------------------------------
// State
// ----------------------------------------------------------------------------

// GetCommittedState calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// GetState calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(addr common.Address, hash common.Hash) common.Hash {
	// TODO: implement
	return common.Hash{}
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(addr common.Address, key, value common.Hash) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Suicide
// ----------------------------------------------------------------------------

// Suicide calls CommitStateDB.Suicide using the passed in context
func (k *Keeper) Suicide(addr common.Address) bool {
	// TODO: implement
	return false
}

// HasSuicided calls CommitStateDB.HasSuicided using the passed in context
func (k *Keeper) HasSuicided(addr common.Address) bool {
	// TODO: implement
	return false
}

// ----------------------------------------------------------------------------
// Account Exist / Empty
// ----------------------------------------------------------------------------

// Exist calls CommitStateDB.Exist using the passed in context
func (k *Keeper) Exist(addr common.Address) bool {
	// TODO: implement
	return false
}

// Empty calls CommitStateDB.Empty using the passed in context
func (k *Keeper) Empty(addr common.Address) bool {
	// TODO: implement
	return false
}

// ----------------------------------------------------------------------------
// Access List
// ----------------------------------------------------------------------------

func (k *Keeper) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList) {
	// TODO: implement
}

func (k *Keeper) AddressInAccessList(addr common.Address) bool {
	// TODO: implement
	return false
}

func (k *Keeper) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	// TODO: implement
	return false, false
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddAddressToAccessList(addr common.Address) {
	// TODO: implement
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (k *Keeper) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot() int {
	// TODO: implement
	return 0
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(revID int) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Log
// ----------------------------------------------------------------------------

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(log *ethtypes.Log) {
	// TODO: implement
}

// AddPreimage calls CommitStateDB.AddPreimage using the passed in context
func (k *Keeper) AddPreimage(hash common.Hash, preimage []byte) {
	// TODO: implement
}

// ----------------------------------------------------------------------------
// Iterator
// ----------------------------------------------------------------------------

// ForEachStorage calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) error {
	// TODO: implement
	return nil
}

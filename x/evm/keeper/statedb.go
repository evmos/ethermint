package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
)

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetBalance calls CommitStateDB.SetBalance using the passed in context
func (k *Keeper) SetBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.CommitStateDB.WithContext(ctx).SetBalance(addr, amount)
}

// AddBalance calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.CommitStateDB.WithContext(ctx).AddBalance(addr, amount)
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.CommitStateDB.WithContext(ctx).SubBalance(addr, amount)
}

// SetNonce calls CommitStateDB.SetNonce using the passed in context
func (k *Keeper) SetNonce(ctx sdk.Context, addr ethcmn.Address, nonce uint64) {
	k.CommitStateDB.WithContext(ctx).SetNonce(addr, nonce)
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(ctx sdk.Context, addr ethcmn.Address, key, value ethcmn.Hash) {
	k.CommitStateDB.WithContext(ctx).SetState(addr, key, value)
}

// SetCode calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(ctx sdk.Context, addr ethcmn.Address, code []byte) {
	k.CommitStateDB.WithContext(ctx).SetCode(addr, code)
}

// SetLogs calls CommitStateDB.SetLogs using the passed in context
func (k *Keeper) SetLogs(ctx sdk.Context, hash ethcmn.Hash, logs []*ethtypes.Log) error {
	return k.CommitStateDB.WithContext(ctx).SetLogs(hash, logs)
}

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(ctx sdk.Context, log *ethtypes.Log) {
	k.CommitStateDB.WithContext(ctx).AddLog(log)
}

// AddPreimage calls CommitStateDB.AddPreimage using the passed in context
func (k *Keeper) AddPreimage(ctx sdk.Context, hash ethcmn.Hash, preimage []byte) {
	k.CommitStateDB.WithContext(ctx).AddPreimage(hash, preimage)
}

// AddRefund calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(ctx sdk.Context, gas uint64) {
	k.CommitStateDB.WithContext(ctx).AddRefund(gas)
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(ctx sdk.Context, gas uint64) {
	k.CommitStateDB.WithContext(ctx).SubRefund(gas)
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// GetBalance calls CommitStateDB.GetBalance using the passed in context
func (k *Keeper) GetBalance(ctx sdk.Context, addr ethcmn.Address) *big.Int {
	return k.CommitStateDB.WithContext(ctx).GetBalance(addr)
}

// GetNonce calls CommitStateDB.GetNonce using the passed in context
func (k *Keeper) GetNonce(ctx sdk.Context, addr ethcmn.Address) uint64 {
	return k.CommitStateDB.WithContext(ctx).GetNonce(addr)
}

// TxIndex calls CommitStateDB.TxIndex using the passed in context
func (k *Keeper) TxIndex(ctx sdk.Context) int {
	return k.CommitStateDB.WithContext(ctx).TxIndex()
}

// BlockHash calls CommitStateDB.BlockHash using the passed in context
func (k *Keeper) BlockHash(ctx sdk.Context) ethcmn.Hash {
	return k.CommitStateDB.WithContext(ctx).BlockHash()
}

// GetCode calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(ctx sdk.Context, addr ethcmn.Address) []byte {
	return k.CommitStateDB.WithContext(ctx).GetCode(addr)
}

// GetCodeSize calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(ctx sdk.Context, addr ethcmn.Address) int {
	return k.CommitStateDB.WithContext(ctx).GetCodeSize(addr)
}

// GetCodeHash calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(ctx sdk.Context, addr ethcmn.Address) ethcmn.Hash {
	return k.CommitStateDB.WithContext(ctx).GetCodeHash(addr)
}

// GetState calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.CommitStateDB.WithContext(ctx).GetState(addr, hash)
}

// GetCommittedState calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.CommitStateDB.WithContext(ctx).GetCommittedState(addr, hash)
}

// GetLogs calls CommitStateDB.GetLogs using the passed in context
func (k *Keeper) GetLogs(ctx sdk.Context, hash ethcmn.Hash) ([]*ethtypes.Log, error) {
	return k.CommitStateDB.WithContext(ctx).GetLogs(hash)
}

// AllLogs calls CommitStateDB.AllLogs using the passed in context
func (k *Keeper) AllLogs(ctx sdk.Context) []*ethtypes.Log {
	return k.CommitStateDB.WithContext(ctx).AllLogs()
}

// GetRefund calls CommitStateDB.GetRefund using the passed in context
func (k *Keeper) GetRefund(ctx sdk.Context) uint64 {
	return k.CommitStateDB.WithContext(ctx).GetRefund()
}

// Preimages calls CommitStateDB.Preimages using the passed in context
func (k *Keeper) Preimages(ctx sdk.Context) map[ethcmn.Hash][]byte {
	return k.CommitStateDB.WithContext(ctx).Preimages()
}

// HasSuicided calls CommitStateDB.HasSuicided using the passed in context
func (k *Keeper) HasSuicided(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.CommitStateDB.WithContext(ctx).HasSuicided(addr)
}

// StorageTrie calls CommitStateDB.StorageTrie using the passed in context
func (k *Keeper) StorageTrie(ctx sdk.Context, addr ethcmn.Address) ethstate.Trie {
	return k.CommitStateDB.WithContext(ctx).StorageTrie(addr)
}

// ----------------------------------------------------------------------------
// Persistence
// ----------------------------------------------------------------------------

// Commit calls CommitStateDB.Commit using the passed in context
func (k *Keeper) Commit(ctx sdk.Context, deleteEmptyObjects bool) (root ethcmn.Hash, err error) {
	return k.CommitStateDB.WithContext(ctx).Commit(deleteEmptyObjects)
}

// Finalise calls CommitStateDB.Finalise using the passed in context
func (k *Keeper) Finalise(ctx sdk.Context, deleteEmptyObjects bool) error {
	return k.CommitStateDB.WithContext(ctx).Finalise(deleteEmptyObjects)
}

// IntermediateRoot calls CommitStateDB.IntermediateRoot using the passed in context
func (k *Keeper) IntermediateRoot(ctx sdk.Context, deleteEmptyObjects bool) error {
	_, err := k.CommitStateDB.WithContext(ctx).IntermediateRoot(deleteEmptyObjects)
	return err
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot(ctx sdk.Context) int {
	return k.CommitStateDB.WithContext(ctx).Snapshot()
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(ctx sdk.Context, revID int) {
	k.CommitStateDB.WithContext(ctx).RevertToSnapshot(revID)
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// Database calls CommitStateDB.Database using the passed in context
func (k *Keeper) Database(ctx sdk.Context) ethstate.Database {
	return k.CommitStateDB.WithContext(ctx).Database()
}

// Empty calls CommitStateDB.Empty using the passed in context
func (k *Keeper) Empty(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.CommitStateDB.WithContext(ctx).Empty(addr)
}

// Exist calls CommitStateDB.Exist using the passed in context
func (k *Keeper) Exist(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.CommitStateDB.WithContext(ctx).Exist(addr)
}

// Error calls CommitStateDB.Error using the passed in context
func (k *Keeper) Error(ctx sdk.Context) error {
	return k.CommitStateDB.WithContext(ctx).Error()
}

// Suicide calls CommitStateDB.Suicide using the passed in context
func (k *Keeper) Suicide(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.CommitStateDB.WithContext(ctx).Suicide(addr)
}

// Reset calls CommitStateDB.Reset using the passed in context
func (k *Keeper) Reset(ctx sdk.Context, root ethcmn.Hash) error {
	return k.CommitStateDB.WithContext(ctx).Reset(root)
}

// Prepare calls CommitStateDB.Prepare using the passed in context
func (k *Keeper) Prepare(ctx sdk.Context, thash, bhash ethcmn.Hash, txi int) {
	k.CommitStateDB.WithContext(ctx).Prepare(thash, bhash, txi)
}

// CreateAccount calls CommitStateDB.CreateAccount using the passed in context
func (k *Keeper) CreateAccount(ctx sdk.Context, addr ethcmn.Address) {
	k.CommitStateDB.WithContext(ctx).CreateAccount(addr)
}

// UpdateAccounts calls CommitStateDB.UpdateAccounts using the passed in context
func (k *Keeper) UpdateAccounts(ctx sdk.Context) {
	k.CommitStateDB.WithContext(ctx).UpdateAccounts()
}

// ClearStateObjects calls CommitStateDB.ClearStateObjects using the passed in context
func (k *Keeper) ClearStateObjects(ctx sdk.Context) {
	k.CommitStateDB.WithContext(ctx).ClearStateObjects()
}

// Copy calls CommitStateDB.Copy using the passed in context
func (k *Keeper) Copy(ctx sdk.Context) ethvm.StateDB {
	return k.CommitStateDB.WithContext(ctx).Copy()
}

// ForEachStorage calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(ctx sdk.Context, addr ethcmn.Address, cb func(key, value ethcmn.Hash) bool) error {
	return k.CommitStateDB.WithContext(ctx).ForEachStorage(addr, cb)
}

// GetOrNewStateObject calls CommitStateDB.GetOrNetStateObject using the passed in context
func (k *Keeper) GetOrNewStateObject(ctx sdk.Context, addr ethcmn.Address) types.StateObject {
	return k.CommitStateDB.WithContext(ctx).GetOrNewStateObject(addr)
}

package evm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	types "github.com/cosmos/ethermint/x/evm/types"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"math/big"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface
type Keeper struct {
	csdb *types.CommitStateDB
	cdc  *codec.Codec
}

func NewKeeper(ak auth.AccountKeeper, storageKey, codeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		csdb: types.NewCommitStateDB(sdk.Context{}, ak, storageKey, codeKey),
		cdc:  cdc,
	}
}

// ----------------------------------------------------------------------------
// Genesis
// ----------------------------------------------------------------------------

// CreateGenesisAccount initializes an account and its balance, code, and storage
func (k *Keeper) CreateGenesisAccount(ctx sdk.Context, account GenesisAccount) {
	csdb := k.csdb.WithContext(ctx)
	csdb.SetBalance(account.Address, account.Balance)
	csdb.SetCode(account.Address, account.Code)
	for _, key := range account.Storage {
		csdb.SetState(account.Address, key, account.Storage[key])
	}

}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// Calls CommitStateDB.SetBalance using the passed in context
func (k *Keeper) SetBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).SetBalance(addr, amount)
}

// Calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).AddBalance(addr, amount)
}

// Calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).SubBalance(addr, amount)
}

// Calls CommitStateDB.SetNonce using the passed in context
func (k *Keeper) SetNonce(ctx sdk.Context, addr ethcmn.Address, nonce uint64) {
	k.csdb.WithContext(ctx).SetNonce(addr, nonce)
}

// Calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(ctx sdk.Context, addr ethcmn.Address, key, value ethcmn.Hash) {
	k.csdb.WithContext(ctx).SetState(addr, key, value)
}

// Calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(ctx sdk.Context, addr ethcmn.Address, code []byte) {
	k.csdb.WithContext(ctx).SetCode(addr, code)
}

// Calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(ctx sdk.Context, log *ethtypes.Log) {
	k.csdb.WithContext(ctx).AddLog(log)
}

// Calls CommitStateDB.AddPreimage using the passed in context
func (k *Keeper) AddPreimage(ctx sdk.Context, hash ethcmn.Hash, preimage []byte) {
	k.csdb.WithContext(ctx).AddPreimage(hash, preimage)
}

// Calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(ctx sdk.Context, gas uint64) {
	k.csdb.WithContext(ctx).AddRefund(gas)
}

// Calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(ctx sdk.Context, gas uint64) {
	k.csdb.WithContext(ctx).SubRefund(gas)
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// Calls CommitStateDB.GetBalance using the passed in context
func (k *Keeper) GetBalance(ctx sdk.Context, addr ethcmn.Address) *big.Int {
	return k.csdb.WithContext(ctx).GetBalance(addr)
}

// Calls CommitStateDB.GetNonce using the passed in context
func (k *Keeper) GetNonce(ctx sdk.Context, addr ethcmn.Address) uint64 {
	return k.csdb.WithContext(ctx).GetNonce(addr)
}

// Calls CommitStateDB.TxIndex using the passed in context
func (k *Keeper) TxIndex(ctx sdk.Context) int {
	return k.csdb.WithContext(ctx).TxIndex()
}

// Calls CommitStateDB.BlockHash using the passed in context
func (k *Keeper) BlockHash(ctx sdk.Context) ethcmn.Hash {
	return k.csdb.WithContext(ctx).BlockHash()
}

// Calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(ctx sdk.Context, addr ethcmn.Address) []byte {
	return k.csdb.WithContext(ctx).GetCode(addr)
}

// Calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(ctx sdk.Context, addr ethcmn.Address) int {
	return k.csdb.WithContext(ctx).GetCodeSize(addr)
}

// Calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(ctx sdk.Context, addr ethcmn.Address) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetCodeHash(addr)
}

// Calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetState(addr, hash)
}

// Calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetCommittedState(addr, hash)
}

// Calls CommitStateDB.GetLogs using the passed in context
func (k *Keeper) GetLogs(ctx sdk.Context, hash ethcmn.Hash) []*ethtypes.Log {
	return k.csdb.WithContext(ctx).GetLogs(hash)
}

// Calls CommitStateDB.Logs using the passed in context
func (k *Keeper) Logs(ctx sdk.Context) []*ethtypes.Log {
	return k.csdb.WithContext(ctx).Logs()
}

// Calls CommitStateDB.GetRefund using the passed in context
func (k *Keeper) GetRefund(ctx sdk.Context) uint64 {
	return k.csdb.WithContext(ctx).GetRefund()
}

// Calls CommitStateDB.Preimages using the passed in context
func (k *Keeper) Preimages(ctx sdk.Context) map[ethcmn.Hash][]byte {
	return k.csdb.WithContext(ctx).Preimages()
}

// Calls CommitStateDB.HasSuicided using the passed in context
func (k *Keeper) HasSuicided(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).HasSuicided(addr)
}

// Calls CommitStateDB.StorageTrie using the passed in context
func (k *Keeper) StorageTrie(ctx sdk.Context, addr ethcmn.Address) ethstate.Trie {
	return k.csdb.WithContext(ctx).StorageTrie(addr)
}

// ----------------------------------------------------------------------------
// Persistence
// ----------------------------------------------------------------------------

// Calls CommitStateDB.Commit using the passed in context
func (k *Keeper) Commit(ctx sdk.Context, deleteEmptyObjects bool) (root ethcmn.Hash, err error) {
	return k.csdb.WithContext(ctx).Commit(deleteEmptyObjects)
}

// Calls CommitStateDB.Finalise using the passed in context
func (k *Keeper) Finalise(ctx sdk.Context, deleteEmptyObjects bool) {
	k.csdb.WithContext(ctx).Finalise(deleteEmptyObjects)
}

// Calls CommitStateDB.IntermediateRoot using the passed in context
func (k *Keeper) IntermediateRoot(ctx sdk.Context, deleteEmptyObjects bool) {
	k.csdb.WithContext(ctx).IntermediateRoot(deleteEmptyObjects)
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot(ctx sdk.Context) int {
	return k.csdb.WithContext(ctx).Snapshot()
}

// Calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(ctx sdk.Context, revID int) {
	k.csdb.WithContext(ctx).RevertToSnapshot(revID)
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// Calls CommitStateDB.Database using the passed in context
func (k *Keeper) Database(ctx sdk.Context) ethstate.Database {
	return k.csdb.WithContext(ctx).Database()
}

// Calls CommitStateDB.Empty using the passed in context
func (k *Keeper) Empty(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Empty(addr)
}

// Calls CommitStateDB.Exist using the passed in context
func (k *Keeper) Exist(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Exist(addr)
}

// Calls CommitStateDB.Error using the passed in context
func (k *Keeper) Error(ctx sdk.Context) error {
	return k.csdb.WithContext(ctx).Error()
}

// Calls CommitStateDB.Suicide using the passed in context
func (k *Keeper) Suicide(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Suicide(addr)
}

// Calls CommitStateDB.Reset using the passed in context
func (k *Keeper) Reset(ctx sdk.Context, root ethcmn.Hash) error {
	return k.csdb.WithContext(ctx).Reset(root)
}

// Calls CommitStateDB.Prepare using the passed in context
func (k *Keeper) Prepare(ctx sdk.Context, thash, bhash ethcmn.Hash, txi int) {
	k.csdb.WithContext(ctx).Prepare(thash, bhash, txi)
}

// Calls CommitStateDB.CreateAccount using the passed in context
func (k *Keeper) CreateAccount(ctx sdk.Context, addr ethcmn.Address) {
	k.csdb.WithContext(ctx).CreateAccount(addr)
}

// Calls CommitStateDB.Copy using the passed in context
func (k *Keeper) Copy(ctx sdk.Context) ethvm.StateDB {
	return k.csdb.WithContext(ctx).Copy()
}

// Calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(ctx sdk.Context, addr ethcmn.Address, cb func(key, value ethcmn.Hash) bool) error {
	return k.csdb.WithContext(ctx).ForEachStorage(addr, cb)
}

// Calls CommitStateDB.GetOrNetStateObject using the passed in context
func (k *Keeper) GetOrNewStateObject(ctx sdk.Context, addr ethcmn.Address) types.StateObject {
	return k.csdb.WithContext(ctx).GetOrNewStateObject(addr)
}

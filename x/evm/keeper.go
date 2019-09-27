package evm

import (
	"bytes"
	"fmt"

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
	csdb     *types.CommitStateDB
	cdc      *codec.Codec
	blockKey sdk.StoreKey
}

// NewKeeper generates new evm module keeper
func NewKeeper(ak auth.AccountKeeper, storageKey, codeKey sdk.StoreKey,
	blockKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		csdb:     types.NewCommitStateDB(sdk.Context{}, ak, storageKey, codeKey),
		cdc:      cdc,
		blockKey: blockKey,
	}
}

// ----------------------------------------------------------------------------
// Block hash mapping functions
// May be removed when using only as module (only required by rpc api)
// ----------------------------------------------------------------------------

// SetBlockHashMapping sets the mapping from block consensus hash to block height
func (k *Keeper) SetBlockHashMapping(ctx sdk.Context, hash []byte, height int64) {
	store := ctx.KVStore(k.blockKey)
	if !bytes.Equal(hash, []byte{}) {
		store.Set(hash, k.cdc.MustMarshalBinaryLengthPrefixed(height))
	}
}

// GetBlockHashMapping gets block height from block consensus hash
func (k *Keeper) GetBlockHashMapping(ctx sdk.Context, hash []byte) (height int64) {
	store := ctx.KVStore(k.blockKey)
	bz := store.Get(hash)
	if bytes.Equal(bz, []byte{}) {
		panic(fmt.Errorf("block with hash %s not found", ethcmn.Bytes2Hex(hash)))
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &height)
	return
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

// SetBalance calls CommitStateDB.SetBalance using the passed in context
func (k *Keeper) SetBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).SetBalance(addr, amount)
}

// AddBalance calls CommitStateDB.AddBalance using the passed in context
func (k *Keeper) AddBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).AddBalance(addr, amount)
}

// SubBalance calls CommitStateDB.SubBalance using the passed in context
func (k *Keeper) SubBalance(ctx sdk.Context, addr ethcmn.Address, amount *big.Int) {
	k.csdb.WithContext(ctx).SubBalance(addr, amount)
}

// SetNonce calls CommitStateDB.SetNonce using the passed in context
func (k *Keeper) SetNonce(ctx sdk.Context, addr ethcmn.Address, nonce uint64) {
	k.csdb.WithContext(ctx).SetNonce(addr, nonce)
}

// SetState calls CommitStateDB.SetState using the passed in context
func (k *Keeper) SetState(ctx sdk.Context, addr ethcmn.Address, key, value ethcmn.Hash) {
	k.csdb.WithContext(ctx).SetState(addr, key, value)
}

// SetCode calls CommitStateDB.SetCode using the passed in context
func (k *Keeper) SetCode(ctx sdk.Context, addr ethcmn.Address, code []byte) {
	k.csdb.WithContext(ctx).SetCode(addr, code)
}

// AddLog calls CommitStateDB.AddLog using the passed in context
func (k *Keeper) AddLog(ctx sdk.Context, log *ethtypes.Log) {
	k.csdb.WithContext(ctx).AddLog(log)
}

// AddPreimage calls CommitStateDB.AddPreimage using the passed in context
func (k *Keeper) AddPreimage(ctx sdk.Context, hash ethcmn.Hash, preimage []byte) {
	k.csdb.WithContext(ctx).AddPreimage(hash, preimage)
}

// AddRefund calls CommitStateDB.AddRefund using the passed in context
func (k *Keeper) AddRefund(ctx sdk.Context, gas uint64) {
	k.csdb.WithContext(ctx).AddRefund(gas)
}

// SubRefund calls CommitStateDB.SubRefund using the passed in context
func (k *Keeper) SubRefund(ctx sdk.Context, gas uint64) {
	k.csdb.WithContext(ctx).SubRefund(gas)
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// GetBalance calls CommitStateDB.GetBalance using the passed in context
func (k *Keeper) GetBalance(ctx sdk.Context, addr ethcmn.Address) *big.Int {
	return k.csdb.WithContext(ctx).GetBalance(addr)
}

// GetNonce calls CommitStateDB.GetNonce using the passed in context
func (k *Keeper) GetNonce(ctx sdk.Context, addr ethcmn.Address) uint64 {
	return k.csdb.WithContext(ctx).GetNonce(addr)
}

// TxIndex calls CommitStateDB.TxIndex using the passed in context
func (k *Keeper) TxIndex(ctx sdk.Context) int {
	return k.csdb.WithContext(ctx).TxIndex()
}

// BlockHash calls CommitStateDB.BlockHash using the passed in context
func (k *Keeper) BlockHash(ctx sdk.Context) ethcmn.Hash {
	return k.csdb.WithContext(ctx).BlockHash()
}

// GetCode calls CommitStateDB.GetCode using the passed in context
func (k *Keeper) GetCode(ctx sdk.Context, addr ethcmn.Address) []byte {
	return k.csdb.WithContext(ctx).GetCode(addr)
}

// GetCodeSize calls CommitStateDB.GetCodeSize using the passed in context
func (k *Keeper) GetCodeSize(ctx sdk.Context, addr ethcmn.Address) int {
	return k.csdb.WithContext(ctx).GetCodeSize(addr)
}

// GetCodeHash calls CommitStateDB.GetCodeHash using the passed in context
func (k *Keeper) GetCodeHash(ctx sdk.Context, addr ethcmn.Address) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetCodeHash(addr)
}

// GetState calls CommitStateDB.GetState using the passed in context
func (k *Keeper) GetState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetState(addr, hash)
}

// GetCommittedState calls CommitStateDB.GetCommittedState using the passed in context
func (k *Keeper) GetCommittedState(ctx sdk.Context, addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	return k.csdb.WithContext(ctx).GetCommittedState(addr, hash)
}

// GetLogs calls CommitStateDB.GetLogs using the passed in context
func (k *Keeper) GetLogs(ctx sdk.Context, hash ethcmn.Hash) []*ethtypes.Log {
	return k.csdb.WithContext(ctx).GetLogs(hash)
}

// Logs calls CommitStateDB.Logs using the passed in context
func (k *Keeper) Logs(ctx sdk.Context) []*ethtypes.Log {
	return k.csdb.WithContext(ctx).Logs()
}

// GetRefund calls CommitStateDB.GetRefund using the passed in context
func (k *Keeper) GetRefund(ctx sdk.Context) uint64 {
	return k.csdb.WithContext(ctx).GetRefund()
}

// Preimages calls CommitStateDB.Preimages using the passed in context
func (k *Keeper) Preimages(ctx sdk.Context) map[ethcmn.Hash][]byte {
	return k.csdb.WithContext(ctx).Preimages()
}

// HasSuicided calls CommitStateDB.HasSuicided using the passed in context
func (k *Keeper) HasSuicided(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).HasSuicided(addr)
}

// StorageTrie calls CommitStateDB.StorageTrie using the passed in context
func (k *Keeper) StorageTrie(ctx sdk.Context, addr ethcmn.Address) ethstate.Trie {
	return k.csdb.WithContext(ctx).StorageTrie(addr)
}

// ----------------------------------------------------------------------------
// Persistence
// ----------------------------------------------------------------------------

// Commit calls CommitStateDB.Commit using the passed { in context
func (k *Keeper) Commit(ctx sdk.Context, deleteEmptyObjects bool) (root ethcmn.Hash, err error) {
	return k.csdb.WithContext(ctx).Commit(deleteEmptyObjects)
}

// Finalise calls CommitStateDB.Finalise using the passed in context
func (k *Keeper) Finalise(ctx sdk.Context, deleteEmptyObjects bool) {
	k.csdb.WithContext(ctx).Finalise(deleteEmptyObjects)
}

// IntermediateRoot calls CommitStateDB.IntermediateRoot using the passed in context
func (k *Keeper) IntermediateRoot(ctx sdk.Context, deleteEmptyObjects bool) {
	k.csdb.WithContext(ctx).IntermediateRoot(deleteEmptyObjects)
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot calls CommitStateDB.Snapshot using the passed in context
func (k *Keeper) Snapshot(ctx sdk.Context) int {
	return k.csdb.WithContext(ctx).Snapshot()
}

// RevertToSnapshot calls CommitStateDB.RevertToSnapshot using the passed in context
func (k *Keeper) RevertToSnapshot(ctx sdk.Context, revID int) {
	k.csdb.WithContext(ctx).RevertToSnapshot(revID)
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// Database calls CommitStateDB.Database using the passed in context
func (k *Keeper) Database(ctx sdk.Context) ethstate.Database {
	return k.csdb.WithContext(ctx).Database()
}

// Empty calls CommitStateDB.Empty using the passed in context
func (k *Keeper) Empty(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Empty(addr)
}

// Exist calls CommitStateDB.Exist using the passed in context
func (k *Keeper) Exist(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Exist(addr)
}

// Error calls CommitStateDB.Error using the passed in context
func (k *Keeper) Error(ctx sdk.Context) error {
	return k.csdb.WithContext(ctx).Error()
}

// Suicide calls CommitStateDB.Suicide using the passed in context
func (k *Keeper) Suicide(ctx sdk.Context, addr ethcmn.Address) bool {
	return k.csdb.WithContext(ctx).Suicide(addr)
}

// Reset calls CommitStateDB.Reset using the passed in context
func (k *Keeper) Reset(ctx sdk.Context, root ethcmn.Hash) error {
	return k.csdb.WithContext(ctx).Reset(root)
}

// Prepare calls CommitStateDB.Prepare using the passed in context
func (k *Keeper) Prepare(ctx sdk.Context, thash, bhash ethcmn.Hash, txi int) {
	k.csdb.WithContext(ctx).Prepare(thash, bhash, txi)
}

// CreateAccount calls CommitStateDB.CreateAccount using the passed in context
func (k *Keeper) CreateAccount(ctx sdk.Context, addr ethcmn.Address) {
	k.csdb.WithContext(ctx).CreateAccount(addr)
}

// Copy calls CommitStateDB.Copy using the passed in context
func (k *Keeper) Copy(ctx sdk.Context) ethvm.StateDB {
	return k.csdb.WithContext(ctx).Copy()
}

// ForEachStorage calls CommitStateDB.ForEachStorage using passed in context
func (k *Keeper) ForEachStorage(ctx sdk.Context, addr ethcmn.Address, cb func(key, value ethcmn.Hash) bool) error {
	return k.csdb.WithContext(ctx).ForEachStorage(addr, cb)
}

// GetOrNewStateObject calls CommitStateDB.GetOrNetStateObject using the passed in context
func (k *Keeper) GetOrNewStateObject(ctx sdk.Context, addr ethcmn.Address) types.StateObject {
	return k.csdb.WithContext(ctx).GetOrNewStateObject(addr)
}

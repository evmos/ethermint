package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethvm "github.com/ethereum/go-ethereum/core/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/x/evm/types"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"math/big"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface.
type Keeper struct {
	// Amino codec
	cdc *codec.Codec
	// Store key required to update the block bloom filter mappings needed for the
	// Web3 API
	blockKey      sdk.StoreKey
	CommitStateDB *types.CommitStateDB
	TxCount       int
	Bloom         *big.Int
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc *codec.Codec, blockKey, codeKey, storeKey sdk.StoreKey,
	ak types.AccountKeeper, bk types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		blockKey:      blockKey,
		CommitStateDB: types.NewCommitStateDB(sdk.Context{}, codeKey, storeKey, ak, bk),
		TxCount:       0,
		Bloom:         big.NewInt(0),
	}
}

// ----------------------------------------------------------------------------
// Block hash mapping functions
// May be removed when using only as module (only required by rpc api)
// ----------------------------------------------------------------------------

// GetBlockHashMapping gets block height from block consensus hash
func (k Keeper) GetBlockHashMapping(ctx sdk.Context, hash []byte) (int64, error) {
	store := ctx.KVStore(k.blockKey)
	bz := store.Get(hash)
	if len(bz) == 0 {
		return 0, fmt.Errorf("block with hash '%s' not found", ethcmn.BytesToHash(hash))
	}

	height := binary.BigEndian.Uint64(bz)
	return int64(height), nil
}

// SetBlockHashMapping sets the mapping from block consensus hash to block height
func (k Keeper) SetBlockHashMapping(ctx sdk.Context, hash []byte, height int64) {
	store := ctx.KVStore(k.blockKey)
	bz := sdk.Uint64ToBigEndian(uint64(height))
	store.Set(hash, bz)
}

// ----------------------------------------------------------------------------
// Block bloom bits mapping functions
// May be removed when using only as module (only required by rpc api)
// ----------------------------------------------------------------------------

// GetBlockBloomMapping gets bloombits from block height
func (k Keeper) GetBlockBloomMapping(ctx sdk.Context, height int64) (ethtypes.Bloom, error) {
	store := ctx.KVStore(k.blockKey)
	heightBz := sdk.Uint64ToBigEndian(uint64(height))
	bz := store.Get(types.BloomKey(heightBz))
	if len(bz) == 0 {
		return ethtypes.Bloom{}, fmt.Errorf("block at height %d not found", height)
	}

	return ethtypes.BytesToBloom(bz), nil
}

// SetBlockBloomMapping sets the mapping from block height to bloom bits
func (k Keeper) SetBlockBloomMapping(ctx sdk.Context, bloom ethtypes.Bloom, height int64) {
	store := ctx.KVStore(k.blockKey)
	heightBz := sdk.Uint64ToBigEndian(uint64(height))
	store.Set(types.BloomKey(heightBz), bloom.Bytes())
}

// SetTransactionLogs sets the transaction's logs in the KVStore
func (k *Keeper) SetTransactionLogs(ctx sdk.Context, logs []*ethtypes.Log, hash []byte) error {
	store := ctx.KVStore(k.blockKey)
	encLogs, err := types.EncodeLogs(logs)
	if err != nil {
		return err
	}

	store.Set(types.LogsKey(hash), encLogs)
	return nil
}

// GetTransactionLogs gets the logs for a transaction from the KVStore
func (k *Keeper) GetTransactionLogs(ctx sdk.Context, hash []byte) ([]*ethtypes.Log, error) {
	store := ctx.KVStore(k.blockKey)
	encLogs := store.Get(types.LogsKey(hash))
	if len(encLogs) == 0 {
		return nil, errors.New("cannot get transaction logs")
	}

	return types.DecodeLogs(encLogs)
}

// ----------------------------------------------------------------------------
// Genesis
// ----------------------------------------------------------------------------

// CreateGenesisAccount initializes an account and its balance, code, and storage
func (k *Keeper) CreateGenesisAccount(ctx sdk.Context, account types.GenesisAccount) {
	csdb := k.CommitStateDB.WithContext(ctx)
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

// Commit calls CommitStateDB.Commit using the passed { in context
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

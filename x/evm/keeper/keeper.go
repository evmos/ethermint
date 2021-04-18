package keeper

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/ethermint/metrics"
	"github.com/cosmos/ethermint/x/evm/types"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryMarshaler
	// Store key required for the EVM Prefix KVStore. It is required by:
	// - storing Account's Storage State
	// - storing Account's Code
	// - storing transaction Logs
	// - storing block height -> bloom filter map. Needed for the Web3 API.
	// - storing block hash -> block height map. Needed for the Web3 API. TODO: remove
	storeKey sdk.StoreKey

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper

	// Ethermint concrete implementation on the EVM StateDB interface
	CommitStateDB *types.CommitStateDB
	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	TxCount int
	Bloom   *big.Int

	// LogsCache keeps mapping of contract address -> eth logs emitted
	// during EVM execution in the current block.
	LogsCache map[common.Address][]*ethtypes.Log

	svcTags metrics.Tags
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc codec.BinaryMarshaler, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bankKeeper types.BankKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	// NOTE: we pass in the parameter space to the CommitStateDB in order to use custom denominations for the EVM operations
	return &Keeper{
		svcTags: metrics.Tags{
			"svc": "evm_k",
		},

		cdc:           cdc,
		accountKeeper: ak,
		bankKeeper:    bankKeeper,
		storeKey:      storeKey,
		CommitStateDB: types.NewCommitStateDB(sdk.Context{}, storeKey, paramSpace, ak, bankKeeper),
		TxCount:       0,
		Bloom:         big.NewInt(0),
		LogsCache:     map[common.Address][]*ethtypes.Log{},
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// ----------------------------------------------------------------------------
// Block bloom bits mapping functions
// Required by Web3 API.
// ----------------------------------------------------------------------------

// GetBlockBloom gets bloombits from block height
func (k Keeper) GetBlockBloom(ctx sdk.Context, height int64) (ethtypes.Bloom, bool) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)

	key := types.BloomKey(height)
	has := store.Has(key)
	if !has {
		return ethtypes.Bloom{}, true // sometimes bloom not found, fix this
	}

	bz := store.Get(key)
	return ethtypes.BytesToBloom(bz), true
}

// SetBlockBloom sets the mapping from block height to bloom bits
func (k Keeper) SetBlockBloom(ctx sdk.Context, height int64, bloom ethtypes.Bloom) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)

	key := types.BloomKey(height)
	store.Set(key, bloom.Bytes())
}

// GetBlockHash gets block height from block consensus hash
func (k Keeper) GetBlockHashFromHeight(ctx sdk.Context, height int64) ([]byte, bool) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyBlockHeightHash(uint64(height)))
	if len(bz) == 0 {
		return common.Hash{}.Bytes(), false
	}

	return common.BytesToHash(bz).Bytes(), true
}

// SetBlockHash sets the mapping from block consensus hash to block height
func (k Keeper) SetBlockHash(ctx sdk.Context, hash []byte, height int64) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(uint64(height))
	store.Set(types.KeyBlockHash(common.BytesToHash(hash)), bz)
}

// GetBlockHash gets block height from block consensus hash
func (k Keeper) GetBlockHeightByHash(ctx sdk.Context, hash common.Hash) (int64, bool) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyBlockHash(hash))
	if len(bz) == 0 {
		return 0, false
	}

	height := sdk.BigEndianToUint64(bz)
	return int64(height), true
}

// SetBlockHash sets the mapping from block consensus hash to block height
func (k Keeper) SetBlockHeightToHash(ctx sdk.Context, hash []byte, height int64) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyBlockHeightHash(uint64(height)), hash)
}

// SetTxReceiptToHash sets the mapping from tx hash to tx receipt
func (k Keeper) SetTxReceiptToHash(ctx sdk.Context, hash common.Hash, receipt *types.TxReceipt) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	data := k.cdc.MustMarshalBinaryBare(receipt)

	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyHashTxReceipt(hash), data)
}

// GetHeightHash returns the block header hash associated with a given block height and chain epoch number.
func (k Keeper) GetHeightHash(ctx sdk.Context, height uint64) common.Hash {
	return k.CommitStateDB.WithContext(ctx).GetHeightHash(height)
}

// SetHeightHash sets the block header hash associated with a given height.
func (k Keeper) SetHeightHash(ctx sdk.Context, height uint64, hash common.Hash) {
	k.CommitStateDB.WithContext(ctx).SetHeightHash(height, hash)
}

// GetTxReceiptFromHash gets tx receipt by tx hash.
func (k Keeper) GetTxReceiptFromHash(ctx sdk.Context, hash common.Hash) (*types.TxReceipt, bool) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	data := store.Get(types.KeyHashTxReceipt(hash))
	if data == nil || len(data) == 0 {
		return nil, false
	}

	var receipt types.TxReceipt
	k.cdc.MustUnmarshalBinaryBare(data, &receipt)

	return &receipt, true
}

// AddTxHashToBlock stores tx hash in the list of tx for the block.
func (k Keeper) AddTxHashToBlock(ctx sdk.Context, blockHeight int64, txHash common.Hash) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	key := types.KeyBlockHeightTxs(uint64(blockHeight))

	list := types.BytesList{}

	store := ctx.KVStore(k.storeKey)
	data := store.Get(key)
	if len(data) > 0 {
		k.cdc.MustUnmarshalBinaryBare(data, &list)
	}

	list.Bytes = append(list.Bytes, txHash.Bytes())

	data = k.cdc.MustMarshalBinaryBare(&list)
	store.Set(key, data)
}

// GetTxsFromBlock returns list of tx hash in the block by height.
func (k Keeper) GetTxsFromBlock(ctx sdk.Context, blockHeight int64) []common.Hash {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	key := types.KeyBlockHeightTxs(uint64(blockHeight))

	store := ctx.KVStore(k.storeKey)
	data := store.Get(key)
	if len(data) > 0 {
		list := types.BytesList{}
		k.cdc.MustUnmarshalBinaryBare(data, &list)

		txs := make([]common.Hash, 0, len(list.Bytes))
		for _, b := range list.Bytes {
			txs = append(txs, common.BytesToHash(b))
		}

		return txs
	}

	return nil
}

// GetTxReceiptsByBlockHeight gets tx receipts by block height.
func (k Keeper) GetTxReceiptsByBlockHeight(ctx sdk.Context, blockHeight int64) []*types.TxReceipt {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	txs := k.GetTxsFromBlock(ctx, blockHeight)
	if len(txs) == 0 {
		return nil
	}

	store := ctx.KVStore(k.storeKey)

	receipts := make([]*types.TxReceipt, 0, len(txs))

	for idx, txHash := range txs {
		data := store.Get(types.KeyHashTxReceipt(txHash))
		if data == nil || len(data) == 0 {
			continue
		}

		var receipt types.TxReceipt
		k.cdc.MustUnmarshalBinaryBare(data, &receipt)
		receipt.Index = uint64(idx)
		receipts = append(receipts, &receipt)
	}

	return receipts
}

// GetTxReceiptsByBlockHash gets tx receipts by block hash.
func (k Keeper) GetTxReceiptsByBlockHash(ctx sdk.Context, hash common.Hash) []*types.TxReceipt {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	blockHeight, ok := k.GetBlockHeightByHash(ctx, hash)
	if !ok {
		return nil
	}

	return k.GetTxReceiptsByBlockHeight(ctx, blockHeight)
}

// GetAllTxLogs return all the transaction logs from the store.
func (k Keeper) GetAllTxLogs(ctx sdk.Context) []types.TransactionLogs {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.KeyPrefixLogs)
	defer iterator.Close()

	txsLogs := []types.TransactionLogs{}
	for ; iterator.Valid(); iterator.Next() {
		var txLog types.TransactionLogs
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &txLog)

		// add a new entry
		txsLogs = append(txsLogs, txLog)
	}
	return txsLogs
}

// GetAccountStorage return state storage associated with an account
func (k Keeper) GetAccountStorage(ctx sdk.Context, address common.Address) (types.Storage, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	storage := types.Storage{}

	err := k.ForEachStorage(ctx, address, func(key, value common.Hash) bool {
		storage = append(storage, types.NewState(key, value))
		return false
	})
	if err != nil {
		return types.Storage{}, err
	}

	return storage, nil
}

// GetChainConfig gets block height from block consensus hash
func (k Keeper) GetChainConfig(ctx sdk.Context) (types.ChainConfig, bool) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixChainConfig)
	if len(bz) == 0 {
		return types.ChainConfig{}, false
	}

	var config types.ChainConfig
	k.cdc.MustUnmarshalBinaryBare(bz, &config)
	return config, true
}

// SetChainConfig sets the mapping from block consensus hash to block height
func (k Keeper) SetChainConfig(ctx sdk.Context, config types.ChainConfig) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&config)
	store.Set(types.KeyPrefixChainConfig, bz)
}

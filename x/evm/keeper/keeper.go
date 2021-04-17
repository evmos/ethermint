package keeper

import (
	"fmt"
	"math/big"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
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
	// - storing block hash -> block height map. Needed for the Web3 API.
	storeKey sdk.StoreKey
	// Account Keeper for fetching accounts
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	// Ethermint concrete implementation on the EVM StateDB interface
	CommitStateDB *types.CommitStateDB
	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	TxCount int
	Bloom   *big.Int
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
		cdc:           cdc,
		accountKeeper: ak,
		bankKeeper:    bankKeeper,
		storeKey:      storeKey,
		CommitStateDB: types.NewCommitStateDB(sdk.Context{}, storeKey, paramSpace, ak, bankKeeper),
		TxCount:       0,
		Bloom:         big.NewInt(0),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ----------------------------------------------------------------------------
// Epoch Height -> hash mapping functions
// Required by EVM context's GetHashFunc
// ----------------------------------------------------------------------------

// GetHeightHash returns the block header hash associated with a given block height and chain epoch number.
func (k Keeper) GetHeightHash(ctx sdk.Context, height uint64) common.Hash {
	return k.CommitStateDB.WithContext(ctx).GetHeightHash(height)
}

// SetHeightHash sets the block header hash associated with a given height.
func (k Keeper) SetHeightHash(ctx sdk.Context, height uint64, hash common.Hash) {
	k.CommitStateDB.WithContext(ctx).SetHeightHash(height, hash)
}

// ----------------------------------------------------------------------------
// Block bloom bits mapping functions
// Required by Web3 API.
// ----------------------------------------------------------------------------

// GetBlockBloom gets bloombits from block height
func (k Keeper) GetBlockBloom(ctx sdk.Context, height int64) (ethtypes.Bloom, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBloom)
	has := store.Has(types.BloomKey(height))
	if !has {
		return ethtypes.Bloom{}, false
	}

	bz := store.Get(types.BloomKey(height))
	return ethtypes.BytesToBloom(bz), true
}

// SetBlockBloom sets the mapping from block height to bloom bits
func (k Keeper) SetBlockBloom(ctx sdk.Context, height int64, bloom ethtypes.Bloom) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBloom)
	store.Set(types.BloomKey(height), bloom.Bytes())
}

// GetAllTxLogs return all the transaction logs from the store.
func (k Keeper) GetAllTxLogs(ctx sdk.Context) []types.TransactionLogs {
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
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&config)
	store.Set(types.KeyPrefixChainConfig, bz)
}

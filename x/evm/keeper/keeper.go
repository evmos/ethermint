package keeper

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Keeper wraps the CommitStateDB, allowing us to pass in SDK context while adhering
// to the StateDB interface.
type Keeper struct {
	// Amino codec
	cdc *codec.Codec
	// Store key required for the EVM Prefix KVStore. It is required by:
	// - storing Account's Storage State
	// - storing Account's Code
	// - storing transaction Logs
	// - storing block height -> bloom filter map. Needed for the Web3 API.
	// - storing block hash -> block height map. Needed for the Web3 API.
	storeKey      sdk.StoreKey
	CommitStateDB *types.CommitStateDB
	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	TxCount int
	Bloom   *big.Int
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc *codec.Codec, storeKey sdk.StoreKey, ak types.AccountKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		CommitStateDB: types.NewCommitStateDB(sdk.Context{}, storeKey, ak),
		TxCount:       0,
		Bloom:         big.NewInt(0),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ----------------------------------------------------------------------------
// Block hash mapping functions
// Required by Web3 API.
//  TODO: remove once tendermint support block queries by hash.
// ----------------------------------------------------------------------------

// GetBlockHash gets block height from block consensus hash
func (k Keeper) GetBlockHash(ctx sdk.Context, hash []byte) (int64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBlockHash)
	bz := store.Get(hash)
	if len(bz) == 0 {
		return 0, false
	}

	height := binary.BigEndian.Uint64(bz)
	return int64(height), true
}

// SetBlockHash sets the mapping from block consensus hash to block height
func (k Keeper) SetBlockHash(ctx sdk.Context, hash []byte, height int64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBlockHash)
	bz := sdk.Uint64ToBigEndian(uint64(height))
	store.Set(hash, bz)
}

// ----------------------------------------------------------------------------
// Block bloom bits mapping functions
// Required by Web3 API.
// ----------------------------------------------------------------------------

// GetBlockBloom gets bloombits from block height
func (k Keeper) GetBlockBloom(ctx sdk.Context, height int64) (ethtypes.Bloom, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixBloom)
	bz := store.Get(types.BloomKey(height))
	if len(bz) == 0 {
		return ethtypes.Bloom{}, false
	}

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
		hash := common.BytesToHash(iterator.Key())
		var logs []*ethtypes.Log
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &logs)

		// add a new entry
		txLog := types.NewTransactionLogs(hash, logs)
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

package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
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
	// Transaction counter in a block. Used on StateSB's Prepare function.
	// It is reset to 0 every block on BeginBlock so there's no point in storing the counter
	// on the KVStore or adding it as a field on the EVM genesis state.
	TxCount int
	Bloom   *big.Int
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

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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
		return 0, fmt.Errorf("block with hash '%s' not found", ethcmn.BytesToHash(hash).Hex())
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
func (k *Keeper) SetTransactionLogs(ctx sdk.Context, hash []byte, logs []*ethtypes.Log) {
	store := ctx.KVStore(k.blockKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(logs)
	store.Set(types.LogsKey(hash), bz)
}

// GetTransactionLogs gets the logs for a transaction from the KVStore
func (k *Keeper) GetTransactionLogs(ctx sdk.Context, hash []byte) ([]*ethtypes.Log, error) {
	store := ctx.KVStore(k.blockKey)
	bz := store.Get(types.LogsKey(hash))
	if len(bz) == 0 {
		return nil, errors.New("cannot get transaction logs")
	}

	var logs []*ethtypes.Log
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &logs)
	return logs, nil
}

package keeper

import (
	"bytes"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/palantir/stacktrace"
	"github.com/tendermint/tendermint/libs/log"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/feemarket/types"
)

// Keeper grants access to the Fee Market module state.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec
	// Store key required for the Fee Market Prefix KVStore.
	storeKey sdk.StoreKey
	// module specific parameter space that can be configured through governance
	paramSpace paramtypes.Subspace
}

// NewKeeper generates new evm module keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey, paramSpace paramtypes.Subspace,
) *Keeper {

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramSpace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}

// ----------------------------------------------------------------------------
// Parent Block Gas Used
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBlockGasUsed returns the last block gas used value from the store.
func (k Keeper) GetBlockGasUsed(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBlockGasUsed)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetBlockGasUsed gets the current block gas consumed to the store.
// CONTRACT: this should be only called during EndBlock.
func (k Keeper) SetBlockGasUsed(ctx sdk.Context) {
	if ctx.BlockGasMeter() == nil {
		k.Logger(ctx).Error("block gas meter is nil when setting block gas used")
		return
	}

	store := ctx.KVStore(k.storeKey)
	gasBz := sdk.Uint64ToBigEndian(ctx.BlockGasMeter().GasConsumedToLimit())
	store.Set(types.KeyPrefixBlockGasUsed, gasBz)
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBlockGasUsed returns the last block gas used value from the store.
func (k Keeper) GetBaseFee(ctx sdk.Context) *big.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBaseFee)
	if len(bz) == 0 {
		return nil
	}

	return new(big.Int).SetBytes(bz)
}

// SetBlockGasUsed gets the current block gas consumed to the store.
// CONTRACT: this should be only called during EndBlock.
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee *big.Int) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyPrefixBaseFee, baseFee.Bytes())
}
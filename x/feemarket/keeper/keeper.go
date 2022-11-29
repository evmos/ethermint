package keeper

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	v010 "github.com/evmos/ethermint/x/feemarket/migrations/v010"
	"github.com/evmos/ethermint/x/feemarket/types"
)

// Keeper grants access to the Fee Market module state.
type Keeper struct {
	// Protobuf codec
	cdc codec.BinaryCodec
	// Store key required for the Fee Market Prefix KVStore.
	storeKey     storetypes.StoreKey
	transientKey storetypes.StoreKey
	// module specific parameter space that can be configured through governance
	paramSpace paramtypes.Subspace
}

// NewKeeper generates new fee market module keeper
func NewKeeper(
	cdc codec.BinaryCodec, paramSpace paramtypes.Subspace, storeKey, transientKey storetypes.StoreKey,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		paramSpace:   paramSpace,
		transientKey: transientKey,
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

// SetBlockGasWanted sets the block gas wanted to the store.
// CONTRACT: this should be only called during EndBlock.
func (k Keeper) SetBlockGasWanted(ctx sdk.Context, gas uint64) {
	store := ctx.KVStore(k.storeKey)
	gasBz := sdk.Uint64ToBigEndian(gas)
	store.Set(types.KeyPrefixBlockGasWanted, gasBz)
}

// GetBlockGasWanted returns the last block gas wanted value from the store.
func (k Keeper) GetBlockGasWanted(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixBlockGasWanted)
	if len(bz) == 0 {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// GetTransientGasWanted returns the gas wanted in the current block from transient store.
func (k Keeper) GetTransientGasWanted(ctx sdk.Context) uint64 {
	store := ctx.TransientStore(k.transientKey)
	bz := store.Get(types.KeyPrefixTransientBlockGasWanted)
	if len(bz) == 0 {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// SetTransientBlockGasWanted sets the block gas wanted to the transient store.
func (k Keeper) SetTransientBlockGasWanted(ctx sdk.Context, gasWanted uint64) {
	store := ctx.TransientStore(k.transientKey)
	gasBz := sdk.Uint64ToBigEndian(gasWanted)
	store.Set(types.KeyPrefixTransientBlockGasWanted, gasBz)
}

// AddTransientGasWanted adds the cumulative gas wanted in the transient store
func (k Keeper) AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error) {
	result := k.GetTransientGasWanted(ctx) + gasWanted
	k.SetTransientBlockGasWanted(ctx, result)
	return result, nil
}

// GetBaseFeeV1 get the base fee from v1 version of states.
// return nil if base fee is not enabled
func (k Keeper) GetBaseFeeV1(ctx sdk.Context) *big.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(v010.KeyPrefixBaseFeeV1)
	if len(bz) == 0 {
		return nil
	}
	return new(big.Int).SetBytes(bz)
}

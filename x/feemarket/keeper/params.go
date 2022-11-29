package keeper

import (
	"github.com/evmos/ethermint/x/feemarket/types"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var isTrue = []byte("0x01")

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	noBaseFee := k.GetNoBaseFee(ctx)
	baseFeeChangeDenom := k.GetBaseFeeChangeDenom(ctx)
	elasticityMultiplier := k.GetElasticityMultiplier(ctx)
	baseFee := k.getBaseFee(ctx)
	enableHeight := k.GetEnableHeight(ctx)
	minGasPrice := k.GetMinGasPrice(ctx)
	minGasPriceMultiplier := k.GetMinGasPriceMultiplier(ctx)

	return types.NewParams(noBaseFee, baseFeeChangeDenom, elasticityMultiplier, baseFee.Uint64(), enableHeight, minGasPrice, minGasPriceMultiplier)
}

// SetParams sets the fee market parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	k.setEnableHeight(ctx, params.EnableHeight)
	k.setElasticityMultiplier(ctx, params.ElasticityMultiplier)
	k.setBaseFeeChangeDenom(ctx, params.BaseFeeChangeDenominator)
	k.setNoBaseFee(ctx, params.NoBaseFee)
	k.setMinGasPrice(ctx, params.MinGasPrice)
	k.setMinGasMultiplier(ctx, params.MinGasMultiplier)
	k.setBaseFee(ctx, params.BaseFee.BigInt())

	return nil
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBaseFeeEnabled returns true if base fee is enabled
func (k Keeper) GetBaseFeeEnabled(ctx sdk.Context) bool {
	var enableHeight types.EnableHeight
	store := ctx.KVStore(k.storeKey)
	noBaseFee := store.Has(types.ParamStoreKeyNoBaseFee)
	enableHeightBz := store.Get(types.ParamStoreKeyEnableHeight)
	k.cdc.MustUnmarshal(enableHeightBz, &enableHeight)

	return !noBaseFee && ctx.BlockHeight() >= enableHeight.EnableHeight
}

// GetNoBaseFee gets the NoBaseFee from the store
func (k Keeper) GetNoBaseFee(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyNoBaseFee)
}

// GetBaseFeeChangeDenom gets the BaseFeeChangeDenominator from the store
func (k Keeper) GetBaseFeeChangeDenom(ctx sdk.Context) (baseFeeChangeDenom types.BaseFeeChangeDenominator) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyBaseFeeChangeDenominator)
	k.cdc.MustUnmarshal(bz, &baseFeeChangeDenom)
	return baseFeeChangeDenom
}

// GetElasticityMultiplier gets the ElasticityMultiplier from the store
func (k Keeper) GetElasticityMultiplier(ctx sdk.Context) (elasticityMultiplier types.ElasticityMultiplier) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyElasticityMultiplier)
	k.cdc.MustUnmarshal(bz, &elasticityMultiplier)
	return elasticityMultiplier
}

// GetMinGasPrice gets the MinGasPrice from the store
func (k Keeper) GetMinGasPrice(ctx sdk.Context) (minGasPrice sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyElasticityMultiplier)
	bigConverted := minGasPrice.BigInt().SetBytes(bz)
	return sdk.NewDec(bigConverted.Int64())
}

// GetMinGasPriceMultiplier gets the MinGasPriceMultiplier from the store
func (k Keeper) GetMinGasPriceMultiplier(ctx sdk.Context) (minGasPriceMultiplier sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyElasticityMultiplier)
	bigConverted := minGasPriceMultiplier.BigInt().SetBytes(bz)
	return sdk.NewDec(bigConverted.Int64())
}

// GetEnableHeight gets the EnableHeight from the store
func (k Keeper) GetEnableHeight(ctx sdk.Context) (enableHeight types.EnableHeight) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyElasticityMultiplier)
	k.cdc.MustUnmarshal(bz, &enableHeight)
	return enableHeight
}

// GetBaseFeeParam gets the BaseFee param from the store
func (k Keeper) getBaseFee(ctx sdk.Context) (baseFee *big.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyBaseFee)
	return baseFee.SetBytes(bz)
}

// GetBaseFee gets the base fee from the store
// return nil if base fee is not enabled
func (k Keeper) GetBaseFee(ctx sdk.Context) *big.Int {
	noBaseFee := k.GetNoBaseFee(ctx)
	if noBaseFee {
		return nil
	}

	// TODO: Not sure if this BaseFeeV1 format will be deleted
	baseFee := k.getBaseFee(ctx)
	if baseFee == nil || baseFee.Sign() == 0 {
		// try v1 format
		return k.GetBaseFeeV1(ctx)
	}

	return baseFee
}

// SetMinGasMultiplier gets the MinGasMultiplier from the store
func (k Keeper) SetMinGasMultiplier(ctx sdk.Context) (minGasMultiplier sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyElasticityMultiplier)
	bigConverted := minGasMultiplier.BigInt().SetBytes(bz)
	return sdk.NewDec(bigConverted.Int64())
}

// SetBaseFee set's the base fee in the store
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee *big.Int) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyBaseFee, baseFee.Bytes())
}

// setNoBaseFee sets the NoBaseFee to true or deletes it from the store
func (k Keeper) setNoBaseFee(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyNoBaseFee, isTrue)
		return
	}

	store.Delete(types.ParamStoreKeyNoBaseFee)
}

// setBaseFeeChangeDenom sets the BaseFeeChangeDenominator in the store
func (k Keeper) setBaseFeeChangeDenom(ctx sdk.Context, baseFeeChangeDenom *types.BaseFeeChangeDenominator) {
	store := ctx.KVStore(k.storeKey)
	baseFeeChangeDenomBz := k.cdc.MustMarshal(baseFeeChangeDenom)
	store.Set(types.ParamStoreKeyBaseFeeChangeDenominator, baseFeeChangeDenomBz)
}

// setElasticityMultiplier sets the ElasticityMultiplier in the store
func (k Keeper) setElasticityMultiplier(ctx sdk.Context, elasticityMultiplier *types.ElasticityMultiplier) {
	store := ctx.KVStore(k.storeKey)
	elasticityMultiplierBz := k.cdc.MustMarshal(elasticityMultiplier)
	store.Set(types.ParamStoreKeyElasticityMultiplier, elasticityMultiplierBz)
}

// setEnableHeight sets the EnableHeight in the store
func (k Keeper) setEnableHeight(ctx sdk.Context, enableHeight *types.EnableHeight) {
	store := ctx.KVStore(k.storeKey)
	enableHeightBz := k.cdc.MustMarshal(enableHeight)
	store.Set(types.ParamStoreKeyEnableHeight, enableHeightBz)
}

// setBaseFee sets the BaseFee in the store
func (k Keeper) setBaseFee(ctx sdk.Context, baseFee *big.Int) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyBaseFee, baseFee.Bytes())
}

// setMinGasPrice sets the MinGasPrice in the store
func (k Keeper) setMinGasPrice(ctx sdk.Context, minGasPrice sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyMinGasPrice, minGasPrice.BigInt().Bytes())
}

// setMinGasMultiplier sets the MinGasMultiplier in the store
func (k Keeper) setMinGasMultiplier(ctx sdk.Context, minGasMultiplier sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyMinGasMultiplier, minGasMultiplier.BigInt().Bytes())
}

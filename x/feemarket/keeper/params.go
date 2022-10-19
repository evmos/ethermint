package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/feemarket/types"
)

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// TODO: update once https://github.com/cosmos/cosmos-sdk/pull/12615 is merged
	// and released
	for _, pair := range params.ParamSetPairs() {
		k.paramSpace.GetIfExists(ctx, pair.Key, pair.Value)
	}
	return params
}

// SetParams sets the fee market parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBaseFeeEnabled returns true if base fee is enabled
func (k Keeper) GetBaseFeeEnabled(ctx sdk.Context) bool {
	noBaseFee := false
	enableHeight := int64(0)
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyNoBaseFee, &noBaseFee)
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyEnableHeight, &enableHeight)
	return !noBaseFee && ctx.BlockHeight() >= enableHeight
}

// GetBaseFee get's the base fee from the paramSpace
// return nil if base fee is not enabled
func (k Keeper) GetBaseFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	if params.NoBaseFee {
		return nil
	}

	baseFee := params.BaseFee.BigInt()
	if baseFee == nil || baseFee.Sign() == 0 {
		// try v1 format
		return k.GetBaseFeeV1(ctx)
	}

	return baseFee
}

// SetBaseFee set's the base fee in the paramSpace
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee *big.Int) {
	k.paramSpace.Set(ctx, types.ParamStoreKeyBaseFee, sdkmath.NewIntFromBigInt(baseFee))
}

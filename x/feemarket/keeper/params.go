package keeper

import (
	"math/big"

	"github.com/evmos/ethermint/x/feemarket/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams returns the total set of fee market parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the fee market params in a single key
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)

	return nil
}

// ----------------------------------------------------------------------------
// Parent Base Fee
// Required by EIP1559 base fee calculation.
// ----------------------------------------------------------------------------

// GetBaseFeeEnabled returns true if base fee is enabled
func (k Keeper) GetBaseFeeEnabled(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return !params.NoBaseFee && ctx.BlockHeight() >= params.EnableHeight
}

// GetBaseFee gets the base fee from the store
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

// SetBaseFee set's the base fee in the store
func (k Keeper) SetBaseFee(ctx sdk.Context, baseFee *big.Int) {
	params := k.GetParams(ctx)
	params.BaseFee = sdk.NewInt(baseFee.Int64())
	err := k.SetParams(ctx, params)
	if err != nil {
		// TODO: Do we throw error ?
		return
	}
}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyParams)
	if bz == nil {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)

	chainCfgBz := k.cdc.MustMarshal(&params.ChainConfig)
	allowUnprotectedTxsBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.AllowUnprotectedTxs})
	enableCallBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCall})
	enableCreateBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCreate})
	// TODO: Figure out how to marshal []int64
	//extraEIPsBz := k.cdc.MustMarshal(&params.ExtraEIPs)
	//store.Set(types.ParamStoreKeyExtraEIPs, params.ExtraEIPs)
	store.Set(types.ParamStoreKeyChainConfig, chainCfgBz)
	store.Set(types.ParamStoreKeyEVMDenom, []byte(params.EvmDenom))
	store.Set(types.ParamStoreKeyAllowUnprotectedTxs, allowUnprotectedTxsBz)
	store.Set(types.ParamStoreKeyEnableCall, enableCallBz)
	store.Set(types.ParamStoreKeyEnableCreate, enableCreateBz)

	return nil
}

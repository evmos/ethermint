package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	evmDenom := k.GetEVMDenom(ctx)
	enableCall := k.GetEnableCall(ctx)
	enableCreate := k.GetEnableCreate(ctx)
	chainCfg := k.GetChainConfig(ctx)
	extraEIPs := k.GetExtraEIPs(ctx)
	allowUnprotectedTx := k.GetAllowUnprotectedTxs(ctx)

	return types.NewParams(evmDenom, allowUnprotectedTx, enableCreate, enableCall, chainCfg, extraEIPs)
}

// SetParams Sets the EVM params each in its individual key for better get performance
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)

	chainCfgBz := k.cdc.MustMarshal(&params.ChainConfig)
	extraEIPsBz := k.cdc.MustMarshal(&params.ExtraEips)
	evmDenomBz := k.cdc.MustMarshal(&gogotypes.StringValue{Value: params.EvmDenom})
	allowUnprotectedTxsBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.AllowUnprotectedTxs})
	enableCallBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCall})
	enableCreateBz := k.cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCreate})

	store.Set(types.ParamStoreKeyExtraEIPs, extraEIPsBz)
	store.Set(types.ParamStoreKeyChainConfig, chainCfgBz)
	store.Set(types.ParamStoreKeyEVMDenom, evmDenomBz)
	store.Set(types.ParamStoreKeyAllowUnprotectedTxs, allowUnprotectedTxsBz)
	store.Set(types.ParamStoreKeyEnableCall, enableCallBz)
	store.Set(types.ParamStoreKeyEnableCreate, enableCreateBz)

	return nil
}

// GetExtraEIPs returns the extra EIPs enabled on the chain.
func (k Keeper) GetExtraEIPs(ctx sdk.Context) types.ExtraEIPs {
	var extraEIPs types.ExtraEIPs
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyExtraEIPs)
	if bz == nil {
		return extraEIPs
	}
	k.cdc.MustUnmarshal(bz, &extraEIPs)
	return extraEIPs
}

// GetChainConfig returns the chain configuration parameter.
func (k Keeper) GetChainConfig(ctx sdk.Context) types.ChainConfig {
	var chainCfg types.ChainConfig
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyChainConfig)
	if bz == nil {
		return chainCfg
	}
	k.cdc.MustUnmarshal(bz, &chainCfg)
	return chainCfg
}

// GetEVMDenom returns the EVM denom.
func (k Keeper) GetEVMDenom(ctx sdk.Context) string {
	var evmDenom gogotypes.StringValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEVMDenom)
	if bz == nil {
		return evmDenom.Value
	}
	k.cdc.MustUnmarshal(bz, &evmDenom)
	return evmDenom.Value
}

// GetEnableCall returns true if the EVM Call operation is enabled.
func (k Keeper) GetEnableCall(ctx sdk.Context) bool {
	var enableCall gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEnableCall)
	if bz == nil {
		return enableCall.Value
	}
	k.cdc.MustUnmarshal(bz, &enableCall)
	return enableCall.Value
}

// GetEnableCreate returns true if the EVM Create contract operation is enabled.
func (k Keeper) GetEnableCreate(ctx sdk.Context) bool {
	var enableCreate gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEnableCreate)
	if bz == nil {
		return enableCreate.Value
	}
	k.cdc.MustUnmarshal(bz, &enableCreate)
	return enableCreate.Value
}

// GetAllowUnprotectedTxs returns true if unprotected txs (i.e non-replay protected as per EIP-155) are supported by the chain.
func (k Keeper) GetAllowUnprotectedTxs(ctx sdk.Context) bool {
	var allowUnprotectedTx gogotypes.BoolValue
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyAllowUnprotectedTxs)
	if bz == nil {
		return allowUnprotectedTx.Value
	}
	k.cdc.MustUnmarshal(bz, &allowUnprotectedTx)
	return allowUnprotectedTx.Value
}

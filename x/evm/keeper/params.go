package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
)

var isTrue = []byte("0x01")

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	evmDenom := k.GetEVMDenom(ctx)
	allowUnprotectedTx := k.GetAllowUnprotectedTxs(ctx)
	enableCreate := k.GetEnableCreate(ctx)
	enableCall := k.GetEnableCall(ctx)
	chainCfg := k.GetChainConfig(ctx)
	extraEIPs := k.GetExtraEIPs(ctx)

	return types.NewParams(evmDenom, allowUnprotectedTx, enableCreate, enableCall, chainCfg, extraEIPs)
}

// SetParams sets the EVM params each in their individual key for better get performance
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	k.setExtraEIPs(ctx, params.ExtraEIPs)
	k.setChainConfig(ctx, params.ChainConfig)
	k.setEvmDenom(ctx, params.EvmDenom)
	k.setEnableCall(ctx, params.EnableCall)
	k.setEnableCreate(ctx, params.EnableCreate)
	k.setAllowUnprotectedTxs(ctx, params.AllowUnprotectedTxs)

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
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEVMDenom)
	if bz == nil {
		return ""
	}
	return string(bz)
}

// GetEnableCall returns true if the EVM Call operation is enabled.
func (k Keeper) GetEnableCall(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableCall)
}

// GetEnableCreate returns true if the EVM Create contract operation is enabled.
func (k Keeper) GetEnableCreate(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyEnableCreate)
}

// GetAllowUnprotectedTxs returns true if unprotected txs (i.e non-replay protected as per EIP-155) are supported by the chain.
func (k Keeper) GetAllowUnprotectedTxs(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ParamStoreKeyAllowUnprotectedTxs)
}

// setChainConfig sets the ChainConfig in the store
func (k Keeper) setChainConfig(ctx sdk.Context, chainCfg types.ChainConfig) {
	store := ctx.KVStore(k.storeKey)
	chainCfgBz := k.cdc.MustMarshal(&chainCfg)
	store.Set(types.ParamStoreKeyChainConfig, chainCfgBz)
}

// setExtraEIPs sets the ExtraEIPs in the store
func (k Keeper) setExtraEIPs(ctx sdk.Context, extraEIPs types.ExtraEIPs) {
	extraEIPsBz := k.cdc.MustMarshal(&extraEIPs)
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyExtraEIPs, extraEIPsBz)
}

// setEvmDenom sets the EVMDenom param in the store
func (k Keeper) setEvmDenom(ctx sdk.Context, evmDenom string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ParamStoreKeyEVMDenom, []byte(evmDenom))
}

// setAllowUnprotectedTxs sets the AllowUnprotectedTxs param in the store
func (k Keeper) setAllowUnprotectedTxs(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyAllowUnprotectedTxs, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyAllowUnprotectedTxs)
}

// setEnableCreate sets the EnableCreate param in the store
func (k Keeper) setEnableCreate(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableCreate, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableCreate)
}

// setEnableCall sets the EnableCall param in the store
func (k Keeper) setEnableCall(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableCall, isTrue)
		return
	}
	store.Delete(types.ParamStoreKeyEnableCall)
}

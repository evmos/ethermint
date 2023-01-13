// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
)

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

// GetLegacyParams returns param set for version before migrate
func (k Keeper) GetLegacyParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.ss.GetParamSetIfExists(ctx, &params)
	return params
}

// GetExtraEIPs returns the extra EIPs enabled on the chain.
func (k Keeper) GetExtraEIPs(ctx sdk.Context) types.ExtraEIPs {
	var extraEIPs types.ExtraEIPs
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyExtraEIPs)
	if len(bz) == 0 {
		return k.GetLegacyParams(ctx).ExtraEIPs
	}
	k.cdc.MustUnmarshal(bz, &extraEIPs)
	return extraEIPs
}

// GetChainConfig returns the chain configuration parameter.
func (k Keeper) GetChainConfig(ctx sdk.Context) types.ChainConfig {
	var chainCfg types.ChainConfig
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyChainConfig)
	if len(bz) == 0 {
		return k.GetLegacyParams(ctx).ChainConfig
	}
	k.cdc.MustUnmarshal(bz, &chainCfg)
	return chainCfg
}

// GetEVMDenom returns the EVM denom.
func (k Keeper) GetEVMDenom(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamStoreKeyEVMDenom)
	if len(bz) == 0 {
		return k.GetLegacyParams(ctx).EvmDenom
	}
	return string(bz)
}

// GetEnableCall returns true if the EVM Call operation is enabled.
func (k Keeper) GetEnableCall(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	exist := store.Has(types.ParamStoreKeyEnableCall)
	if !exist {
		exist = k.GetLegacyParams(ctx).EnableCall
	}
	return exist
}

// GetEnableCreate returns true if the EVM Create contract operation is enabled.
func (k Keeper) GetEnableCreate(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	exist := store.Has(types.ParamStoreKeyEnableCreate)
	if !exist {
		exist = k.GetLegacyParams(ctx).EnableCreate
	}
	return exist
}

// GetAllowUnprotectedTxs returns true if unprotected txs (i.e non-replay protected as per EIP-155) are supported by the chain.
func (k Keeper) GetAllowUnprotectedTxs(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	exist := store.Has(types.ParamStoreKeyAllowUnprotectedTxs)
	if !exist {
		exist = k.GetLegacyParams(ctx).AllowUnprotectedTxs
	}
	return exist
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
		store.Set(types.ParamStoreKeyAllowUnprotectedTxs, []byte{0x01})
		return
	}
	store.Delete(types.ParamStoreKeyAllowUnprotectedTxs)
}

// setEnableCreate sets the EnableCreate param in the store
func (k Keeper) setEnableCreate(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableCreate, []byte{0x01})
		return
	}
	store.Delete(types.ParamStoreKeyEnableCreate)
}

// setEnableCall sets the EnableCall param in the store
func (k Keeper) setEnableCall(ctx sdk.Context, enable bool) {
	store := ctx.KVStore(k.storeKey)
	if enable {
		store.Set(types.ParamStoreKeyEnableCall, []byte{0x01})
		return
	}
	store.Delete(types.ParamStoreKeyEnableCall)
}

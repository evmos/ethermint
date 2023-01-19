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
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if len(bz) == 0 {
		var p types.Params
		k.ss.GetParamSetIfExists(ctx, &p)
		return p
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the EVM params each in their individual key for better get performance
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store.Set(types.ParamsKey, bz)
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
	params := k.GetParams(ctx)
	return params.ExtraEIPs
}

// GetChainConfig returns the chain configuration parameter.
func (k Keeper) GetChainConfig(ctx sdk.Context) types.ChainConfig {
	params := k.GetParams(ctx)
	return params.ChainConfig
}

// GetEVMDenom returns the EVM denom.
func (k Keeper) GetEVMDenom(ctx sdk.Context) string {
	params := k.GetParams(ctx)
	return params.EvmDenom
}

// GetEnableCall returns true if the EVM Call operation is enabled.
func (k Keeper) GetEnableCall(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return params.EnableCall
}

// GetEnableCreate returns true if the EVM Create contract operation is enabled.
func (k Keeper) GetEnableCreate(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return params.EnableCreate
}

// GetAllowUnprotectedTxs returns true if unprotected txs (i.e non-replay protected as per EIP-155) are supported by the chain.
func (k Keeper) GetAllowUnprotectedTxs(ctx sdk.Context) bool {
	params := k.GetParams(ctx)
	return params.AllowUnprotectedTxs
}

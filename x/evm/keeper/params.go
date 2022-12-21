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
	// TODO: update once https://github.com/cosmos/cosmos-sdk/pull/12615 is merged
	// and released
	for _, pair := range params.ParamSetPairs() {
		k.paramSpace.GetIfExists(ctx, pair.Key, pair.Value)
	}

	return params
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetChainConfig returns the chain configuration parameter.
func (k Keeper) GetChainConfig(ctx sdk.Context) types.ChainConfig {
	var chainCfg types.ChainConfig
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyChainConfig, &chainCfg)
	return chainCfg
}

// GetEVMDenom returns the EVM denom.
func (k Keeper) GetEVMDenom(ctx sdk.Context) string {
	var evmDenom string
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyEVMDenom, &evmDenom)
	return evmDenom
}

// GetEnableCall returns true if the EVM Call operation is enabled.
func (k Keeper) GetEnableCall(ctx sdk.Context) bool {
	var enableCall bool
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyEnableCall, &enableCall)
	return enableCall
}

// GetEnableCreate returns true if the EVM Create contract operation is enabled.
func (k Keeper) GetEnableCreate(ctx sdk.Context) bool {
	var enableCreate bool
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyEnableCreate, &enableCreate)
	return enableCreate
}

// GetAllowUnprotectedTxs returns true if unprotected txs (i.e non-replay protected as per EIP-155) are supported by the chain.
func (k Keeper) GetAllowUnprotectedTxs(ctx sdk.Context) bool {
	var allowUnprotectedTx bool
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyAllowUnprotectedTxs, &allowUnprotectedTx)
	return allowUnprotectedTx
}

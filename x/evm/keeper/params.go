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

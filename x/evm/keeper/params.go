package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/x/evm/types"
)

// GetParams returns the total set of evm parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the evm parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetChainConfig gets block height from block consensus hash
func (k Keeper) GetChainConfig(ctx sdk.Context) (types.ChainConfig, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyPrefixChainConfig)
	if len(bz) == 0 {
		return types.ChainConfig{}, false
	}

	var config types.ChainConfig
	k.cdc.MustUnmarshal(bz, &config)
	return config, true
}

// SetChainConfig sets the mapping from block consensus hash to block height
func (k Keeper) SetChainConfig(ctx sdk.Context, config types.ChainConfig) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&config)
	store.Set(types.KeyPrefixChainConfig, bz)
}

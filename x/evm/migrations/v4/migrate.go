package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/exported"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// MigrateStore migrates the x/evm module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/evm module state.
func MigrateStore(
	ctx sdk.Context,
	store sdk.KVStore,
	legacySubspace exported.Subspace,
	cdc codec.BinaryCodec,
) error {
	var params v4types.Params
	legacySubspace.GetParams(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}
	
	chainCfgBz := cdc.MustMarshal(&params.ChainConfig)
	extraEIPsBz := cdc.MustMarshal(&params.ExtraEips)
	evmDenomBz := cdc.MustMarshal(&gogotypes.StringValue{Value: params.EvmDenom})
	allowUnprotectedTxsBz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.AllowUnprotectedTxs})
	enableCallBz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCall})
	enableCreateBz := cdc.MustMarshal(&gogotypes.BoolValue{Value: params.EnableCreate})

	store.Set(v4types.ParamStoreKeyExtraEIPs, extraEIPsBz)
	store.Set(v4types.ParamStoreKeyChainConfig, chainCfgBz)
	store.Set(v4types.ParamStoreKeyEVMDenom, evmDenomBz)
	store.Set(v4types.ParamStoreKeyAllowUnprotectedTxs, allowUnprotectedTxsBz)
	store.Set(v4types.ParamStoreKeyEnableCall, enableCallBz)
	store.Set(v4types.ParamStoreKeyEnableCreate, enableCreateBz)

	return nil
}

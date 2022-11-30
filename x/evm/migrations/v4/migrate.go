package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	"github.com/evmos/ethermint/x/evm/types"
)

var isTrue = []byte("0x01")

// MigrateStore migrates the x/evm module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the Cosmos SDK params module and stores them directly into the x/evm module state.
func MigrateStore(
	ctx sdk.Context,
	storeKey storetypes.StoreKey,
	legacySubspace types.Subspace,
	cdc codec.BinaryCodec,
) error {
	store := ctx.KVStore(storeKey)
	var params v4types.Params

	legacySubspace.GetParamSetIfExists(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	chainCfgBz := cdc.MustMarshal(&params.ChainConfig)
	extraEIPsBz := cdc.MustMarshal(&v4types.ExtraEIPs{ExtraEIPs: v4types.AvailableExtraEIPs})

	store.Set(v4types.ParamStoreKeyEVMDenom, []byte(params.EvmDenom))
	store.Set(v4types.ParamStoreKeyExtraEIPs, extraEIPsBz)
	store.Set(v4types.ParamStoreKeyChainConfig, chainCfgBz)

	if params.AllowUnprotectedTxs {
		store.Set(v4types.ParamStoreKeyAllowUnprotectedTxs, isTrue)
	}

	if params.EnableCall {
		store.Set(v4types.ParamStoreKeyEnableCall, isTrue)
	}

	if params.EnableCreate {
		store.Set(v4types.ParamStoreKeyEnableCreate, isTrue)
	}

	return nil
}

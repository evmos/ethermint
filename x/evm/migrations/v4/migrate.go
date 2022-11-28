package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	"github.com/evmos/ethermint/x/evm/types"
)

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
	legacySubspace.GetParamSet(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	chainCfgBz := cdc.MustMarshal(&params.ChainConfig)
	extraEIPsBz := cdc.MustMarshal(&v4types.ExtraEIPs{ExtraEIPs: v4types.AvailableExtraEIPs})

	evmDenomBz := []byte(params.EvmDenom)

	allowUnprotectedTxsBz := []byte("0x00")
	if params.AllowUnprotectedTxs {
		allowUnprotectedTxsBz = []byte("0x01")
	}

	enableCallBz := []byte("0x00")
	if params.EnableCall {
		enableCallBz = []byte("0x01")
	}

	enableCreateBz := []byte("0x00")
	if params.EnableCreate {
		enableCreateBz = []byte("0x01")
	}

	store.Set(v4types.ParamStoreKeyExtraEIPs, extraEIPsBz)
	store.Set(v4types.ParamStoreKeyChainConfig, chainCfgBz)
	store.Set(v4types.ParamStoreKeyEVMDenom, evmDenomBz)
	store.Set(v4types.ParamStoreKeyAllowUnprotectedTxs, allowUnprotectedTxsBz)
	store.Set(v4types.ParamStoreKeyEnableCall, enableCallBz)
	store.Set(v4types.ParamStoreKeyEnableCreate, enableCreateBz)

	return nil
}

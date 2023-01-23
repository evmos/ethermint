package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/x/evm/types"
)

var (
	ParamStoreKeyEVMDenom            = []byte("EVMDenom")
	ParamStoreKeyEnableCreate        = []byte("EnableCreate")
	ParamStoreKeyEnableCall          = []byte("EnableCall")
	ParamStoreKeyExtraEIPs           = []byte("EnableExtraEIPs")
	ParamStoreKeyChainConfig         = []byte("ChainConfig")
	ParamStoreKeyAllowUnprotectedTxs = []byte("AllowUnprotectedTxs")
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
	var params types.Params

	legacySubspace.GetParamSetIfExists(ctx, &params)

	if err := params.Validate(); err != nil {
		params = migratevRC1(cdc, ctx, storeKey)
	}

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	store := ctx.KVStore(storeKey)
	store.Set(types.KeyPrefixParams, bz)

	return nil
}

func migratevRC1(cdc codec.BinaryCodec, ctx sdk.Context, storeKey storetypes.StoreKey) (params types.Params) {
	var (
		denom       string
		extraEIPs   types.ExtraEIPs
		chainConfig types.ChainConfig
	)

	store := ctx.KVStore(storeKey)
	if store.Has(ParamStoreKeyEVMDenom) {
		denom = string(store.Get(ParamStoreKeyEVMDenom))
		store.Delete(ParamStoreKeyEVMDenom)
	}

	if store.Has(ParamStoreKeyExtraEIPs) {
		extraEIPsBz := store.Get(ParamStoreKeyExtraEIPs)
		cdc.MustUnmarshal(extraEIPsBz, &extraEIPs)
		store.Delete(ParamStoreKeyExtraEIPs)
	}

	if store.Has(ParamStoreKeyChainConfig) {
		chainCfgBz := store.Get(ParamStoreKeyChainConfig)
		cdc.MustUnmarshal(chainCfgBz, &chainConfig)
		store.Delete(ParamStoreKeyChainConfig)
	}

	params.EvmDenom = denom
	params.ExtraEIPs = extraEIPs
	params.ChainConfig = chainConfig
	params.EnableCreate = store.Has(ParamStoreKeyEnableCreate)
	params.EnableCall = store.Has(ParamStoreKeyEnableCall)
	params.AllowUnprotectedTxs = store.Has(ParamStoreKeyAllowUnprotectedTxs)

	store.Delete(ParamStoreKeyEnableCreate)
	store.Delete(ParamStoreKeyEnableCall)
	store.Delete(ParamStoreKeyAllowUnprotectedTxs)

	return params
}

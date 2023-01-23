package v4

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
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

	err := params.Validate()
	if err != nil {
		params = migratevRC1(cdc, ctx, storeKey)
	} else {
		params.ChainConfig = types.ChainConfig{
			HomesteadBlock:      params.ChainConfig.HomesteadBlock,
			DAOForkBlock:        params.ChainConfig.DAOForkBlock,
			DAOForkSupport:      params.ChainConfig.DAOForkSupport,
			EIP150Block:         params.ChainConfig.EIP150Block,
			EIP150Hash:          params.ChainConfig.EIP150Hash,
			EIP155Block:         params.ChainConfig.EIP155Block,
			EIP158Block:         params.ChainConfig.EIP158Block,
			ByzantiumBlock:      params.ChainConfig.ByzantiumBlock,
			ConstantinopleBlock: params.ChainConfig.ConstantinopleBlock,
			PetersburgBlock:     params.ChainConfig.PetersburgBlock,
			IstanbulBlock:       params.ChainConfig.IstanbulBlock,
			MuirGlacierBlock:    params.ChainConfig.MuirGlacierBlock,
			BerlinBlock:         params.ChainConfig.BerlinBlock,
			LondonBlock:         params.ChainConfig.LondonBlock,
			GrayGlacierBlock:    params.ChainConfig.GrayGlacierBlock,
			ArrowGlacierBlock:   params.ChainConfig.ArrowGlacierBlock,
			MergeNetsplitBlock:  params.ChainConfig.MergeNetsplitBlock,
			ShanghaiBlock:       params.ChainConfig.ShanghaiBlock,
			CancunBlock:         params.ChainConfig.CancunBlock,
		}
	}

	if err := params.Validate(); err != nil {
		return err
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
		extraEIPs   v4types.V4ExtraEIPs
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
	params.ExtraEIPs = extraEIPs.EIPs
	params.ChainConfig = chainConfig
	params.EnableCreate = store.Has(ParamStoreKeyEnableCreate)
	params.EnableCall = store.Has(ParamStoreKeyEnableCall)
	params.AllowUnprotectedTxs = store.Has(ParamStoreKeyAllowUnprotectedTxs)

	store.Delete(ParamStoreKeyEnableCreate)
	store.Delete(ParamStoreKeyEnableCall)
	store.Delete(ParamStoreKeyAllowUnprotectedTxs)

	return params
}

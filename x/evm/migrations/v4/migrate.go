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
	var (
		v4Params v4types.V4Params
		params   types.Params
	)

	legacySubspace.GetParamSetIfExists(ctx, &v4Params)

	err := v4Params.Validate()
	if err != nil {
		params = migratevRC1(cdc, ctx, storeKey)
	} else {
		params.AllowUnprotectedTxs = v4Params.AllowUnprotectedTxs
		params.EnableCall = v4Params.EnableCall
		params.EnableCreate = v4Params.EnableCreate
		params.EvmDenom = v4Params.EvmDenom
		params.ExtraEIPs.EIPs = v4Params.ExtraEIPs
		params.ChainConfig = types.ChainConfig{
			HomesteadBlock:      v4Params.ChainConfig.HomesteadBlock,
			DAOForkBlock:        v4Params.ChainConfig.DAOForkBlock,
			DAOForkSupport:      v4Params.ChainConfig.DAOForkSupport,
			EIP150Block:         v4Params.ChainConfig.EIP150Block,
			EIP150Hash:          v4Params.ChainConfig.EIP150Hash,
			EIP155Block:         v4Params.ChainConfig.EIP155Block,
			EIP158Block:         v4Params.ChainConfig.EIP158Block,
			ByzantiumBlock:      v4Params.ChainConfig.ByzantiumBlock,
			ConstantinopleBlock: v4Params.ChainConfig.ConstantinopleBlock,
			PetersburgBlock:     v4Params.ChainConfig.PetersburgBlock,
			IstanbulBlock:       v4Params.ChainConfig.IstanbulBlock,
			MuirGlacierBlock:    v4Params.ChainConfig.MuirGlacierBlock,
			BerlinBlock:         v4Params.ChainConfig.BerlinBlock,
			LondonBlock:         v4Params.ChainConfig.LondonBlock,
			ArrowGlacierBlock:   v4Params.ChainConfig.ArrowGlacierBlock,
			MergeNetsplitBlock:  v4Params.ChainConfig.MergeNetsplitBlock,
			ShanghaiBlock:       v4Params.ChainConfig.ShanghaiBlock,
			CancunBlock:         v4Params.ChainConfig.CancunBlock,
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

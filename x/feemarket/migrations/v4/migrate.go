package v4

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4types "github.com/evmos/ethermint/x/feemarket/migrations/v4/types"
	"github.com/evmos/ethermint/x/feemarket/types"
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

	legacySubspace.GetParamSetIfExists(ctx, &params)

	if err := params.Validate(); err != nil {
		return err
	}

	fmt.Println(params)
	baseFeeChangeDenom := cdc.MustMarshal(params.BaseFeeChangeDenominator)
	elasticityMultiplier := cdc.MustMarshal(params.ElasticityMultiplier)
	enableHeight := cdc.MustMarshal(params.EnableHeight)
	baseFeeBz := params.BaseFee.BigInt().Bytes()
	minGasPriceBz := params.MinGasPrice.BigInt().Bytes()
	minGasMultiplierBz := params.MinGasMultiplier.BigInt().Bytes()

	store.Set(v4types.ParamStoreKeyBaseFeeChangeDenominator, baseFeeChangeDenom)
	store.Set(v4types.ParamStoreKeyElasticityMultiplier, elasticityMultiplier)
	store.Set(v4types.ParamStoreKeyEnableHeight, enableHeight)
	store.Set(v4types.ParamStoreKeyBaseFee, baseFeeBz)
	store.Set(v4types.ParamStoreKeyMinGasPrice, []byte(minGasPriceBz))
	store.Set(v4types.ParamStoreKeyMinGasMultiplier, []byte(minGasMultiplierBz))

	if params.NoBaseFee {
		store.Set(v4types.ParamStoreKeyNoBaseFee, []byte{0x01})
	}

	return nil
}

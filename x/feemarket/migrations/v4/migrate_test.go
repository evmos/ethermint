package v4_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v4 "github.com/evmos/ethermint/x/feemarket/migrations/v4"
	v4types "github.com/evmos/ethermint/x/feemarket/migrations/v4/types"
	"github.com/evmos/ethermint/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps v4types.Params
}

func newMockSubspace(ps v4types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v4types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(v4types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	// Get all the new parameters from the kvStore
	var baseFeeChangeDenom types.BaseFeeChangeDenominator
	bz := kvStore.Get(v4types.ParamStoreKeyBaseFeeChangeDenominator)
	cdc.MustUnmarshal(bz, &baseFeeChangeDenom)

	var elasticityMultiplier types.ElasticityMultiplier
	bz = kvStore.Get(v4types.ParamStoreKeyElasticityMultiplier)
	cdc.MustUnmarshal(bz, &elasticityMultiplier)

	var enableHeight types.EnableHeight
	bz = kvStore.Get(v4types.ParamStoreKeyEnableHeight)
	cdc.MustUnmarshal(bz, &enableHeight)

	var baseFee big.Int
	bz = kvStore.Get(v4types.ParamStoreKeyBaseFee)
	baseFee.SetBytes(bz)

	var minGasPrice big.Int
	bz = kvStore.Get(v4types.ParamStoreKeyMinGasPrice)
	minGasPrice.SetBytes(bz)

	var minGasMultiplier big.Int
	bz = kvStore.Get(v4types.ParamStoreKeyMinGasMultiplier)
	minGasMultiplier.SetBytes(bz)

	noBaseFee := kvStore.Has(v4types.ParamStoreKeyNoBaseFee)

	params := types.NewParams(noBaseFee, baseFeeChangeDenom, elasticityMultiplier, baseFee.Uint64(), enableHeight,
		sdk.NewDec(minGasPrice.Int64()), sdk.NewDec(minGasMultiplier.Int64()),
	)

	fmt.Println(params)

	outcome := reflect.DeepEqual(params, legacySubspace.ps)
	require.Equal(t, outcome, true)
}

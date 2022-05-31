package v011_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tharsis/ethermint/encoding"

	"github.com/tharsis/ethermint/app"
	v010types "github.com/tharsis/ethermint/x/feemarket/migrations/v010/types"
	v011 "github.com/tharsis/ethermint/x/feemarket/migrations/v011"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

func TestMigrateStore(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	feemarketKey := sdk.NewKVStoreKey(feemarkettypes.StoreKey)
	tFeeMarketKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", feemarkettypes.StoreKey))
	ctx := testutil.DefaultContext(feemarketKey, tFeeMarketKey)
	paramstore := paramtypes.NewSubspace(
		encCfg.Marshaler, encCfg.Amino, feemarketKey, tFeeMarketKey, "feemarket",
	)

	paramstore = paramstore.WithKeyTable(feemarkettypes.ParamKeyTable())
	require.True(t, paramstore.HasKeyTable())

	// check no MinGasPrice param
	require.False(t, paramstore.Has(ctx, feemarkettypes.ParamStoreKeyMinGasPrice))

	// Run migrations
	err := v011.MigrateStore(ctx, &paramstore)
	require.NoError(t, err)

	// Make sure the params are set
	require.True(t, paramstore.Has(ctx, feemarkettypes.ParamStoreKeyMinGasPrice))

	var minGasPrice sdk.Dec

	// Make sure the new params are set
	require.NotPanics(t, func() {
		paramstore.Get(ctx, feemarkettypes.ParamStoreKeyMinGasPrice, &minGasPrice)
	})

	// check the params are updated
	require.True(t, minGasPrice.IsZero())
}

func TestMigrateJSON(t *testing.T) {
	rawJson := `{
		"block_gas": "0",
		"params": {
			"base_fee_change_denominator": 8,
			"elasticity_multiplier": 2,
			"enable_height": "0",
			"base_fee": "1000000000",
			"no_base_fee": false
		}
  }`
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	var genState v010types.GenesisState
	err := encCfg.Marshaler.UnmarshalJSON([]byte(rawJson), &genState)
	require.NoError(t, err)

	migratedGenState := v011.MigrateJSON(genState)

	require.True(t, migratedGenState.Params.MinGasPrice.IsZero())
}

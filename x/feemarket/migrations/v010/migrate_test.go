package v010_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tharsis/ethermint/encoding"

	"github.com/tharsis/ethermint/app"
	feemarketkeeper "github.com/tharsis/ethermint/x/feemarket/keeper"
	v010 "github.com/tharsis/ethermint/x/feemarket/migrations/v010"
	"github.com/tharsis/ethermint/x/feemarket/types"
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
	fmKeeper := feemarketkeeper.NewKeeper(encCfg.Marshaler, feemarketKey, paramstore)
	fmKeeper.SetParams(ctx, types.DefaultParams())
	require.True(t, paramstore.HasKeyTable())

	// check that the fee market is not nil
	err := v010.MigrateStore(ctx, &paramstore, feemarketKey)
	require.NoError(t, err)
	require.False(t, ctx.KVStore(feemarketKey).Has(v010.KeyPrefixBaseFeeV1))

	params := fmKeeper.GetParams(ctx)
	require.False(t, params.BaseFee.IsNil())

	baseFee := fmKeeper.GetBaseFee(ctx)
	require.NotNil(t, baseFee)

	require.Equal(t, baseFee.Int64(), params.BaseFee.Int64())
}

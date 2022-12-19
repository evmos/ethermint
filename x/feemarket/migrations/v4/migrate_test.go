package v4_test

import (
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

func newMockSubspaceEmpty() mockSubspace {
	return mockSubspace{}
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

	legacySubspaceEmpty := newMockSubspaceEmpty()
	require.Error(t, v4.MigrateStore(ctx, storeKey, legacySubspaceEmpty, cdc))

	legacySubspace := newMockSubspace(v4types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	paramsBz := kvStore.Get(v4types.ParamsKey)
	var params v4types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	require.Equal(t, params, legacySubspace.ps)
}

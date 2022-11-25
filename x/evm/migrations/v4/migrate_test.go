package v4_test

import (
	"github.com/evmos/ethermint/x/evm/types"
	gogotypes "github.com/gogo/protobuf/types"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v4 "github.com/evmos/ethermint/x/evm/migrations/v4"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(v4types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, store, legacySubspace, cdc))

	// Get all the new parameters from the store
	var evmDenom gogotypes.StringValue
	bz := store.Get(v4types.ParamStoreKeyEVMDenom)
	cdc.MustUnmarshal(bz, &evmDenom)

	var allowUnprotectedTx gogotypes.BoolValue
	bz = store.Get(v4types.ParamStoreKeyAllowUnprotectedTxs)
	cdc.MustUnmarshal(bz, &allowUnprotectedTx)

	var enableCreate gogotypes.BoolValue
	bz = store.Get(v4types.ParamStoreKeyEnableCreate)
	cdc.MustUnmarshal(bz, &enableCreate)

	var enableCall gogotypes.BoolValue
	bz = store.Get(v4types.ParamStoreKeyEnableCall)
	cdc.MustUnmarshal(bz, &enableCall)

	var chainCfg types.ChainConfig
	bz = store.Get(v4types.ParamStoreKeyChainConfig)
	cdc.MustUnmarshal(bz, &chainCfg)

	var extraEIPs types.ExtraEIPs
	bz = store.Get(v4types.ParamStoreKeyExtraEIPs)
	cdc.MustUnmarshal(bz, &extraEIPs)

	params := types.NewParams(evmDenom.Value, allowUnprotectedTx.Value, enableCreate.Value, enableCall.Value, chainCfg, extraEIPs)
	require.Equal(t, legacySubspace.ps, params)
}

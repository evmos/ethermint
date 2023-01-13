package v4_test

import (
	"testing"

	"github.com/evmos/ethermint/x/evm/types"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v4 "github.com/evmos/ethermint/x/evm/migrations/v4"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.DefaultParams())
	require.NoError(t, v4.MigrateStore(ctx, storeKey, legacySubspace, cdc))

	// Get all the new parameters from the kvStore
	var evmDenom string
	bz := kvStore.Get(types.ParamStoreKeyEVMDenom)
	evmDenom = string(bz)

	var allowUnprotectedTx gogotypes.BoolValue
	bz = kvStore.Get(types.ParamStoreKeyAllowUnprotectedTxs)
	cdc.MustUnmarshal(bz, &allowUnprotectedTx)

	enableCreate := kvStore.Has(types.ParamStoreKeyEnableCreate)
	enableCall := kvStore.Has(types.ParamStoreKeyEnableCall)

	var chainCfg types.ChainConfig
	bz = kvStore.Get(types.ParamStoreKeyChainConfig)
	cdc.MustUnmarshal(bz, &chainCfg)

	var extraEIPs types.ExtraEIPs
	bz = kvStore.Get(types.ParamStoreKeyExtraEIPs)
	cdc.MustUnmarshal(bz, &extraEIPs)
	require.Equal(t, types.AvailableExtraEIPs, extraEIPs.EIPs)

	params := types.NewParams(evmDenom, allowUnprotectedTx.Value, enableCreate, enableCall, chainCfg, extraEIPs)
	require.Equal(t, legacySubspace.ps, params)
}

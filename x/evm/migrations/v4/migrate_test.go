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
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	"github.com/stretchr/testify/require"
)

type mockSubspace struct {
	ps v4types.V4Params
}

func newMockSubspace(ps v4types.V4Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v4types.V4Params) = ms.ps
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
	var evmDenom string
	bz := kvStore.Get(v4types.ParamStoreKeyEVMDenom)
	evmDenom = string(bz)

	var allowUnprotectedTx gogotypes.BoolValue
	bz = kvStore.Get(v4types.ParamStoreKeyAllowUnprotectedTxs)
	cdc.MustUnmarshal(bz, &allowUnprotectedTx)

	enableCreate := kvStore.Has(v4types.ParamStoreKeyEnableCreate)
	enableCall := kvStore.Has(v4types.ParamStoreKeyEnableCall)

	var chainCfg v4types.ChainConfig
	bz = kvStore.Get(v4types.ParamStoreKeyChainConfig)
	cdc.MustUnmarshal(bz, &chainCfg)

	var extraEIPs v4types.ExtraEIPs
	bz = kvStore.Get(v4types.ParamStoreKeyExtraEIPs)
	cdc.MustUnmarshal(bz, &extraEIPs)

	params := v4types.NewParams(evmDenom, allowUnprotectedTx.Value, enableCreate, enableCall, chainCfg, extraEIPs)
	require.Equal(t, legacySubspace.ps, params)
}

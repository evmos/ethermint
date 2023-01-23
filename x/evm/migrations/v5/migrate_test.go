package v5_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	v5 "github.com/evmos/ethermint/x/evm/migrations/v5"
	v5types "github.com/evmos/ethermint/x/evm/migrations/v5/types"
	"github.com/evmos/ethermint/x/evm/types"
)

func TestMigrate(t *testing.T) {
	encCfg := encoding.MakeConfig(app.ModuleBasics)
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(types.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	kvStore := ctx.KVStore(storeKey)

	extraEIPs := v5types.V5ExtraEIPs{EIPs: types.AvailableExtraEIPs}
	extraEIPsBz := cdc.MustMarshal(&extraEIPs)
	chainConfig := types.DefaultChainConfig()
	chainConfigBz := cdc.MustMarshal(&chainConfig)

	// Set the params in the store
	kvStore.Set(v5.ParamStoreKeyEVMDenom, []byte("aphoton"))
	kvStore.Set(v5.ParamStoreKeyEnableCreate, []byte{0x01})
	kvStore.Set(v5.ParamStoreKeyEnableCall, []byte{0x01})
	kvStore.Set(v5.ParamStoreKeyExtraEIPs, extraEIPsBz)
	kvStore.Set(v5.ParamStoreKeyChainConfig, chainConfigBz)

	paramsBz := kvStore.Get(types.KeyPrefixParams)
	var params types.Params
	cdc.MustUnmarshal(paramsBz, &params)

	// TODO: test
	// require.Equal(t, params, legacySubspace.ps)
}

package v2_test

// TODO: Figure out if these will be deleted
//func TestMigrateStore(t *testing.T) {
//	encCfg := encoding.MakeConfig(app.ModuleBasics)
//	feemarketKey := sdk.NewKVStoreKey(types.StoreKey)
//	tFeeMarketKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", types.StoreKey))
//	ctx := testutil.DefaultContext(feemarketKey, tFeeMarketKey)
//	paramstore := paramtypes.NewSubspace(
//		encCfg.Codec, encCfg.Amino, feemarketKey, tFeeMarketKey, "evm",
//	).WithKeyTable(v2types.ParamKeyTable())
//
//	params := v2types.DefaultParams()
//	paramstore.SetParamSet(ctx, &params)
//
//	require.Panics(t, func() {
//		var result bool
//		paramstore.Get(ctx, types.ParamStoreKeyAllowUnprotectedTxs, &result)
//	})
//
//	paramstore = paramtypes.NewSubspace(
//		encCfg.Codec, encCfg.Amino, feemarketKey, tFeeMarketKey, "evm",
//	).WithKeyTable(v2types.ParamKeyTable())
//	err := v2.MigrateStore(ctx, &paramstore)
//	require.NoError(t, err)
//
//	var result bool
//	paramstore.Get(ctx, types.ParamStoreKeyAllowUnprotectedTxs, &result)
//	require.Equal(t, types.DefaultAllowUnprotectedTxs, result)
//}

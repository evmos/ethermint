package v3_test

// TODO: Figure out if these will be deleted
//func TestMigrateStore(t *testing.T) {
//	encCfg := encoding.MakeConfig(app.ModuleBasics)
//	evmKey := sdk.NewKVStoreKey(types.StoreKey)
//	tEvmKey := sdk.NewTransientStoreKey(fmt.Sprintf("%s_test", types.StoreKey))
//	ctx := testutil.DefaultContext(evmKey, tEvmKey)
//	paramstore := paramtypes.NewSubspace(
//		encCfg.Codec, encCfg.Amino, evmKey, tEvmKey, "evm",
//	).WithKeyTable(v3types.ParamKeyTable())
//
//	params := v3types.DefaultParams()
//	paramstore.SetParamSet(ctx, &params)
//
//	require.Panics(t, func() {
//		var preMigrationConfig types.ChainConfig
//		paramstore.Get(ctx, types.ParamStoreKeyChainConfig, &preMigrationConfig)
//	})
//	var preMigrationConfig v3types.ChainConfig
//	paramstore.Get(ctx, types.ParamStoreKeyChainConfig, &preMigrationConfig)
//	require.NotNil(t, preMigrationConfig.MergeForkBlock)
//
//	paramstore = paramtypes.NewSubspace(
//		encCfg.Codec, encCfg.Amino, evmKey, tEvmKey, "evm",
//	).WithKeyTable(v3types.ParamKeyTable())
//	err := v3.MigrateStore(ctx, &paramstore)
//	require.NoError(t, err)
//
//	updatedDefaultConfig := types.DefaultChainConfig()
//
//	var postMigrationConfig types.ChainConfig
//	paramstore.Get(ctx, types.ParamStoreKeyChainConfig, &postMigrationConfig)
//	require.Equal(t, postMigrationConfig.GrayGlacierBlock, updatedDefaultConfig.GrayGlacierBlock)
//	require.Equal(t, postMigrationConfig.MergeNetsplitBlock, updatedDefaultConfig.MergeNetsplitBlock)
//	require.Panics(t, func() {
//		var preMigrationConfig v3types.ChainConfig
//		paramstore.Get(ctx, types.ParamStoreKeyChainConfig, &preMigrationConfig)
//	})
//}

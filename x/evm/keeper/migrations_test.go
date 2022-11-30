package keeper_test

import (
	evmkeeper "github.com/Entangle-Protocol/entangle-blockchain/x/evm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestMigrations() {
	migrator := evmkeeper.NewMigrator(*suite.app.EvmKeeper)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		{
			"Run Migrate1to2",
			migrator.Migrate1to2,
		},
		{
			"Run Migrate2to3",
			migrator.Migrate2to3,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.migrateFunc(suite.ctx)
			suite.Require().NoError(err)
		})
	}
}

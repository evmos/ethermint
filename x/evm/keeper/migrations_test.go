package keeper_test

import (
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
)

func (suite *KeeperTestSuite) TestMigrations() {
	migrator := evmkeeper.NewMigrator(*suite.app.EvmKeeper)

	suite.Run("Run Migrate1to2", func() {
		err := migrator.Migrate1to2(suite.ctx)
		suite.Require().NoError(err)
	})

	suite.Run("Run Migrate2to3", func() {
		err := migrator.Migrate2to3(suite.ctx)
		suite.Require().NoError(err)
	})
}

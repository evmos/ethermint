package keeper_test

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestBeginBlock() {
	// update the counters
	suite.app.EvmKeeper.Bloom.SetInt64(10)
	suite.app.EvmKeeper.TxCount = 10

	suite.app.EvmKeeper.BeginBlock(suite.ctx, abci.RequestBeginBlock{})
	suite.Require().NotZero(suite.app.EvmKeeper.Bloom.Int64())
	suite.Require().NotZero(suite.app.EvmKeeper.TxCount)
}

func (suite *KeeperTestSuite) TestEndBlock() {
	// update the counters
	suite.app.EvmKeeper.Bloom.SetInt64(10)

	// set gas limit to 1 to ensure no gas is consumed during the operation
	initialConsumed := suite.ctx.GasMeter().GasConsumed()

	_ = suite.app.EvmKeeper.EndBlock(suite.ctx, abci.RequestEndBlock{Height: 100})

	suite.Require().Equal(int64(initialConsumed), int64(suite.ctx.GasMeter().GasConsumed()))

	bloom, found := suite.app.EvmKeeper.GetBlockBloom(suite.ctx, 100)
	suite.Require().True(found)
	suite.Require().Equal(int64(10), bloom.Big().Int64())

}

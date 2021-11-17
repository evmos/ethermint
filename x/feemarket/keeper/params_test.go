package keeper_test

import (
	"github.com/tharsis/ethermint/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)
	params.ElasticityMultiplier = 3
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)
}

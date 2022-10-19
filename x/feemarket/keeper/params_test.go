package keeper_test

import (
	"github.com/evmos/ethermint/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestSetGetParams() {
	// Checks if the default params are set correctly
	params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)

	// Check ElasticityMultiplier is set to 3 and can be retrieved correctly
	params.ElasticityMultiplier = 3
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
	newParams := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
	suite.Require().Equal(newParams, params)

	// Check BaseFeeEnabled is computed with its default params and can be retrieved correctly
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
	isBaseFeeEnabled := suite.app.FeeMarketKeeper.GetBaseFeeEnabled(suite.ctx)
	suite.Require().Equal(isBaseFeeEnabled, true)

	// Check BaseFeeEnabled is computed with alternate params and can be retrieved correctly
	params.NoBaseFee = true
	params.EnableHeight = 5
	suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
	isBaseFeeEnabled = suite.app.FeeMarketKeeper.GetBaseFeeEnabled(suite.ctx)
	suite.Require().Equal(isBaseFeeEnabled, false)
}

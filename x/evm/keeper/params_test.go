package keeper_test

import (
	"github.com/evmos/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestParams() {
	// Checks if the default params are set correctly
	params := suite.app.EvmKeeper.GetParams(suite.ctx)
	suite.Require().Equal(types.DefaultParams(), params)

	// Check EvmDenom param is set to "inj" and can be retrieved correctly
	params.EvmDenom = "inj"
	suite.app.EvmKeeper.SetParams(suite.ctx, params)
	evmDenom := suite.app.EvmKeeper.GetEVMDenom(suite.ctx)
	suite.Require().Equal(evmDenom, params.EvmDenom)

	// Check EnableCreate param is set to false and can be retrieved correctly
	params.EnableCreate = false
	suite.app.EvmKeeper.SetParams(suite.ctx, params)
	enableCreate := suite.app.EvmKeeper.GetEnableCreate(suite.ctx)
	suite.Require().Equal(enableCreate, params.EnableCreate)

	// Check EnableCall param is set to false and can be retrieved correctly
	params.EnableCall = false
	suite.app.EvmKeeper.SetParams(suite.ctx, params)
	enableCall := suite.app.EvmKeeper.GetEnableCall(suite.ctx)
	suite.Require().Equal(enableCall, params.EnableCall)

	// Check AllowUnprotectedTxs param is set to false and can be retrieved correctly
	params.AllowUnprotectedTxs = false
	suite.app.EvmKeeper.SetParams(suite.ctx, params)
	allowUnprotectedTxs := suite.app.EvmKeeper.GetAllowUnprotectedTxs(suite.ctx)
	suite.Require().Equal(allowUnprotectedTxs, params.AllowUnprotectedTxs)

	// Check ChainConfig param is set do the DefaultChainConfig and can be retrieved correctly
	params.ChainConfig = types.DefaultChainConfig()
	suite.app.EvmKeeper.SetParams(suite.ctx, params)
	chainConfig := suite.app.EvmKeeper.GetChainConfig(suite.ctx)
	suite.Require().Equal(chainConfig, params.ChainConfig)

}

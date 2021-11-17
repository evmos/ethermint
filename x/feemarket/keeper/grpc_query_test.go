package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

func (suite *KeeperTestSuite) TestQueryBlockGas() {
	ctx := sdk.WrapSDKContext(suite.ctx)

	res, err := suite.queryClient.BlockGas(ctx, &types.QueryBlockGasRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(int64(0), res.Gas)
}

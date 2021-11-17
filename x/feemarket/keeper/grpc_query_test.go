package keeper_test

import (
	"fmt"

	"github.com/tharsis/ethermint/x/feemarket/types"
)

func (suite *KeeperTestSuite) TestQueryParams() {
	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
		exp := &types.QueryParamsResponse{Params: params}

		res, err := suite.queryClient.Params(suite.ctx.Context(), &types.QueryParamsRequest{})
		if tc.expPass {
			suite.Require().Equal(exp, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryBaseFee() {
	// var aux sdk.Int

	testCases := []struct {
		name     string
		malleate func()
		expRes   *types.QueryBaseFeeResponse
		expPass  bool
	}{
		{
			"pass - nil Base Fee",
			func() {},
			&types.QueryBaseFeeResponse{},
			true,
		},
		// TODO non-nil Base Fee
		// {
		// 	"pass - non-nil Base Fee",
		// 	func() {
		// 		baseFee := sdk.OneInt().BigInt()
		// 		suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, baseFee)

		// 		aux = sdk.NewIntFromBigInt(baseFee)
		// 	},
		// 	&types.QueryBaseFeeResponse{BaseFee: &aux},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		fee := suite.app.FeeMarketKeeper.GetBaseFee(suite.ctx)
		fmt.Printf("baseFee: %v", fee)

		res, err := suite.queryClient.BaseFee(suite.ctx.Context(), &types.QueryBaseFeeRequest{})
		if tc.expPass {
			suite.Require().Equal(tc.expRes, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestQueryBlockGas() {
	testCases := []struct {
		name    string
		expPass bool
	}{
		{
			"pass",
			true,
		},
	}
	for _, tc := range testCases {
		gas := suite.app.FeeMarketKeeper.GetBlockGasUsed(suite.ctx)
		exp := &types.QueryBlockGasResponse{Gas: int64(gas)}

		res, err := suite.queryClient.BlockGas(suite.ctx.Context(), &types.QueryBlockGasRequest{})
		if tc.expPass {
			suite.Require().Equal(exp, res, tc.name)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

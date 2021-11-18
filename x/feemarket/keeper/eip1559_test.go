package keeper_test

func (suite *KeeperTestSuite) TestCalculateBaseFee() {
	testCases := []struct {
		name      string
		NoBaseFee bool
		malleate  func()
		expFee    int64
	}{
		{
			"without BaseFee",
			true,
			func() {},
			0,
		},
		{
			"with BaseFee - initial EIP-1559 block",
			false,
			func() {},
			suite.app.FeeMarketKeeper.GetParams(suite.ctx).InitialBaseFee,
		},
		// {
		// 	"with BaseFee - non-initial EIP-1559 block",
		// 	false,
		// 	func() {

		// 		suite.app.Commit()
		// 	},
		// 	suite.app.FeeMarketKeeper.GetParams(suite.ctx).InitialBaseFee,
		// },
	}
	for _, tc := range testCases {
		params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
		params.NoBaseFee = tc.NoBaseFee
		suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

		tc.malleate()

		fee := suite.app.FeeMarketKeeper.CalculateBaseFee(suite.ctx)

		exp := tc.expFee
		if tc.NoBaseFee {
			suite.Require().Nil(fee, tc.name)
		} else {
			suite.Require().Equal(exp, fee, tc.name)
		}
	}
}

package keeper_test

func (suite *KeeperTestSuite) TestCalculateBaseFee() {
	testCases := []struct {
		name      string
		NoBaseFee bool
	}{
		{
			"without BaseFee",
			true,
		},
		{
			"with BaseFee",
			false,
		},
	}
	for _, tc := range testCases {
		params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
		params.NoBaseFee = tc.NoBaseFee
		suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

		exp := params.InitialBaseFee

		fee := suite.app.FeeMarketKeeper.CalculateBaseFee(suite.ctx)
		if tc.NoBaseFee {
			suite.Require().Nil(fee, tc.name)
		} else {
			suite.Require().Equal(exp, fee, tc.name)
		}
	}
}

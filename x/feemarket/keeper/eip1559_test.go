package keeper_test

import (
	"fmt"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestCalculateBaseFee() {
	testCases := []struct {
		name      string
		NoBaseFee bool
		malleate  func()
		expFee    *big.Int
	}{
		{
			"without BaseFee",
			true,
			func() {},
			nil,
		},
		{
			"with BaseFee - initial EIP-1559 block",
			false,
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(0)
			},
			big.NewInt(suite.app.FeeMarketKeeper.GetParams(suite.ctx).InitialBaseFee),
		},
		// TODO: get the maxGas to change the BaseFee?
		{
			"with BaseFee - with gas Limit (consParams)",
			false,
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(1)
				blockParams := abci.BlockParams{
					MaxGas:   1,
					MaxBytes: 1,
				}
				consParams := abci.ConsensusParams{Block: &blockParams}
				suite.ctx = suite.ctx.WithConsensusParams(&consParams)
				fmt.Println(suite.ctx.ConsensusParams())
			},
			big.NewInt(875000000),
		},
		{
			"with BaseFee - non-initial EIP-1559 block",
			false,
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(1)
			},
			big.NewInt(875000000),
		},
		{
			"with BaseFee - non-initial EIP-1559 block",
			false,
			func() {
				suite.ctx = suite.ctx.WithBlockHeight(1)
			},
			big.NewInt(875000000),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
			params.NoBaseFee = tc.NoBaseFee
			suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)

			tc.malleate()

			fee := suite.app.FeeMarketKeeper.CalculateBaseFee(suite.ctx)
			if tc.NoBaseFee {
				suite.Require().Nil(fee, tc.name)
			} else {

				suite.Require().Equal(tc.expFee, fee, tc.name)
			}
		})
	}
}

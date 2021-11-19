package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	testCases := []struct {
		name       string
		malleate   func()
		expGasUsed uint64
	}{
		{
			"basFee nil",
			func() {
				params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
				params.NoBaseFee = true
				suite.app.FeeMarketKeeper.SetParams(suite.ctx, params)
			},
			uint64(0),
		},
		{
			"Block gas meter is nil",
			func() {},
			uint64(0),
		},
		{
			"pass",
			func() {
				meter := sdk.NewGasMeter(uint64(1000000000))
				suite.ctx = suite.ctx.WithBlockGasMeter(meter)
				suite.ctx.BlockGasMeter().ConsumeGas(uint64(5000000), "consume gas")
			},
			uint64(5000000),
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()

			req := abci.RequestEndBlock{Height: 1}
			suite.app.FeeMarketKeeper.EndBlock(suite.ctx, req)
			gasUsed := suite.app.FeeMarketKeeper.GetBlockGasUsed(suite.ctx)
			// TODO TEST LOGGING EVENT
			suite.Require().Equal(tc.expGasUsed, gasUsed, tc.name)
		})
	}
}

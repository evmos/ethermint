package keeper_test

import (
	"math/big"

	"github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestQuerier() {

	testCases := []struct {
		msg      string
		path     []string
		malleate func()
		expPass  bool
	}{
		{"protocol version", []string{types.QueryProtocolVersion}, func() {}, true},
		{"balance", []string{types.QueryBalance, addrHex}, func() {
			suite.app.EvmKeeper.SetBalance(suite.ctx, address, big.NewInt(5))
		}, true},
		// {"balance", []string{types.QueryBalance, "0x01232"}, func() {}, false},
		{"block number", []string{types.QueryBlockNumber, "0x0"}, func() {}, true},
		{"storage", []string{types.QueryStorage, "0x0", "0x0"}, func() {}, true},
		{"code", []string{types.QueryCode, "0x0"}, func() {}, true},
		// {"hash to height", []string{types.QueryHashToHeight, "0x0"}, func() {}, true},
		{"tx logs", []string{types.QueryTransactionLogs, "0x0"}, func() {}, true},
		// {"logs bloom", []string{types.QueryLogsBloom, "0x0"}, func() {}, true},
		{"logs", []string{types.QueryLogs, "0x0"}, func() {}, true},
		{"account", []string{types.QueryAccount, "0x0"}, func() {}, true},
		{"unknown request", []string{"other"}, func() {}, false},
	}

	for i, tc := range testCases {
		suite.Run("", func() {
			//nolint
			tc := tc
			suite.SetupTest() // reset
			//nolint
			tc.malleate()

			bz, err := suite.querier(suite.ctx, tc.path, abci.RequestQuery{})

			//nolint
			if tc.expPass {
				//nolint
				suite.Require().NoError(err, "valid test %d failed: %s", i, tc.msg)
				suite.Require().NotZero(len(bz))
			} else {
				//nolint
				suite.Require().Error(err, "invalid test %d passed: %s", i, tc.msg)
			}
		})
	}
}

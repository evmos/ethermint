package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestCheckGasConsumption() {
	// TODO this should probably be a util or included in evm/types
	// Strip 0x prefix if exists
	addrHex := strings.TrimPrefix(suite.address.String(), "0x")
	addr, err := sdk.AccAddressFromHex(addrHex)
	suite.Require().NoError(err)
	// Get account state
	acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr)
	suite.Require().NotNil(acc)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"Happy Path",
			func() {
				// // Create a contract tx
				// signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
			},
			true,
		},
		{
			"Intrinsic Gas",
			func() {

			},
			true,
		},
		{
			"Inconsistent Gas",
			func() {
				// What should this function do?

			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset
			tc.malleate()
			// ctx := sdk.WrapSDKContext(suite.ctx)
			// err := suite.app.EvmKeeper.CheckGasConsumption()
			if tc.expPass {
				suite.Require().NoError(err)
				// suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
			} else {
				suite.Require().Error(err)
			}
		})

	}

}

// func (suite *KeeperTestSuite) TestApplyTransaction() {
// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{"data",
// 			func() {
// 				// What should this function do?

// 			},
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		suite.Run(tc.name, func() {
// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				// suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})

// 	}

// }

package keeper_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (suite *KeeperTestSuite) TestCreateAccount() {
	testCases := []struct {
		name     string
		addr     common.Address
		malleate func(common.Address)
		callback func(common.Address)
	}{
		{
			"reset account",
			suite.address,
			func(addr common.Address) {
				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(100))
				suite.Require().NotZero(suite.app.EvmKeeper.GetBalance(addr).Int64())
			},
			func(addr common.Address) {
				suite.Require().Zero(suite.app.EvmKeeper.GetBalance(addr).Int64())
			},
		},
		{
			"create account",
			common.HexToAddress("0x49c601A5DC5FA68b19CBbbd0b296eFF9a66805e"),
			func(addr common.Address) {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr.Bytes())
				suite.Require().Nil(acc)
			},
			func(addr common.Address) {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, addr.Bytes())
				suite.Require().NotNil(acc)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate(tc.addr)
			suite.app.EvmKeeper.CreateAccount(tc.addr)
			tc.callback(tc.addr)
		})
	}
}

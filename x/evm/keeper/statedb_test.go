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

func (suite *KeeperTestSuite) TestAddBalance() {
	testCases := []struct {
		name   string
		amount *big.Int
		isNoOp bool
	}{
		{
			"positive amount",
			big.NewInt(100),
			false,
		},
		{
			"zero amount",
			big.NewInt(0),
			true,
		},
		{
			"negative amount",
			big.NewInt(-1),
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			prev := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.app.EvmKeeper.AddBalance(suite.address, tc.amount)
			post := suite.app.EvmKeeper.GetBalance(suite.address)

			if tc.isNoOp {
				suite.Require().Equal(prev.Int64(), post.Int64())
			} else {
				suite.Require().Equal(new(big.Int).Add(prev, tc.amount).Int64(), post.Int64())
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSubBalance() {
	testCases := []struct {
		name     string
		amount   *big.Int
		malleate func()
		isNoOp   bool
	}{
		{
			"positive amount, below zero",
			big.NewInt(100),
			func() {},
			true,
		},
		{
			"positive amount, below zero",
			big.NewInt(50),
			func() {
				suite.app.EvmKeeper.AddBalance(suite.address, big.NewInt(100))
			},
			true,
		},
		{
			"zero amount",
			big.NewInt(0),
			func() {},
			true,
		},
		{
			"negative amount",
			big.NewInt(-1),
			func() {},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			prev := suite.app.EvmKeeper.GetBalance(suite.address)
			suite.app.EvmKeeper.SubBalance(suite.address, tc.amount)
			post := suite.app.EvmKeeper.GetBalance(suite.address)

			if tc.isNoOp {
				suite.Require().Equal(prev.Int64(), post.Int64())
			} else {
				suite.Require().Equal(new(big.Int).Sub(prev, tc.amount).Int64(), post.Int64())
			}
		})
	}
}

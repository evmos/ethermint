package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/tests"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
			tests.GenerateAddress(),
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
			false,
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

func (suite *KeeperTestSuite) TestGetNonce() {
	testCases := []struct {
		name          string
		address       common.Address
		expectedNonce uint64
		malleate      func()
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			0,
			func() {},
		},
		{
			"existing account",
			suite.address,
			1,
			func() {
				suite.app.EvmKeeper.SetNonce(suite.address, 1)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			nonce := suite.app.EvmKeeper.GetNonce(tc.address)
			suite.Require().Equal(tc.expectedNonce, nonce)

		})
	}
}

func (suite *KeeperTestSuite) TestSetNonce() {
	testCases := []struct {
		name     string
		address  common.Address
		nonce    uint64
		malleate func()
	}{
		{
			"new account",
			tests.GenerateAddress(),
			10,
			func() {},
		},
		{
			"existing account",
			suite.address,
			99,
			func() {},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.app.EvmKeeper.SetNonce(tc.address, tc.nonce)
			nonce := suite.app.EvmKeeper.GetNonce(tc.address)
			suite.Require().Equal(tc.nonce, nonce)
		})
	}
}

func (suite *KeeperTestSuite) TestGetCodeHash() {
	addr := tests.GenerateAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name     string
		address  common.Address
		expHash  common.Hash
		malleate func()
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			common.BytesToHash(types.EmptyCodeHash),
			func() {},
		},
		{
			"account not EthAccount type",
			addr,
			common.BytesToHash(types.EmptyCodeHash),
			func() {},
		},
		{
			"existing account",
			suite.address,
			crypto.Keccak256Hash([]byte("codeHash")),
			func() {
				suite.app.EvmKeeper.SetCode(suite.address, []byte("codeHash"))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			hash := suite.app.EvmKeeper.GetCodeHash(tc.address)
			suite.Require().Equal(tc.expHash, hash)
		})
	}
}

func (suite *KeeperTestSuite) TestSetCode() {
	addr := tests.GenerateAddress()
	baseAcc := &authtypes.BaseAccount{Address: sdk.AccAddress(addr.Bytes()).String()}
	suite.app.AccountKeeper.SetAccount(suite.ctx, baseAcc)

	testCases := []struct {
		name    string
		address common.Address
		code    []byte
		isNoOp  bool
	}{
		{
			"account not found",
			tests.GenerateAddress(),
			[]byte("code"),
			false,
		},
		{
			"account not EthAccount type",
			addr,
			nil,
			true,
		},
		{
			"existing account",
			suite.address,
			[]byte("code"),
			false,
		},
		{
			"existing account, code deleted from store",
			suite.address,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			prev := suite.app.EvmKeeper.GetCode(tc.address)
			suite.app.EvmKeeper.SetCode(tc.address, tc.code)
			post := suite.app.EvmKeeper.GetCode(tc.address)

			if tc.isNoOp {
				suite.Require().Equal(prev, post)
			} else {
				suite.Require().Equal(tc.code, post)
			}

			suite.Require().Equal(len(post), suite.app.EvmKeeper.GetCodeSize(tc.address))
		})
	}
}

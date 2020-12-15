package keeper_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	ethermint "github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

func (suite *KeeperTestSuite) TestBalanceInvariant() {
	privkey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	address := ethcmn.HexToAddress(privkey.PubKey().Address().String())

	testCases := []struct {
		name      string
		malleate  func()
		expBroken bool
	}{
		{
			"balance mismatch",
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, address.Bytes())
				suite.Require().NotNil(acc)
				err := acc.SetCoins(sdk.NewCoins(ethermint.NewPhotonCoinInt64(1)))
				suite.Require().NoError(err)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.SetBalance(suite.ctx, address, big.NewInt(1000))
			},
			true,
		},
		{
			"balance ok",
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, address.Bytes())
				suite.Require().NotNil(acc)
				err := acc.SetCoins(sdk.NewCoins(ethermint.NewPhotonCoinInt64(1)))
				suite.Require().NoError(err)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.SetBalance(suite.ctx, address, big.NewInt(1))
			},
			false,
		},
		{
			"invalid account type",
			func() {
				acc := authtypes.NewBaseAccountWithAddress(address.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, &acc)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset values

			tc.malleate()

			_, broken := suite.app.EvmKeeper.BalanceInvariant()(suite.ctx)
			if tc.expBroken {
				suite.Require().True(broken)
			} else {
				suite.Require().False(broken)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestNonceInvariant() {
	privkey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	address := ethcmn.HexToAddress(privkey.PubKey().Address().String())

	testCases := []struct {
		name      string
		malleate  func()
		expBroken bool
	}{
		{
			"nonce mismatch",
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, address.Bytes())
				suite.Require().NotNil(acc)
				err := acc.SetSequence(1)
				suite.Require().NoError(err)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.SetNonce(suite.ctx, address, 100)
			},
			true,
		},
		{
			"nonce ok",
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, address.Bytes())
				suite.Require().NotNil(acc)
				err := acc.SetSequence(1)
				suite.Require().NoError(err)
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.SetNonce(suite.ctx, address, 1)
			},
			false,
		},
		{
			"invalid account type",
			func() {
				acc := authtypes.NewBaseAccountWithAddress(address.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, &acc)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset values

			tc.malleate()

			_, broken := suite.app.EvmKeeper.NonceInvariant()(suite.ctx)
			if tc.expBroken {
				suite.Require().True(broken)
			} else {
				suite.Require().False(broken)
			}
		})
	}
}

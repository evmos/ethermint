package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestCheckGasConsumption() {
	chainID := suite.app.EvmKeeper.ChainID()
	cfg, found := suite.app.EvmKeeper.GetChainConfig(suite.ctx)
	suite.Require().True(found)
	ethCfg := cfg.EthereumConfig(chainID)

	addr, privKey := tests.NewAddrKey()
	signer := tests.NewSigner(privKey)
	ethSigner := ethtypes.LatestSignerForChainID(chainID)

	testCases := []struct {
		name        string
		msg         *types.MsgEthereumTx
		gasConsumed uint64
		expPass     bool
	}{
		{
			"consistent gas",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			53000,
			true,
		},
		{
			"inconsistent gas",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			0,
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			//
			tc.msg.From = addr.Hex()
			err := tc.msg.Sign(ethSigner, signer)
			suite.Require().NoError(err)

			coreMsg, err := tc.msg.AsMessage(ethSigner)
			suite.Require().NoError(err)

			err = suite.app.EvmKeeper.CheckGasConsumption(coreMsg, ethCfg, tc.gasConsumed, tc.msg.To() == nil)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})

	}
}

func (suite *KeeperTestSuite) TestRefundGas() {
	chainID := suite.app.EvmKeeper.ChainID()

	addr, privKey := tests.NewAddrKey()
	signer := tests.NewSigner(privKey)
	ethSigner := ethtypes.LatestSignerForChainID(chainID)

	testCases := []struct {
		name        string
		msg         *types.MsgEthereumTx
		leftoverGas uint64
		malleate    func()
		expPass     bool
	}{
		{
			"leftover gas greater than msg gas",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			200000,
			func() {},
			false,
		},
		{
			"leftover gas greater than msg gas after refund",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			1000,
			func() {
				suite.app.EvmKeeper.AddRefund(200000)
			},
			false,
		},
		{
			"fee collector doesn't have enough funds",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			1000,
			func() {
				suite.app.EvmKeeper.AddRefund(20000)
			},
			false,
		},
		{
			"refund and set gas meter",
			types.NewMsgEthereumTx(chainID, 1, nil, big.NewInt(10), 100000, big.NewInt(1), nil, nil),
			1000,
			func() {
				suite.app.EvmKeeper.AddRefund(20000)

				refund := sdk.Coins{sdk.NewCoin(types.DefaultEVMDenom, sdk.NewInt(50500))}
				// need to mint using the mint module account due to permissions
				err := suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, refund)
				suite.Require().NoError(err)

				err = suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, refund)
				suite.Require().NoError(err)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			//
			tc.msg.From = addr.Hex()
			err := tc.msg.Sign(ethSigner, signer)
			suite.Require().NoError(err)

			coreMsg, err := tc.msg.AsMessage(ethSigner)
			suite.Require().NoError(err)

			tc.malleate()
			err = suite.app.EvmKeeper.RefundGas(coreMsg, tc.leftoverGas)
			if tc.expPass {
				suite.Require().NoError(err)
				gasConsumed := suite.app.EvmKeeper.Context().GasMeter().GasConsumed()
				suite.Require().NotZero(gasConsumed)
				suite.Require().Less(gasConsumed, coreMsg.Gas()-tc.leftoverGas)
			} else {
				suite.Require().Error(err)
			}
		})

	}
}

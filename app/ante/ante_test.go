package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func (suite AnteTestSuite) TestAnteHandler() {
	addr, privKey := newTestAddrKey()
	to, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	err := suite.app.BankKeeper.SetBalance(suite.ctx, addr.Bytes(), sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(10000000000)))
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		txFn      func() sdk.Tx
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 2, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 3, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1)
				return tx
			}, false, true, true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			suite.ctx = suite.ctx.WithIsCheckTx(tc.reCheckTx).WithIsReCheckTx(tc.reCheckTx)

			// expConsumed := params.TxGasContractCreation + params.TxGas
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)

			// suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
				// suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
			} else {
				suite.Require().Error(err)
			}

		})
	}

}

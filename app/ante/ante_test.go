package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/params"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func (suite AnteTestSuite) TestAnteHandler() {
	addr, privKey := newTestAddrKey()
	to, _ := newTestAddrKey()

	signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
	signedContractTx.From = addr.Hex()

	signedTx := evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 2, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil)
	signedTx.From = addr.Hex()

	txContract := suite.CreateTestTx(signedContractTx, privKey, 1)

	suite.txBuilder = suite.clientCtx.TxConfig.NewTxBuilder()
	tx := suite.CreateTestTx(signedTx, privKey, 1)

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	err := suite.app.BankKeeper.SetBalance(suite.ctx, addr.Bytes(), sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(10000000000)))
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		tx        sdk.Tx
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		{"success - DeliverTx (contract)", txContract, false, false, true},
		{"success - CheckTx (contract)", txContract, true, false, true},
		{"success - ReCheckTx (contract)", txContract, false, true, true},
		{"success - DeliverTx", tx, false, false, true},
		{"success - CheckTx", tx, true, false, true},
		{"success - ReCheckTx", tx, false, true, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			suite.ctx = suite.ctx.WithIsCheckTx(tc.reCheckTx).WithIsReCheckTx(tc.reCheckTx)

			expConsumed := params.TxGasContractCreation + params.TxGas
			_, err := suite.anteHandler(suite.ctx, tc.tx, false)

			// suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int(expConsumed), int(suite.ctx.GasMeter().GasConsumed()))
			} else {
				suite.Require().Error(err)
			}

		})
	}

}

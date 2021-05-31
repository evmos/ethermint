package ante_test

import (
	"math/big"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	"github.com/cosmos/ethermint/tests"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func (suite AnteTestSuite) TestAnteHandler() {
	addr, privKey := newTestAddrKey()

	signedContractTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	signedContractTx.From = addr.Hex()
	err := signedContractTx.Sign(suite.app.EvmKeeper.ChainID(), tests.NewSigner(privKey))
	suite.Require().NoError(err)

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	suite.Require().NoError(err)

	builder, ok := suite.txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(signedContractTx)
	suite.Require().NoError(err)

	fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewIntFromBigInt(signedContractTx.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(signedContractTx.GetGas())

	tx, err := suite.CreateTestTx([]cryptotypes.PrivKey{privKey}, []uint64{1}, []uint64{1})
	suite.Require().NoError(err)

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	testCases := []struct {
		name      string
		tx        sdk.Tx
		malleate  func()
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		{"success - DeliverTx (contract)", tx, func() {}, false, false, true},
		{"success - CheckTx (contract)", tx, func() {}, true, false, true},
		{"success - ReCheckTx (contract)", tx, func() {}, false, true, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			suite.ctx = suite.ctx.WithIsCheckTx(tc.reCheckTx).WithIsReCheckTx(tc.reCheckTx)

			tc.malleate()
			// consumed := suite.ctx.GasMeter().GasConsumed()
			_, err := suite.anteHandler(suite.ctx, tc.tx, false)
			// suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}

}

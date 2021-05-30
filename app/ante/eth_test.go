package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app/ante"
	"github.com/cosmos/ethermint/tests"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func (suite AnteTestSuite) TestEthMempoolFeeDecorator() {
	dec := ante.NewEthMempoolFeeDecorator(suite.app.EvmKeeper)

	addr, _ := newTestAddrKey()

	testCases := []struct {
		name    string
		tx      sdk.Tx
		checkTx bool
		expPass bool
	}{
		{"not CheckTx", nil, false, true},
		{"invalid transaction type", nil, true, false},
		{
			"insufficient fees for tx cost",
			evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 1, &addr, big.NewInt(10), 0, big.NewInt(0), nil, nil),
			true,
			false,
		},
		{
			"suficient fees",
			evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 1, &addr, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			true,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(tc.checkTx), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite AnteTestSuite) TestEthSigVerificationDecorator() {
	dec := ante.NewEthSigVerificationDecorator(suite.app.EvmKeeper)
	addr, privKey := newTestAddrKey()

	signedTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	signedTx.From = addr.Hex()
	err := signedTx.Sign(suite.app.EvmKeeper.ChainID(), tests.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		tx        sdk.Tx
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", nil, true, true},
		{"invalid transaction type", nil, false, false},
		{
			"invalid sender",
			evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 1, &addr, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			false,
			false,
		},
		{"successful signature verification", signedTx, false, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TestNewEthAccountVerificationDecorator() {
	dec := ante.NewEthAccountVerificationDecorator(
		suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper,
	)

	addr, _ := newTestAddrKey()

	tx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	tx.From = addr.Hex()

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		checkTx  bool
		expPass  bool
	}{
		{"not CheckTx", nil, func() {}, false, true},
		{"invalid transaction type", nil, func() {}, true, false},
		{
			"sender not set to msg",
			evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			func() {},
			true,
			false,
		},
		{
			"not enough balance to cover tx cost",
			tx,
			func() {},
			true,
			false,
		},
		{
			"success new account",
			tx,
			func() {
				err := suite.app.BankKeeper.SetBalance(suite.ctx, addr.Bytes(), sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(1000000)))
				suite.Require().NoError(err)
			},
			true,
			true,
		},
		{
			"success existing account",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				err := suite.app.BankKeeper.SetBalance(suite.ctx, addr.Bytes(), sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(1000000)))
				suite.Require().NoError(err)
			},
			true,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.malleate()

			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(tc.checkTx), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TestEthNonceVerificationDecorator() {
	dec := ante.NewEthNonceVerificationDecorator(suite.app.AccountKeeper)

	addr, _ := newTestAddrKey()

	tx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	tx.From = addr.Hex()

	testCases := []struct {
		name      string
		tx        sdk.Tx
		malleate  func()
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", nil, func() {}, true, true},
		{"invalid transaction type", nil, func() {}, false, false},
		{"sender account not found", tx, func() {}, false, false},
		{
			"sender nonce missmatch",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.Require().NoError(acc.SetSequence(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()
			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TestEthGasConsumeDecorator() {
	dec := ante.NewEthGasConsumeDecorator(
		suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper,
	)
	testCases := []struct {
		name    string
		tx      sdk.Tx
		expPass bool
	}{}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := dec.AnteHandle(suite.ctx, tc.tx, false, nextFn)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				// TODO: check gas meter consumption remains 0
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TestEthIncrementSenderSequenceDecorator() {
	dec := ante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)
	addr, privKey := newTestAddrKey()

	signedTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	signedTx.From = addr.Hex()
	err := signedTx.Sign(suite.app.EvmKeeper.ChainID(), tests.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
		expPanic bool
	}{
		{"invalid transaction type", nil, func() {}, false, false},
		{
			"no signers",
			evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			func() {},
			false, true,
		},
		{
			"account not set to store",
			signedTx,
			func() {},
			false, false,
		},
		{
			"success",
			signedTx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			true, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()
			consumed := suite.ctx.GasMeter().GasConsumed()

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(suite.ctx, tc.tx, false, nextFn)
				})
				return
			}

			ctx, err := dec.AnteHandle(suite.ctx, tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

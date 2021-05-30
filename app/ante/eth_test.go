package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app/ante"
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
	addr, _ := newTestAddrKey()

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
			true,
		},
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

	testCases := []struct {
		name    string
		tx      sdk.Tx
		checkTx bool
		expPass bool
	}{
		{"not CheckTx", nil, false, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(tc.checkTx), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				// TODO: check gas meter consumption remains 0
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TesEthNonceVerificationDecorator() {
	dec := ante.NewEthNonceVerificationDecorator(suite.app.AccountKeeper)

	testCases := []struct {
		name      string
		tx        sdk.Tx
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", nil, true, true},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := dec.AnteHandle(suite.ctx.WithIsReCheckTx(tc.reCheckTx), tc.tx, false, nextFn)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				// TODO: check gas meter consumption remains 0
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TesEthGasConsumeDecorator() {
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

func (suite AnteTestSuite) TesEthIncrementSenderSequenceDecorator() {
	dec := ante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)

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

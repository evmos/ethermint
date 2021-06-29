package ante_test

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/app/ante"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func nextFn(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
	return ctx, nil
}

func (suite AnteTestSuite) TestEthSigVerificationDecorator() {
	dec := ante.NewEthSigVerificationDecorator(suite.app.EvmKeeper)
	addr, privKey := newTestAddrKey()

	signedTx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	signedTx.From = addr.Hex()
	err := signedTx.Sign(suite.ethSigner, tests.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name      string
		tx        sdk.Tx
		reCheckTx bool
		expPass   bool
	}{
		{"ReCheckTx", nil, true, false},
		{"invalid transaction type", &invalidTx{}, false, false},
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
		{"invalid transaction type", &invalidTx{}, func() {}, true, false},
		{
			"sender not set to msg",
			evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			func() {},
			true,
			false,
		},
		{
			"sender not EOA",
			tx,
			func() {
				// set not as an EOA
				suite.app.EvmKeeper.SetCode(addr, []byte("1"))
			},
			true,
			false,
		},
		{
			"not enough balance to cover tx cost",
			tx,
			func() {
				// reset back to EOA
				suite.app.EvmKeeper.SetCode(addr, nil)
			},
			true,
			false,
		},
		{
			"success new account",
			tx,
			func() {
				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))
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

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))

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
		{"invalid transaction type", &invalidTx{}, func() {}, false, false},
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

	addr, _ := newTestAddrKey()

	tx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	tx.From = addr.Hex()

	tx2 := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000000, big.NewInt(1), nil, &ethtypes.AccessList{{Address: addr, StorageKeys: nil}})
	tx2.From = addr.Hex()

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
		expPanic bool
	}{
		{"invalid transaction type", &invalidTx{}, func() {}, false, false},
		{
			"sender not found",
			evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			func() {},
			false, false,
		},
		{
			"gas limit too low",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false, false,
		},
		{
			"not enough balance for fees",
			tx2,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false, false,
		},
		{
			"not enough tx gas",
			tx2,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))
			},
			false, true,
		},
		{
			"not enough block gas",
			tx2,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))

				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1))
			},
			false, true,
		},
		{
			"success",
			tx2,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))

				suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(10000000000000000000))
			},
			true, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			if tc.expPanic {
				suite.Require().Panics(func() {
					_, _ = dec.AnteHandle(suite.ctx.WithIsCheckTx(true).WithGasMeter(sdk.NewGasMeter(1)), tc.tx, false, nextFn)
				})
				return
			}

			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true), tc.tx, false, nextFn)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(int(params.TxGasContractCreation+params.TxAccessListAddressGas), int(ctx.GasMeter().GasConsumed()-consumed))
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite AnteTestSuite) TestCanTransferDecorator() {
	dec := ante.NewCanTransferDecorator(suite.app.EvmKeeper)

	addr, privKey := newTestAddrKey()

	tx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, &ethtypes.AccessList{})
	tx2 := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, &ethtypes.AccessList{})

	tx.From = addr.Hex()

	err := tx.Sign(suite.ethSigner, tests.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
	}{
		{"invalid transaction type", &invalidTx{}, func() {}, false},
		{"AsMessage failed", tx2, func() {}, false},
		{
			"evm CanTransfer failed",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			false,
		},
		{
			"success",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite AnteTestSuite) TestAccessListDecorator() {
	dec := ante.NewAccessListDecorator(suite.app.EvmKeeper)

	addr, _ := newTestAddrKey()
	al := &ethtypes.AccessList{
		{Address: addr, StorageKeys: []common.Hash{{}}},
	}

	tx := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	tx2 := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 1, big.NewInt(10), 1000, big.NewInt(1), nil, al)

	tx.From = addr.Hex()
	tx2.From = addr.Hex()

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
	}{
		{"invalid transaction type", &invalidTx{}, func() {}, false},
		{
			"success - no access list",
			tx,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))
			},
			true,
		},
		{
			"success - with access list",
			tx2,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				suite.app.EvmKeeper.AddBalance(addr, big.NewInt(1000000))
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			tc.malleate()

			consumed := suite.ctx.GasMeter().GasConsumed()
			ctx, err := dec.AnteHandle(suite.ctx.WithIsCheckTx(true), tc.tx, false, nextFn)
			suite.Require().Equal(consumed, ctx.GasMeter().GasConsumed())

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite AnteTestSuite) TestEthIncrementSenderSequenceDecorator() {
	dec := ante.NewEthIncrementSenderSequenceDecorator(suite.app.AccountKeeper)
	addr, privKey := newTestAddrKey()

	contract := evmtypes.NewMsgEthereumTxContract(suite.app.EvmKeeper.ChainID(), 0, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	contract.From = addr.Hex()

	to := tests.GenerateAddress()
	tx := evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 0, &to, big.NewInt(10), 1000, big.NewInt(1), nil, nil)
	tx.From = addr.Hex()

	err := contract.Sign(suite.ethSigner, tests.NewSigner(privKey))
	suite.Require().NoError(err)

	err = tx.Sign(suite.ethSigner, tests.NewSigner(privKey))
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		tx       sdk.Tx
		malleate func()
		expPass  bool
		expPanic bool
	}{
		{
			"invalid transaction type",
			&invalidTx{},
			func() {},
			false, false,
		},
		{
			"no signers",
			evmtypes.NewMsgEthereumTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 1000, big.NewInt(1), nil, nil),
			func() {},
			false, true,
		},
		{
			"account not set to store",
			tx,
			func() {},
			false, false,
		},
		{
			"success - create contract",
			contract,
			func() {
				acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
			},
			true, false,
		},
		{
			"success - call",
			tx,
			func() {},
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
				msg := tc.tx.(*evmtypes.MsgEthereumTx)
				nonce := suite.app.EvmKeeper.GetNonce(addr)
				if msg.To() == nil {
					suite.Require().Equal(msg.Data.Nonce, nonce)
				} else {
					suite.Require().Equal(msg.Data.Nonce+1, nonce)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

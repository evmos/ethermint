package ante_test

import (
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/tharsis/ethermint/tests"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

func (suite AnteTestSuite) TestAnteHandler() {
	suite.dynamicTxFee = false
	suite.SetupTest() // reset

	addr, privKey := tests.NewAddrKey()
	to := tests.GenerateAddress()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	suite.app.EvmKeeper.AddBalance(addr, big.NewInt(10000000000))

	suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))

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
				signedContractTx := evmtypes.NewTxContract(
					suite.app.EvmKeeper.ChainID(),
					1,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
					suite.app.EvmKeeper.ChainID(),
					2,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
					suite.app.EvmKeeper.ChainID(),
					3,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					4,
					&to,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					5,
					&to,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					3,
					&to,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					4,
					&to,
					big.NewInt(10),
					100000,
					big.NewInt(150),
					big.NewInt(200),
					nil,
					nil,
					nil,
				)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 4, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 5, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (ExtensionOptionsEthereumTx not set)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 5, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false, true)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		// Based on EVMBackend.SendTransaction, for cosmos tx, forcing null for some fields except ExtensionOptions, Fee, MsgEthereumTx
		// should be part of consensus
		{
			"fail - DeliverTx (cosmos tx signed)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, true)
				return tx
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with memo)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo("memo for cosmos tx not allowed")
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (cosmos tx with timeoutheight)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetTimeoutHeight(10)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee amount)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				txData, err := evmtypes.UnpackTxData(signedTx.Data)
				suite.Require().NoError(err)

				expFee := txData.Fee()
				invalidFee := new(big.Int).Add(expFee, big.NewInt(1))
				invalidFeeAmount := sdk.Coins{sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewIntFromBigInt(invalidFee))}
				txBuilder.SetFeeAmount(invalidFeeAmount)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fail - DeliverTx (invalid fee gaslimit)",
			func() sdk.Tx {
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), nonce, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)

				expGasLimit := signedTx.GetGas()
				invalidGasLimit := expGasLimit + 1
				txBuilder.SetGasLimit(invalidGasLimit)
				return txBuilder.GetTx()
			}, false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)

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

func (suite AnteTestSuite) TestAnteHandlerWithDynamicTxFee() {
	suite.dynamicTxFee = true
	suite.SetupTest() // reset

	addr, privKey := tests.NewAddrKey()
	to := tests.GenerateAddress()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
	suite.Require().NoError(acc.SetSequence(1))
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

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
				signedContractTx :=
					evmtypes.NewTxContract(
						suite.app.EvmKeeper.ChainID(),
						1,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx :=
					evmtypes.NewTxContract(
						suite.app.EvmKeeper.ChainID(),
						2,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx :=
					evmtypes.NewTxContract(
						suite.app.EvmKeeper.ChainID(),
						3,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedContractTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedContractTx, privKey, 1, false)
				return tx
			},
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						4,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						5,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			},
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						3,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						4,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				tx := suite.CreateTestTx(signedTx, privKey, 1, false)
				return tx
			}, false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						4,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				// bigger than MaxGasWanted
				txBuilder.SetGasLimit(uint64(1 << 63))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx :=
					evmtypes.NewTx(
						suite.app.EvmKeeper.ChainID(),
						5,
						&to,
						big.NewInt(10),
						100000,
						nil,
						big.NewInt(ethparams.InitialBaseFee+1),
						big.NewInt(1),
						nil,
						&types.AccessList{},
					)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			suite.app.EvmKeeper.AddBalance(addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.dynamicTxFee = false
}

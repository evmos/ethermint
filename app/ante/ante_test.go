package ante_test

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/core/types"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/evmos/ethermint/app/ante"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite AnteTestSuite) TestAnteHandler() {
	var acc authtypes.AccountI
	addr, privKey := tests.NewAddrKey()
	to := tests.GenerateAddress()

	setup := func() {
		suite.enableFeemarket = false
		suite.SetupTest() // reset

		acc = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt(10000000000))

		suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
	}

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
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
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
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
					1,
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
					1,
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
					1,
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
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
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
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
				signedTx.From = addr.Hex()

				txBuilder := suite.CreateTestTxBuilder(signedTx, privKey, 1, false)
				txBuilder.SetMemo(strings.Repeat("*", 257))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fail - CheckTx (ExtensionOptionsEthereumTx not set)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil)
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
				invalidFeeAmount := sdk.Coins{sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewIntFromBigInt(invalidFee))}
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
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with DelegateMsg",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas)))
				amount := sdk.NewCoins(coinAmount)
				txBuilder := suite.CreateTestEIP712TxBuilderMsgDelegate(from, privKey, "ethermint_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgCreateValidator(from, privKey, "ethermint_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgSubmitProposal",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				// reusing the gasAmount for deposit
				deposit := sdk.NewCoins(coinAmount)
				txBuilder := suite.CreateTestEIP712SubmitProposal(from, privKey, "ethermint_9000-1", gas, gasAmount, deposit)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrant",
			func() sdk.Tx {
				from := acc.GetAddress()
				grantee := sdk.AccAddress("_______grantee______")
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				blockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
				expiresAt := blockTime.Add(time.Hour)
				msg, err := authz.NewMsgGrant(
					from, grantee, &banktypes.SendAuthorization{SpendLimit: gasAmount}, &expiresAt,
				)
				suite.Require().NoError(err)
				return suite.CreateTestEIP712CosmosTxBuilder(from, privKey, "ethermint_9000-1", gas, gasAmount, msg).GetTx()
			}, false, false, true,
		},

		{
			"success- DeliverTx EIP712 MsgGrantAllowance",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712GrantAllowance(from, privKey, "ethermint_9000-1", gas, gasAmount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 edit validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgEditValidator(from, privKey, "ethermint_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit evidence",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgEditValidator(from, privKey, "ethermint_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9002-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9001-1", gas, amount)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with empty signature",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9001-1", gas, amount)
				sigsV2 := signing.SignatureV2{}
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9001-1", gas, amount)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					},
					Sequence: nonce - 1,
				}
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid signMode",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9001-1", gas, amount)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_UNSPECIFIED,
					},
					Sequence: nonce,
				}
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - invalid from",
			func() sdk.Tx {
				msg := evmtypes.NewTxContract(
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
				msg.From = addr.Hex()
				tx := suite.CreateTestTx(msg, privKey, 1, false)
				msg = tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
				msg.From = addr.Hex()
				return tx
			}, true, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

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
	addr, privKey := tests.NewAddrKey()
	to := tests.GenerateAddress()

	testCases := []struct {
		name           string
		txFn           func() sdk.Tx
		enableLondonHF bool
		checkTx        bool
		reCheckTx      bool
		expPass        bool
	}{
		{
			"success - DeliverTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			true,
			false, false, true,
		},
		{
			"success - CheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx (contract)",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			true,
			false, true, true,
		},
		{
			"success - DeliverTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			true,
			false, false, true,
		},
		{
			"success - CheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			true,
			true, false, true,
		},
		{
			"success - ReCheckTx",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			true,
			false, true, true,
		},
		{
			"success - CheckTx (cosmos tx not signed)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			true,
			false, true, true,
		},
		{
			"fail - CheckTx (cosmos tx is not valid)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			},
			true,
			true, false, false,
		},
		{
			"fail - CheckTx (memo too long)",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			},
			true,
			true, false, false,
		},
		{
			"fail - DynamicFeeTx without london hark fork",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			false,
			false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.enableFeemarket = true
			suite.enableLondonHF = tc.enableLondonHF
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)
			suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.enableFeemarket = false
	suite.enableLondonHF = true
}

func (suite AnteTestSuite) TestAnteHandlerWithParams() {
	addr, privKey := tests.NewAddrKey()
	to := tests.GenerateAddress()

	testCases := []struct {
		name         string
		txFn         func() sdk.Tx
		enableCall   bool
		enableCreate bool
		expErr       error
	}{
		{
			"fail - Contract Creation Disabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			true, false,
			evmtypes.ErrCreateDisabled,
		},
		{
			"success - Contract Creation Enabled",
			func() sdk.Tx {
				signedContractTx := evmtypes.NewTxContract(
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
			true, true,
			nil,
		},
		{
			"fail - EVM Call Disabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			false, true,
			evmtypes.ErrCallDisabled,
		},
		{
			"success - EVM Call Enabled",
			func() sdk.Tx {
				signedTx := evmtypes.NewTx(
					suite.app.EvmKeeper.ChainID(),
					1,
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
			true, true,
			nil,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.evmParamsOption = func(params *evmtypes.Params) {
				params.EnableCall = tc.enableCall
				params.EnableCreate = tc.enableCreate
			}
			suite.SetupTest() // reset

			acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
			suite.Require().NoError(acc.SetSequence(1))
			suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

			suite.ctx = suite.ctx.WithIsCheckTx(true)
			suite.app.EvmKeeper.SetBalance(suite.ctx, addr, big.NewInt((ethparams.InitialBaseFee+10)*100000))
			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)
			if tc.expErr == nil {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
				suite.Require().True(errors.Is(err, tc.expErr))
			}
		})
	}
	suite.evmParamsOption = nil
}

func (suite *AnteTestSuite) TestConsumeSignatureVerificationGas() {
	params := authtypes.DefaultParams()
	msg := []byte{1, 2, 3, 4}
	cdc := simapp.MakeTestEncodingConfig().Amino

	p := authtypes.DefaultParams()
	skR1, _ := secp256r1.GenPrivKey()
	pkSet1, sigSet1, err := generatePubKeysAndSignatures(5, msg, false)
	suite.Require().NoError(err)

	multisigKey1 := kmultisig.NewLegacyAminoPubKey(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	
	for i := 0; i < len(pkSet1); i++ {
		stdSig := legacytx.StdSignature{PubKey: pkSet1[i], Signature: sigSet1[i]}
		sigV2, err := legacytx.StdSignatureToSignatureV2(cdc, stdSig)
		suite.Require().NoError(err)
		err = multisig.AddSignatureV2(multisignature1, sigV2, pkSet1)
		suite.Require().NoError(err)
	}

	type args struct {
		meter  sdk.GasMeter
		sig    signing.SignatureData
		pubkey cryptotypes.PubKey
		params authtypes.Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed uint64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{sdk.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, p.SigVerifyCostED25519, true},
		{"PubKeyEthSecp256k1", args{sdk.NewInfiniteGasMeter(), nil, pkSet1[0], params}, 21_000, false},
		{"PubKeySecp256r1", args{sdk.NewInfiniteGasMeter(), nil, skR1.PubKey(), params}, p.SigVerifyCostSecp256r1(), false},
		{"Multisig", args{sdk.NewInfiniteGasMeter(), multisignature1, multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{sdk.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		sigV2 := signing.SignatureV2{
			PubKey:   tt.args.pubkey,
			Data:     tt.args.sig,
			Sequence: 0, // Arbitrary account sequence
		}
		err := ante.DefaultSigVerificationGasConsumer(tt.args.meter, sigV2, tt.args.params)

		if tt.shouldErr {
			suite.Require().NotNil(err)
		} else {
			suite.Require().Nil(err)
			suite.Require().Equal(tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
		}
	}
}

func generatePubKeysAndSignatures(n int, msg []byte, _ bool) (pubkeys []cryptotypes.PubKey, signatures [][]byte, err error) {
	pubkeys = make([]cryptotypes.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		privkey, err := ethsecp256k1.GenerateKey()
		if err != nil {
			return nil, nil, err
		}

		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

func expectedGasCostByKeys(pubkeys []cryptotypes.PubKey) uint64 {
	cost := uint64(0)
	for _, pubkey := range pubkeys {
		pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
		switch {
		case strings.Contains(pubkeyType, "ed25519"):
			cost += authtypes.DefaultParams().SigVerifyCostED25519
		case strings.Contains(pubkeyType, "ethsecp256k1"):
			cost += 21_000
		case strings.Contains(pubkeyType, "secp256k1"):
			cost += authtypes.DefaultParams().SigVerifyCostSecp256k1
		default:
			panic("unexpected key type")
		}
	}
	return cost
}

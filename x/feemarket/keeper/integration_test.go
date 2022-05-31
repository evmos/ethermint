package keeper_test

import (
	"encoding/json"
	"math/big"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/app"
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/testutil"
	"github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

var _ = Describe("Ethermint App min gas prices settings: ", func() {
	var (
		privKey *ethsecp256k1.PrivKey
		address sdk.AccAddress
		msg     banktypes.MsgSend
	)

	var setupChain = func(cliMinGasPricesStr string) {
		// Initialize the app, so we can use SetMinGasPrices to set the
		// validator-specific min-gas-prices setting
		db := dbm.NewMemDB()
		newapp := app.NewEthermintApp(
			log.NewNopLogger(),
			db,
			nil,
			true,
			map[int64]bool{},
			app.DefaultNodeHome,
			5,
			encoding.MakeConfig(app.ModuleBasics),
			simapp.EmptyAppOptions{},
			baseapp.SetMinGasPrices(cliMinGasPricesStr),
		)

		genesisState := app.NewDefaultGenesisState()
		genesisState[types.ModuleName] = newapp.AppCodec().MustMarshalJSON(types.DefaultGenesisState())

		stateBytes, err := json.MarshalIndent(genesisState, "", "  ")
		s.Require().NoError(err)

		// Initialize the chain
		newapp.InitChain(
			abci.RequestInitChain{
				ChainId:         "ethermint_9000-1",
				Validators:      []abci.ValidatorUpdate{},
				AppStateBytes:   stateBytes,
				ConsensusParams: simapp.DefaultConsensusParams,
			},
		)

		s.app = newapp
		s.SetupApp(false)
	}

	var setupTest = func(cliMinGasPrices string) {
		setupChain(cliMinGasPrices)

		privKey, address = generateKey()
		amount, ok := sdk.NewIntFromString("10000000000000000000")
		s.Require().True(ok)
		initBalance := sdk.Coins{sdk.Coin{
			Denom:  s.denom,
			Amount: amount,
		}}
		testutil.FundAccount(s.app.BankKeeper, s.ctx, address, initBalance)

		msg = banktypes.MsgSend{
			FromAddress: address.String(),
			ToAddress:   address.String(),
			Amount: sdk.Coins{sdk.Coin{
				Denom:  s.denom,
				Amount: sdk.NewInt(10000),
			}},
		}
		s.Commit()
	}

	var setupContext = func(cliMinGasPrice string, minGasPrice sdk.Dec) {
		setupTest(cliMinGasPrice + s.denom)
		params := types.DefaultParams()
		params.MinGasPrice = minGasPrice
		s.app.FeeMarketKeeper.SetParams(s.ctx, params)
		s.Commit()
	}

	Context("with Cosmos transactions", func() {
		Context("min-gas-prices (local) < MinGasPrices (feemarket param)", func() {
			cliMinGasPrice := "1"
			minGasPrice := sdk.NewDecWithPrec(3, 0)
			BeforeEach(func() {
				setupContext(cliMinGasPrice, minGasPrice)
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdk.NewInt(2)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdk.NewInt(3)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdk.NewInt(2)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdk.NewInt(3)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with min-gas-prices (local) == MinGasPrices (feemarket param)", func() {
			cliMinGasPrice := "3"
			minGasPrice := sdk.NewDecWithPrec(3, 0)
			BeforeEach(func() {
				setupContext(cliMinGasPrice, minGasPrice)
			})

			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < min-gas-prices", func() {
					gasPrice := sdk.NewInt(2)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"insufficient fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdk.NewInt(3)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdk.NewInt(2)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with gasPrice >= MinGasPrices", func() {
					gasPrice := sdk.NewInt(3)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})

		Context("with MinGasPrices (feemarket param) < min-gas-prices (local)", func() {
			cliMinGasPrice := "5"
			minGasPrice := sdk.NewDecWithPrec(3, 0)
			BeforeEach(func() {
				setupContext(cliMinGasPrice, minGasPrice)
			})
			Context("during CheckTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdk.NewInt(2)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"insufficient fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should reject transactions with MinGasPrices < gasPrice < min-gas-prices", func() {
					gasPrice := sdk.NewInt(4)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"insufficient fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with gasPrice > min-gas-prices", func() {
					gasPrice := sdk.NewInt(5)
					res := checkTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})

			Context("during DeliverTx", func() {
				It("should reject transactions with gasPrice < MinGasPrices", func() {
					gasPrice := sdk.NewInt(2)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
					Expect(
						strings.Contains(res.GetLog(),
							"provided fee < minimum global fee"),
					).To(BeTrue(), res.GetLog())
				})

				It("should accept transactions with MinGasPrices < gasPrice < than min-gas-prices", func() {
					gasPrice := sdk.NewInt(4)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})

				It("should accept transactions with gasPrice >= min-gas-prices", func() {
					gasPrice := sdk.NewInt(5)
					res := deliverTx(privKey, &gasPrice, &msg)
					Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
				})
			})
		})
	})

	Context("with EVM transactions", func() {
		type txParams struct {
			gasPrice  *big.Int
			gasFeeCap *big.Int
			gasTipCap *big.Int
			accesses  *ethtypes.AccessList
		}
		type getprices func() txParams

		getBaseFee := func() int64 {
			paramsEvm := s.app.EvmKeeper.GetParams(s.ctx)
			ethCfg := paramsEvm.ChainConfig.EthereumConfig(s.app.EvmKeeper.ChainID())
			return s.app.EvmKeeper.GetBaseFee(s.ctx, ethCfg).Int64()
		}
		Context("with BaseFee (feemarket) < MinGasPrices (feemarket param)", func() {
			var baseFee int64
			BeforeEach(func() {
				baseFee = getBaseFee()
				setupContext("1", sdk.NewDecWithPrec(baseFee+30000000000, 0))
			})

			Context("during CheckTx", func() {
				DescribeTable("should reject transactions with EffectivePrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := checkEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 20000000000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 20000000000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 29000000000), big.NewInt(29000000000), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice < MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 40000000000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := checkEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 31000000000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 31000000000), big.NewInt(31000000000), &ethtypes.AccessList{}}
					}),
				)
			})

			Context("during DeliverTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := deliverEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 20000000000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 20000000000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap < MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 29000000000), big.NewInt(29000000000), &ethtypes.AccessList{}}
					}),
					Entry("dynamic tx with GasFeeCap > MinGasPrices, EffectivePrice < MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 40000000000), big.NewInt(0), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice >= MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := deliverEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 30000000001), nil, nil, nil}
					}),
					// the base fee decreases in this test, so we use a large gas tip
					// to maintain an EffectivePrice > MinGasPrices
					Entry("dynamic tx, EffectivePrice > MinGasPrices", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 30000000001), big.NewInt(30000000001), &ethtypes.AccessList{}}
					}),
				)
			})
		})

		Context("with MinGasPrices (feemarket param) < BaseFee (feemarket)", func() {
			var baseFee int64
			BeforeEach(func() {
				baseFee = getBaseFee()
				s.Require().Greater(baseFee, int64(10))
				setupContext("5", sdk.NewDecWithPrec(10, 0))
			})

			Context("during CheckTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := checkEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(2), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(2), big.NewInt(2), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should reject transactions with MinGasPrices < tx gasPrice < EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := checkEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"insufficient fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(20), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee - 1), big.NewInt(20), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice > EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := checkEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 1000000000), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 1000000000), big.NewInt(10), &ethtypes.AccessList{}}
					}),
				)
			})

			Context("during DeliverTx", func() {
				DescribeTable("should reject transactions with gasPrice < MinGasPrices",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := deliverEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"provided fee < minimum global fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(2), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(2), big.NewInt(2), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should reject transactions with MinGasPrices < gasPrice < EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := deliverEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(false), "transaction should have failed")
						Expect(
							strings.Contains(res.GetLog(),
								"insufficient fee"),
						).To(BeTrue(), res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(20), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(20), big.NewInt(20), &ethtypes.AccessList{}}
					}),
				)

				DescribeTable("should accept transactions with gasPrice > EffectivePrice",
					func(malleate getprices) {
						p := malleate()
						to := tests.GenerateAddress()
						msgEthereumTx := buildEthTx(privKey, &to, p.gasPrice, p.gasFeeCap, p.gasTipCap, p.accesses)
						res := deliverEthTx(privKey, msgEthereumTx)
						Expect(res.IsOK()).To(Equal(true), "transaction should have succeeded", res.GetLog())
					},
					Entry("legacy tx", func() txParams {
						return txParams{big.NewInt(baseFee + 10), nil, nil, nil}
					}),
					Entry("dynamic tx", func() txParams {
						return txParams{nil, big.NewInt(baseFee + 10), big.NewInt(10), &ethtypes.AccessList{}}
					}),
				)
			})
		})
	})
})

func generateKey() (*ethsecp256k1.PrivKey, sdk.AccAddress) {
	address, priv := tests.NewAddrKey()
	return priv.(*ethsecp256k1.PrivKey), sdk.AccAddress(address.Bytes())
}

func getNonce(addressBytes []byte) uint64 {
	return s.app.EvmKeeper.GetNonce(
		s.ctx,
		common.BytesToAddress(addressBytes),
	)
}

func buildEthTx(
	priv *ethsecp256k1.PrivKey,
	to *common.Address,
	gasPrice *big.Int,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	accesses *ethtypes.AccessList,
) *evmtypes.MsgEthereumTx {
	chainID := s.app.EvmKeeper.ChainID()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	nonce := getNonce(from.Bytes())
	data := make([]byte, 0)
	gasLimit := uint64(100000)
	msgEthereumTx := evmtypes.NewTx(
		chainID,
		nonce,
		to,
		nil,
		gasLimit,
		gasPrice,
		gasFeeCap,
		gasTipCap,
		data,
		accesses,
	)
	msgEthereumTx.From = from.String()
	return msgEthereumTx
}

func prepareEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	s.Require().NoError(err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	s.Require().True(ok)
	builder.SetExtensionOptions(option)

	err = msgEthereumTx.Sign(s.ethSigner, tests.NewSigner(priv))
	s.Require().NoError(err)

	err = txBuilder.SetMsgs(msgEthereumTx)
	s.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msgEthereumTx.Data)
	s.Require().NoError(err)

	evmDenom := s.app.EvmKeeper.GetParams(s.ctx).EvmDenom
	fees := sdk.Coins{{Denom: evmDenom, Amount: sdk.NewIntFromBigInt(txData.Fee())}}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msgEthereumTx.GetGas())

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	return bz
}

func deliverEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseDeliverTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func checkEthTx(priv *ethsecp256k1.PrivKey, msgEthereumTx *evmtypes.MsgEthereumTx) abci.ResponseCheckTx {
	bz := prepareEthTx(priv, msgEthereumTx)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.BaseApp.CheckTx(req)
	return res
}

func prepareCosmosTx(priv *ethsecp256k1.PrivKey, gasPrice *sdk.Int, msgs ...sdk.Msg) []byte {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	txBuilder.SetGasLimit(1000000)
	if gasPrice == nil {
		_gasPrice := sdk.NewInt(1)
		gasPrice = &_gasPrice
	}
	fees := &sdk.Coins{{Denom: s.denom, Amount: gasPrice.MulRaw(1000000)}}
	txBuilder.SetFeeAmount(*fees)
	err := txBuilder.SetMsgs(msgs...)
	s.Require().NoError(err)

	seq, err := s.app.AccountKeeper.GetSequence(s.ctx, accountAddress)
	s.Require().NoError(err)

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encodingConfig.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := s.app.AccountKeeper.GetAccount(s.ctx, accountAddress).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       s.ctx.ChainID(),
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		encodingConfig.TxConfig.SignModeHandler().DefaultMode(), signerData,
		txBuilder, priv, encodingConfig.TxConfig,
		seq,
	)
	s.Require().NoError(err)

	sigsV2 = []signing.SignatureV2{sigV2}
	err = txBuilder.SetSignatures(sigsV2...)
	s.Require().NoError(err)

	// bz are bytes to be broadcasted over the network
	bz, err := encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	return bz
}

func deliverTx(priv *ethsecp256k1.PrivKey, gasPrice *sdk.Int, msgs ...sdk.Msg) abci.ResponseDeliverTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestDeliverTx{Tx: bz}
	res := s.app.BaseApp.DeliverTx(req)
	return res
}

func checkTx(priv *ethsecp256k1.PrivKey, gasPrice *sdk.Int, msgs ...sdk.Msg) abci.ResponseCheckTx {
	bz := prepareCosmosTx(priv, gasPrice, msgs...)
	req := abci.RequestCheckTx{Tx: bz}
	res := s.app.BaseApp.CheckTx(req)
	return res
}

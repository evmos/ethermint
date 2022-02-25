package ante_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/tharsis/ethermint/ethereum/eip712"
	"github.com/tharsis/ethermint/types"
	"math"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/tharsis/ethermint/app"
	ante "github.com/tharsis/ethermint/app/ante"
	"github.com/tharsis/ethermint/encoding"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/statedb"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type AnteTestSuite struct {
	suite.Suite

	ctx             sdk.Context
	app             *app.EthermintApp
	clientCtx       client.Context
	anteHandler     sdk.AnteHandler
	ethSigner       ethtypes.Signer
	enableFeemarket bool
	enableLondonHF  bool
}

func (suite *AnteTestSuite) StateDB() *statedb.StateDB {
	return statedb.New(suite.ctx, suite.app.EvmKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(suite.ctx.HeaderHash().Bytes())))
}

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx, func(app *app.EthermintApp, genesis simapp.GenesisState) simapp.GenesisState {
		if suite.enableFeemarket {
			// setup feemarketGenesis params
			feemarketGenesis := feemarkettypes.DefaultGenesisState()
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
			// Verify feeMarket genesis
			err := feemarketGenesis.Validate()
			suite.Require().NoError(err)
			genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}
		if !suite.enableLondonHF {
			evmGenesis := evmtypes.DefaultGenesisState()
			maxInt := sdk.NewInt(math.MaxInt64)
			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
			evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.MergeForkBlock = &maxInt
			genesis[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
		}
		return genesis
	})

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 2, ChainID: "ethermint_9000-1", Time: time.Now().UTC()})
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(evmtypes.DefaultEVMDenom, sdk.OneInt())))
	suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000000000000000))
	suite.app.EvmKeeper.WithChainID(suite.ctx)

	infCtx := suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	suite.app.AccountKeeper.SetParams(infCtx, authtypes.DefaultParams())

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	// We're using TestMsg amino encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)

	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	options := ante.HandlerOptions{
		AccountKeeper:    suite.app.AccountKeeper,
		BankKeeper:       suite.app.BankKeeper,
		EvmKeeper:        suite.app.EvmKeeper,
		FeegrantKeeper:   suite.app.FeeGrantKeeper,
		IBCChannelKeeper: suite.app.IBCKeeper.ChannelKeeper,
		FeeMarketKeeper:  suite.app.FeeMarketKeeper,
		SignModeHandler:  encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:   ante.DefaultSigVerificationGasConsumer,
	}

	suite.Require().NoError(options.Validate())

	suite.anteHandler = ante.NewAnteHandler(options)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{
		enableLondonHF: true,
	})
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (suite *AnteTestSuite) CreateTestTx(
	msg *evmtypes.MsgEthereumTx, priv cryptotypes.PrivKey, accNum uint64, signCosmosTx bool,
	unsetExtensionOptions ...bool,
) authsigning.Tx {
	return suite.CreateTestTxBuilder(msg, priv, accNum, signCosmosTx).GetTx()
}

// CreateTestTxBuilder is a helper function to create a tx builder given multiple inputs.
func (suite *AnteTestSuite) CreateTestTxBuilder(
	msg *evmtypes.MsgEthereumTx, priv cryptotypes.PrivKey, accNum uint64, signCosmosTx bool,
	unsetExtensionOptions ...bool,
) client.TxBuilder {
	var option *codectypes.Any
	var err error
	if len(unsetExtensionOptions) == 0 {
		option, err = codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
		suite.Require().NoError(err)
	}

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	if len(unsetExtensionOptions) == 0 {
		builder.SetExtensionOptions(option)
	}

	err = msg.Sign(suite.ethSigner, tests.NewSigner(priv))
	suite.Require().NoError(err)

	err = builder.SetMsgs(msg)
	suite.Require().NoError(err)

	txData, err := evmtypes.UnpackTxData(msg.Data)
	suite.Require().NoError(err)

	fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdk.NewIntFromBigInt(txData.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())

	if signCosmosTx {
		// First round: we gather all the signer infos. We use the "set empty
		// signature" hack to do that.
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: txData.GetNonce(),
		}

		sigsV2 := []signing.SignatureV2{sigV2}

		err = txBuilder.SetSignatures(sigsV2...)
		suite.Require().NoError(err)

		// Second round: all signer infos are set, so each signer can sign.

		signerData := authsigning.SignerData{
			ChainID:       suite.ctx.ChainID(),
			AccountNumber: accNum,
			Sequence:      txData.GetNonce(),
		}
		sigV2, err = tx.SignWithPrivKey(
			suite.clientCtx.TxConfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv, suite.clientCtx.TxConfig, txData.GetNonce(),
		)
		suite.Require().NoError(err)

		sigsV2 = []signing.SignatureV2{sigV2}

		err = txBuilder.SetSignatures(sigsV2...)
		suite.Require().NoError(err)
	}

	return txBuilder
}

func (suite *AnteTestSuite) GetTypedData(chainId uint64, msg sdk.Msg, gas uint64, amount sdk.Coins, from sdk.AccAddress) apitypes.TypedData {
	var ethermintCodec codec.ProtoCodecMarshaler
	fee := legacytx.NewStdFee(gas, amount)
	data := legacytx.StdSignBytes("ethermint_9000-1", 1, 1, 0, fee, []sdk.Msg{msg}, "")
	typedData, err := eip712.WrapTxToTypedData(ethermintCodec, chainId, msg, data, &eip712.FeeDelegationOptions{
		FeePayer: from,
	})
	suite.Require().NoError(err)
	return typedData
}

func (suite *AnteTestSuite) CreateTestEIP712CosmosTxBuilder(
	priv cryptotypes.PrivKey, from sdk.AccAddress, to common.Address,
) client.TxBuilder {
	var option *codectypes.Any
	var err error

	amount := sdk.NewCoins(sdk.NewCoin("aphoton", sdk.NewInt(20)))
	gas := uint64(200000)

	//bech32PrefixAccAddr := "ethm"
	//
	//sdk.GetConfig().SetBech32PrefixForAccount("ethm", "ethm")

	// sdk.GetConfig().SetBech32PrefixForAccount("ethm")

	//bz, err := sdk.GetFromBech32("ethm1tfegf50n5xl0hd5cxfzjca3ylsfpg0fned5gqm", bech32PrefixAccAddr)
	//suite.Require().NoError(err)
	//
	//err = sdk.VerifyAddressFormat(bz)
	//suite.Require().NoError(err)
	//
	//recipient := sdk.AccAddress(bz)
	//suite.Require().NoError(err)

	recipient := sdk.AccAddress(common.Address{}.Bytes())

	msgSend := types2.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin("aphoton", sdk.NewInt(1))))

	typedData := suite.GetTypedData(9000, msgSend, gas, amount, from)

	// Sign ethereum TypeData tx
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	suite.Require().NoError(err)

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	suite.Require().NoError(err)

	keyringSigner := tests.NewSigner(priv)
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	sigHash := crypto.Keccak256(rawData)
	signature, _, err := keyringSigner.SignByAddress(from, sigHash)
	suite.Require().NoError(err)

	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	// Add ExtensionOptionsWeb3Tx extension
	option, err = codectypes.NewAnyWithValue(&types.ExtensionOptionsWeb3Tx{
		FeePayer:         from.String(),
		TypedDataChainID: 9000,
		FeePayerSig:      signature,
	})
	suite.Require().NoError(err)

	suite.clientCtx.TxConfig.SignModeHandler()
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	builder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	suite.Require().True(ok)

	builder.SetExtensionOptions(option)
	//sendBytes, err := msgSend.Marshal()
	//suite.Require().NoError(err)

	//msgAny := codectypes.Any{
	//	Value: sendBytes,
	//}

	//msgs := []*codectypes.Any{&msgAny}

	//txBody := tx2.TxBody{
	//	Messages: msgs,
	//}
	//
	//txRaw := tx2.TxRaw{}

	builder.SetFeeAmount(amount)
	builder.SetGasLimit(gas)

	sigsV2 := signing.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		Sequence: 1,
	}

	err = builder.SetSignatures(sigsV2)
	suite.Require().NoError(err)

	err = builder.SetMsgs(msgSend)
	suite.Require().NoError(err)

	// sign the messages
	// set the messages for
	// set fee amount
	// set gas limit

	// return tx builder
	return builder
}

var _ sdk.Tx = &invalidTx{}

type invalidTx struct{}

func (invalidTx) GetMsgs() []sdk.Msg   { return []sdk.Msg{nil} }
func (invalidTx) ValidateBasic() error { return nil }

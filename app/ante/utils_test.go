package ante_test

import (
	"math"
	"testing"
	"time"

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

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx, func(app *app.EthermintApp, genesis simapp.GenesisState) simapp.GenesisState {
		if suite.enableFeemarket {
			// setup feemarketGenesis params
			feemarketGenesis := feemarkettypes.DefaultGenesisState()
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
			feemarketGenesis.BaseFee = sdk.NewInt(feemarketGenesis.Params.InitialBaseFee)
			// Verify feeMarket genesis
			err := feemarketGenesis.Validate()
			suite.Require().NoError(err)
			genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}
		if !suite.enableLondonHF {
			evmGenesis := evmtypes.DefaultGenesisState()
			maxInt := sdk.NewInt(math.MaxInt64)
			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
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

var _ sdk.Tx = &invalidTx{}

type invalidTx struct{}

func (invalidTx) GetMsgs() []sdk.Msg   { return []sdk.Msg{nil} }
func (invalidTx) ValidateBasic() error { return nil }

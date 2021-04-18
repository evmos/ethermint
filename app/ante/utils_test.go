package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/app"
	ante "github.com/cosmos/ethermint/app/ante"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	sidechain "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type AnteTestSuite struct {
	suite.Suite

	ctx            sdk.Context
	app            *app.EthermintApp
	encodingConfig params.EncodingConfig
	anteHandler    sdk.AnteHandler
}

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 2, ChainID: "injective-3", Time: time.Now().UTC()})
	suite.app.AccountKeeper.SetParams(suite.ctx, authtypes.DefaultParams())
	suite.app.EvmKeeper.SetParams(suite.ctx, evmtypes.DefaultParams())

	suite.encodingConfig = simapp.MakeEncodingConfig()
	// We're using TestMsg amino encoding in some tests, so register it here.
	suite.encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)

	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper, suite.encodingConfig.TxConfig.SignModeHandler())
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.NewCoins(sidechain.NewPhotonCoinInt64(500000000))
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(220000, sdk.NewCoins(sidechain.NewPhotonCoinInt64(150)))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	privkey, _ := ethsecp256k1.GenerateKey()
	addr := ethcrypto.PubkeyToAddress(privkey.ToECDSA().PublicKey)

	return sdk.AccAddress(addr.Bytes()), privkey
}

func newTestSDKTx(
	ctx sdk.Context, msgs []sdk.Msg, privs []tmcrypto.PrivKey,
	accNums []uint64, seqs []uint64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(privs))
	for i, priv := range privs {
		signBytes := auth.StdSignBytes(ctx.ChainID(), accNums[i], seqs[i], fee, msgs, "")

		sig, err := priv.Sign(signBytes)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    priv.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
}

func newTestEthTx(ctx sdk.Context, msg *evmtypes.MsgEthereumTx, priv tmcrypto.PrivKey) (sdk.Tx, error) {
	chainIDEpoch, err := sidechain.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	privkey := &ethsecp256k1.PrivKey{Key: priv.Bytes()}

	if err := msg.Sign(chainIDEpoch, privkey.ToECDSA()); err != nil {
		return nil, err
	}

	return msg, nil
}

package ante_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/app"
	ante "github.com/cosmos/ethermint/app/ante"
	"github.com/cosmos/ethermint/crypto"
	emint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

type AnteTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.EthermintApp
	anteHandler sdk.AnteHandler
}

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.app.Codec().RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, ChainID: "3", Time: time.Now().UTC()})
	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.SupplyKeeper)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func newTestMsg(addrs ...sdk.AccAddress) *sdk.TestMsg {
	return sdk.NewTestMsg(addrs...)
}

func newTestCoins() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(emint.DenomDefault, 500000000))
}

func newTestStdFee() auth.StdFee {
	return auth.NewStdFee(220000, sdk.NewCoins(sdk.NewInt64Coin(emint.DenomDefault, 150)))
}

// GenerateAddress generates an Ethereum address.
func newTestAddrKey() (sdk.AccAddress, tmcrypto.PrivKey) {
	privkey, _ := crypto.GenerateKey()
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

func newTestEthTx(ctx sdk.Context, msg evmtypes.MsgEthereumTx, priv tmcrypto.PrivKey) (sdk.Tx, error) {
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return nil, fmt.Errorf("invalid chainID: %s", ctx.ChainID())
	}

	privkey, ok := priv.(crypto.PrivKeySecp256k1)
	if !ok {
		return nil, fmt.Errorf("invalid private key type: %T", priv)
	}

	err := msg.Sign(chainID, privkey.ToECDSA())
	if err != nil {
		return nil, err
	}

	return msg, nil
}

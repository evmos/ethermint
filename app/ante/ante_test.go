package ante_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/app/ante"
	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func requireValidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, sim bool,
) {
	_, err := anteHandler(ctx, tx, sim)
	require.NoError(t, err)
}

func requireInvalidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context,
	tx sdk.Tx, sim bool,
) {
	_, err := anteHandler(ctx, tx, sim)
	require.Error(t, err)
}

func (suite *AnteTestSuite) TestValidEthTx() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr2)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc2)
	err = suite.app.BankKeeper.SetBalances(suite.ctx, acc2.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)
	requireValidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestValidTx() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr2)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc2)
	err = suite.app.BankKeeper.SetBalances(suite.ctx, acc2.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid SDK tx to pass
	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)
	msgs := []sdk.Msg{msg1}

	privKeys := []cryptotypes.PrivKey{priv1, priv2}
	accNums := []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)

	requireValidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestSDKInvalidSigs() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()
	addr3, priv3 := newTestAddrKey()

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr2)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc2)
	err = suite.app.BankKeeper.SetBalances(suite.ctx, acc2.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)

	// require validation failure with no signers
	msgs := []sdk.Msg{msg1}

	privKeys := []cryptotypes.PrivKey{}
	accNums := []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)

	// require validation failure with invalid number of signers
	msgs = []sdk.Msg{msg1}

	privKeys = []cryptotypes.PrivKey{priv1}
	accNums = []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs = []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx = newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)

	// require validation failure with an invalid signer
	msg2 := newTestMsg(addr1, addr3)
	msgs = []sdk.Msg{msg1, msg2}

	privKeys = []cryptotypes.PrivKey{priv1, priv2, priv3}
	accNums = []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber(), 0}
	accSeqs = []uint64{acc1.GetSequence(), acc2.GetSequence(), 0}

	tx = newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestSDKInvalidAcc() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc1.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1)
	msgs := []sdk.Msg{msg1}
	privKeys := []cryptotypes.PrivKey{priv1}

	// require validation failure with invalid account number
	accNums := []uint64{1}
	accSeqs := []uint64{acc1.GetSequence()}

	tx := newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)

	// require validation failure with invalid sequence (nonce)
	accNums = []uint64{acc1.GetAccountNumber()}
	accSeqs = []uint64{1}

	tx = newTestSDKTx(suite.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestEthInvalidSig() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	_, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)

	ctx := suite.ctx.WithChainID("ethermint-4")
	requireInvalidTx(suite.T(), suite.anteHandler, ctx, tx, false)
}

func (suite *AnteTestSuite) TestEthInvalidNonce() {

	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	err := acc.SetSequence(10)
	suite.Require().NoError(err)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err = suite.app.BankKeeper.SetBalances(suite.ctx, acc.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestEthInsufficientBalance() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestEthInvalidIntrinsicGas() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	gasLimit := uint64(1000)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, gasLimit, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx.WithIsCheckTx(true), tx, false)
}

func (suite *AnteTestSuite) TestEthInvalidMempoolFees() {
	// setup app with checkTx = true
	suite.app = app.Setup(true)
	suite.ctx = suite.app.BaseApp.NewContext(true, tmproto.Header{Height: 1, ChainID: "ethermint-3", Time: time.Now().UTC()})
	suite.app.EvmKeeper.SetParams(suite.ctx, evmtypes.DefaultParams())

	suite.anteHandler = ante.NewAnteHandler(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper, suite.encodingConfig.TxConfig.SignModeHandler())
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(types.NewPhotonDecCoin(sdk.NewInt(500000))))
	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("payload"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)
	requireInvalidTx(suite.T(), suite.anteHandler, suite.ctx, tx, false)
}

func (suite *AnteTestSuite) TestEthInvalidChainID() {
	suite.ctx = suite.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr1)
	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	err := suite.app.BankKeeper.SetBalances(suite.ctx, acc.GetAddress(), newTestCoins())
	suite.Require().NoError(err)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(suite.chainID, 0, &to, amt, 22000, gas, []byte("test"), nil)

	tx, err := suite.newTestEthTx(ethMsg, priv1)
	suite.Require().NoError(err)

	ctx := suite.ctx.WithChainID("bad-chain-id")
	requireInvalidTx(suite.T(), suite.anteHandler, ctx, tx, false)
}

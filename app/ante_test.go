package app

import (
	"math/big"
	"testing"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

func requireValidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, sim bool,
) {
	_, err := anteHandler(ctx, tx, sim)
	require.True(t, err == nil)
}

func requireInvalidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context,
	tx sdk.Tx, sim bool, code sdk.CodeType,
) {

	_, err := anteHandler(ctx, tx, sim)
	// require.Equal(t, code, err, fmt.Sprintf("invalid result: %v", err))
	require.Error(t, err)

	if code == sdk.CodeOutOfGas {
		_, ok := tx.(auth.StdTx)
		require.True(t, ok, "tx must be in form auth.StdTx")
	}
}

func TestValidEthTx(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc1 := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc1.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc1)

	acc2 := input.accKeeper.NewAccountWithAddress(input.ctx, addr2)
	// nolint:errcheck
	acc2.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc2)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	requireValidTx(t, input.anteHandler, input.ctx, tx, false)
}

func TestValidTx(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()

	acc1 := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc1.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc1)

	acc2 := input.accKeeper.NewAccountWithAddress(input.ctx, addr2)
	// nolint:errcheck
	acc2.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc2)

	// require a valid SDK tx to pass
	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)
	msgs := []sdk.Msg{msg1}

	privKeys := []tmcrypto.PrivKey{priv1, priv2}
	accNums := []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)

	requireValidTx(t, input.anteHandler, input.ctx, tx, false)
}

func TestSDKInvalidSigs(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()
	addr3, priv3 := newTestAddrKey()

	acc1 := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc1.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc1)

	acc2 := input.accKeeper.NewAccountWithAddress(input.ctx, addr2)
	// nolint:errcheck
	acc2.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc2)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)

	// require validation failure with no signers
	msgs := []sdk.Msg{msg1}

	privKeys := []tmcrypto.PrivKey{}
	accNums := []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeNoSignatures)

	// require validation failure with invalid number of signers
	msgs = []sdk.Msg{msg1}

	privKeys = []tmcrypto.PrivKey{priv1}
	accNums = []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs = []uint64{acc1.GetSequence(), acc2.GetSequence()}

	tx = newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeUnauthorized)

	// require validation failure with an invalid signer
	msg2 := newTestMsg(addr1, addr3)
	msgs = []sdk.Msg{msg1, msg2}

	privKeys = []tmcrypto.PrivKey{priv1, priv2, priv3}
	accNums = []uint64{acc1.GetAccountNumber(), acc2.GetAccountNumber(), 0}
	accSeqs = []uint64{acc1.GetSequence(), acc2.GetSequence(), 0}

	tx = newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeUnknownAddress)
}

func TestSDKInvalidAcc(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()

	acc1 := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc1.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc1)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1)
	msgs := []sdk.Msg{msg1}
	privKeys := []tmcrypto.PrivKey{priv1}

	// require validation failure with invalid account number
	accNums := []uint64{1}
	accSeqs := []uint64{acc1.GetSequence()}

	tx := newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeUnauthorized)

	// require validation failure with invalid sequence (nonce)
	accNums = []uint64{acc1.GetAccountNumber()}
	accSeqs = []uint64{1}

	tx = newTestSDKTx(input.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeUnauthorized)
}

func TestEthInvalidSig(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	_, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	ctx := input.ctx.WithChainID("4")
	requireInvalidTx(t, input.anteHandler, ctx, tx, false, sdk.CodeUnauthorized)
}

func TestEthInvalidNonce(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc.SetCoins(newTestCoins())
	// nolint:errcheck
	acc.SetSequence(10)
	input.accKeeper.SetAccount(input.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeInvalidSequence)
}

func TestEthInsufficientBalance(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	input.accKeeper.SetAccount(input.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeInsufficientFunds)
}

func TestEthInvalidIntrinsicGas(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	gasLimit := uint64(1000)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, gasLimit, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeInternal)
}

func TestEthInvalidMempoolFees(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)
	input.ctx = input.ctx.WithMinGasPrices(sdk.DecCoins{sdk.NewDecCoin(types.DenomDefault, sdk.NewInt(500000))})

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	requireInvalidTx(t, input.anteHandler, input.ctx, tx, false, sdk.CodeInsufficientFee)
}

func TestEthInvalidChainID(t *testing.T) {
	input := newTestSetup()
	input.ctx = input.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, _ := newTestAddrKey()

	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr1)
	// nolint:errcheck
	acc.SetCoins(newTestCoins())
	input.accKeeper.SetAccount(input.ctx, acc)

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewEthereumTxMsg(0, &to, amt, 22000, gas, []byte("test"))

	tx := newTestEthTx(input.ctx, ethMsg, priv1)
	ctx := input.ctx.WithChainID("bad-chain-id")
	requireInvalidTx(t, input.anteHandler, ctx, tx, false, types.CodeInvalidChainID)
}

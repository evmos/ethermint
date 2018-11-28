package app

import (
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func requireValidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, sim bool,
) {

	_, result, abort := anteHandler(ctx, tx, sim)
	require.Equal(t, sdk.CodeOK, result.Code, result.Log)
	require.False(t, abort)
	require.True(t, result.IsOK())
}

func requireInvalidTx(
	t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context,
	tx sdk.Tx, sim bool, code sdk.CodeType,
) {

	newCtx, result, abort := anteHandler(ctx, tx, sim)
	require.True(t, abort)
	require.Equal(t, code, result.Code, fmt.Sprintf("invalid result: %v", result))

	if code == sdk.CodeOutOfGas {
		stdTx, ok := tx.(auth.StdTx)
		require.True(t, ok, "tx must be in form auth.StdTx")

		// require GasWanted is set correctly
		require.Equal(t, stdTx.Fee.Gas, result.GasWanted, "'GasWanted' wanted not set correctly")
		require.True(t, result.GasUsed > result.GasWanted, "'GasUsed' not greater than GasWanted")

		// require that context is set correctly
		require.Equal(t, result.GasUsed, newCtx.GasMeter().GasConsumed(), "Context not updated correctly")
	}
}

func TestValidTx(t *testing.T) {
	setup := newTestSetup()
	setup.ctx = setup.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()

	acc1 := setup.accKeeper.NewAccountWithAddress(setup.ctx, addr1)
	acc1.SetCoins(newTestCoins())
	setup.accKeeper.SetAccount(setup.ctx, acc1)

	acc2 := setup.accKeeper.NewAccountWithAddress(setup.ctx, addr2)
	acc2.SetCoins(newTestCoins())
	setup.accKeeper.SetAccount(setup.ctx, acc2)

	// require a valid SDK tx to pass
	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)
	msgs := []sdk.Msg{msg1}

	privKeys := []crypto.PrivKey{priv1, priv2}
	accNums := []int64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []int64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireValidTx(t, setup.anteHandler, setup.ctx, tx, false)

	// require accounts to update
	acc1 = setup.accKeeper.GetAccount(setup.ctx, addr1)
	acc2 = setup.accKeeper.GetAccount(setup.ctx, addr2)
	require.Equal(t, accSeqs[0]+1, acc1.GetSequence())
	require.Equal(t, accSeqs[1]+1, acc2.GetSequence())

	// require a valid Ethereum tx to pass
	to := ethcmn.BytesToAddress(addr2.Bytes())
	amt := big.NewInt(32)
	gas := big.NewInt(20)
	ethMsg := evmtypes.NewMsgEthereumTx(0, to, amt, 20000, gas, []byte("test"))

	tx = newTestEthTx(setup.ctx, ethMsg, priv1)
	requireValidTx(t, setup.anteHandler, setup.ctx, tx, false)
}

func TestSDKInvalidSigs(t *testing.T) {
	setup := newTestSetup()
	setup.ctx = setup.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()
	addr2, priv2 := newTestAddrKey()
	addr3, priv3 := newTestAddrKey()

	acc1 := setup.accKeeper.NewAccountWithAddress(setup.ctx, addr1)
	acc1.SetCoins(newTestCoins())
	setup.accKeeper.SetAccount(setup.ctx, acc1)

	acc2 := setup.accKeeper.NewAccountWithAddress(setup.ctx, addr2)
	acc2.SetCoins(newTestCoins())
	setup.accKeeper.SetAccount(setup.ctx, acc2)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1, addr2)

	// require validation failure with no signers
	msgs := []sdk.Msg{msg1}

	privKeys := []crypto.PrivKey{}
	accNums := []int64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs := []int64{acc1.GetSequence(), acc2.GetSequence()}

	tx := newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, setup.anteHandler, setup.ctx, tx, false, sdk.CodeUnauthorized)

	// require validation failure with invalid number of signers
	msgs = []sdk.Msg{msg1}

	privKeys = []crypto.PrivKey{priv1}
	accNums = []int64{acc1.GetAccountNumber(), acc2.GetAccountNumber()}
	accSeqs = []int64{acc1.GetSequence(), acc2.GetSequence()}

	tx = newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, setup.anteHandler, setup.ctx, tx, false, sdk.CodeUnauthorized)

	// require validation failure with an invalid signer
	msg2 := newTestMsg(addr1, addr3)
	msgs = []sdk.Msg{msg1, msg2}

	privKeys = []crypto.PrivKey{priv1, priv2, priv3}
	accNums = []int64{acc1.GetAccountNumber(), acc2.GetAccountNumber(), 0}
	accSeqs = []int64{acc1.GetSequence(), acc2.GetSequence(), 0}

	tx = newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, setup.anteHandler, setup.ctx, tx, false, sdk.CodeUnknownAddress)
}

func TestSDKInvalidAcc(t *testing.T) {
	setup := newTestSetup()
	setup.ctx = setup.ctx.WithBlockHeight(1)

	addr1, priv1 := newTestAddrKey()

	acc1 := setup.accKeeper.NewAccountWithAddress(setup.ctx, addr1)
	acc1.SetCoins(newTestCoins())
	setup.accKeeper.SetAccount(setup.ctx, acc1)

	fee := newTestStdFee()
	msg1 := newTestMsg(addr1)
	msgs := []sdk.Msg{msg1}
	privKeys := []crypto.PrivKey{priv1}

	// require validation failure with invalid account number
	accNums := []int64{1}
	accSeqs := []int64{acc1.GetSequence()}

	tx := newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, setup.anteHandler, setup.ctx, tx, false, sdk.CodeInvalidSequence)

	// require validation failure with invalid sequence (nonce)
	accNums = []int64{acc1.GetAccountNumber()}
	accSeqs = []int64{1}

	tx = newTestSDKTx(setup.ctx, msgs, privKeys, accNums, accSeqs, fee)
	requireInvalidTx(t, setup.anteHandler, setup.ctx, tx, false, sdk.CodeInvalidSequence)
}

func TestSDKGasConsumption(t *testing.T) {
	// TODO: Test gas consumption and OOG once ante handler implementation stabilizes
	t.SkipNow()
}

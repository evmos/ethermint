package handlers

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/ethermint/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

func TestEthTxBadSig(t *testing.T) {
	tx := types.NewTransaction(uint64(0), types.TestAddr1, big.NewInt(10), 1000, big.NewInt(100), []byte{})

	// create bad signature
	tx.Sign(big.NewInt(100), types.TestPrivKey2)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, *tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEthTxInsufficientGas(t *testing.T) {
	tx := types.NewTransaction(uint64(0), types.TestAddr1, big.NewInt(0), 0, big.NewInt(0), []byte{})
	tx.Sign(types.TestChainID, types.TestPrivKey1)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, *tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEthTxIncorrectNonce(t *testing.T) {
	// create transaction with wrong nonce 12
	tx := types.NewTransaction(12, types.TestAddr2, big.NewInt(50), 1000, big.NewInt(1000), []byte("test_bytes"))
	tx.Sign(types.TestChainID, types.TestPrivKey1)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, log.NewNopLogger())
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// set account in accountMapper
	acc := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	accMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, *tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEthTxValidTx(t *testing.T) {
	tx := types.NewTransaction(0, types.TestAddr1, big.NewInt(50), 1000, big.NewInt(1000), []byte{})
	tx.Sign(types.TestChainID, types.TestPrivKey1)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// set account in accountMapper
	acc := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	accMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, *tx)

	require.False(t, abort, "expected ante handler to not abort")
	require.True(t, res.IsOK(), fmt.Sprintf("result not OK on valid Tx: %s", res.Log))

	// ensure account state updated correctly
	seq, _ := accMapper.GetSequence(ctx, types.TestAddr1[:])
	require.Equal(t, int64(1), seq, "account nonce did not increment correctly")
}

func TestEmbeddedTxBadSig(t *testing.T) {
	testCodec := types.NewTestCodec()
	testFee := types.NewTestStdFee()

	msgs := []sdk.Msg{sdk.NewTestMsg()}
	tx := types.NewTestStdTx(
		types.TestChainID, msgs, []int64{0}, []int64{0}, []*ecdsa.PrivateKey{types.TestPrivKey1}, testFee,
	)

	// create bad signature
	signBytes := types.GetStdTxSignBytes(big.NewInt(100).String(), 1, 1, testFee, msgs, "")
	sig, _ := ethcrypto.Sign(signBytes, types.TestPrivKey1)
	(tx.(auth.StdTx)).Signatures[0].Signature = sig

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(testCodec, key, auth.ProtoBaseAccount)

	// set account in accountMapper
	acc := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	accMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEmbeddedTxInvalidMultiMsg(t *testing.T) {
	testCodec := types.NewTestCodec()
	testCodec.RegisterConcrete(stake.MsgDelegate{}, "test/MsgDelegate", nil)

	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
		stake.NewMsgDelegate(types.TestAddr2.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
	}

	// create transaction with only one signer
	tx := types.NewTestStdTx(
		types.TestChainID, msgs, []int64{0}, []int64{0}, []*ecdsa.PrivateKey{types.TestPrivKey1}, types.NewTestStdFee(),
	)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(testCodec, key, auth.ProtoBaseAccount)

	// set accounts in accountMapper
	acc1 := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	accMapper.SetAccount(ctx, acc1)

	acc2 := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	accMapper.SetAccount(ctx, acc2)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEmbeddedTxInvalidAccountNumber(t *testing.T) {
	testCodec := types.NewTestCodec()
	testCodec.RegisterConcrete(stake.MsgDelegate{}, "test/MsgDelegate", nil)

	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
	}

	// create a transaction with an invalid account number
	tx := types.NewTestStdTx(
		types.TestChainID, msgs, []int64{3}, []int64{0}, []*ecdsa.PrivateKey{types.TestPrivKey1}, types.NewTestStdFee(),
	)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(testCodec, key, auth.ProtoBaseAccount)

	// set account in accountMapper
	acc := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	acc.SetAccountNumber(4)
	accMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEmbeddedTxInvalidSequence(t *testing.T) {
	testCodec := types.NewTestCodec()
	testCodec.RegisterConcrete(stake.MsgDelegate{}, "test/MsgDelegate", nil)

	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
	}

	// create transaction with an invalid sequence (nonce)
	tx := types.NewTestStdTx(
		types.TestChainID, msgs, []int64{4}, []int64{2}, []*ecdsa.PrivateKey{types.TestPrivKey1}, types.NewTestStdFee(),
	)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// set account in accountMapper
	acc := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	acc.SetAccountNumber(4)
	acc.SetSequence(3)
	accMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, tx)

	require.True(t, abort, "expected ante handler to abort")
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, fmt.Sprintf("invalid code returned on bad tx: %s", res.Log))
}

func TestEmbeddedTxValid(t *testing.T) {
	testCodec := types.NewTestCodec()
	testCodec.RegisterConcrete(stake.MsgDelegate{}, "test/MsgDelegate", nil)

	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
		stake.NewMsgDelegate(types.TestAddr2.Bytes(), types.TestAddr2.Bytes(), sdk.NewCoin("steak", sdk.NewInt(50))),
	}

	// create a valid transaction
	tx := types.NewTestStdTx(
		types.TestChainID, msgs, []int64{4, 5}, []int64{3, 1},
		[]*ecdsa.PrivateKey{types.TestPrivKey1, types.TestPrivKey2}, types.NewTestStdFee(),
	)

	ms, key := createTestMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)
	accMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// set accounts in the accountMapper
	acc1 := accMapper.NewAccountWithAddress(ctx, types.TestAddr1.Bytes())
	acc1.SetAccountNumber(4)
	acc1.SetSequence(3)
	accMapper.SetAccount(ctx, acc1)

	acc2 := accMapper.NewAccountWithAddress(ctx, types.TestAddr2.Bytes())
	acc2.SetAccountNumber(5)
	acc2.SetSequence(1)
	accMapper.SetAccount(ctx, acc2)

	handler := AnteHandler(accMapper, auth.FeeCollectionKeeper{})
	_, res, abort := handler(ctx, tx)

	require.False(t, abort, "expected ante handler to not abort")
	require.True(t, res.IsOK(), fmt.Sprintf("result not OK on valid Tx: %s", res.Log))

	// Ensure account state updated correctly
	seq1, _ := accMapper.GetSequence(ctx, types.TestAddr1.Bytes())
	seq2, _ := accMapper.GetSequence(ctx, types.TestAddr2.Bytes())

	require.Equal(t, int64(4), seq1, "account nonce did not increment correctly")
	require.Equal(t, int64(2), seq2, "account nonce did not increment correctly")
}

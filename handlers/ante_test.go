package handlers

import (
	"fmt"
	"testing"
	"math/big"
	"crypto/ecdsa"

	"github.com/cosmos/ethermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test EthTx Antehandler
// -----------------------------------------------------------------------------------------------------------------------------------

func TestBadSig(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.Signature = types.NewEthSignature(new(big.Int), new(big.Int), new(big.Int))

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

}

func TestInsufficientGas(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	tx.Data.GasLimit = 0

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x1000c), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

}

func TestIncorrectNonce(t *testing.T) {
	// Create transaction with wrong nonce 12
	tx := types.NewTransaction(12, types.TestAddr2, sdk.NewInt(50), 1000, sdk.NewInt(1000), []byte("test_bytes"))
	tx.Sign(types.TestChainID, types.TestPrivKey1)

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, log.NewNopLogger())

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort")
	require.Equal(t, sdk.ABCICodeType(0x10003), res.Code, fmt.Sprintf("Result has wrong code on bad tx: %s", res.Log))

}

func TestValidTx(t *testing.T) {
	tx := types.NewTestEthTxs(types.TestChainID, []*ecdsa.PrivateKey{types.TestPrivKey1}, []ethcmn.Address{types.TestAddr1})[0]

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.False(t, abort, "Antehandler abort on valid tx")
	require.True(t, res.IsOK(), fmt.Sprintf("Result not OK on valid Tx: %s", res.Log))

	// Ensure account state updated correctly
	seq, _ := accountMapper.GetSequence(ctx, types.TestAddr1[:])
	require.Equal(t, int64(1), seq, "AccountNonce did not increment correctly")
}

// Test EmbeddedTx Antehandler
// -----------------------------------------------------------------------------------------------------------------------------------

func TestInvalidSigEmbeddedTx(t *testing.T) {
	// Create msg to be embedded
	msgs := []sdk.Msg{stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)})}

	// Create transaction with incorrect signer
	tx := types.NewTestEmbeddedTx(types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey2},
		 []int64{0}, []int64{0}, types.NewStdFee())

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort on invalid embedded tx")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")
}

func TestInvalidMultiMsgEmbeddedTx(t *testing.T) {
	// Create msg to be embedded
	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
		stake.NewMsgDelegate(types.TestAddr2[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
	}

	// Create transaction with only one signer
	tx := types.NewTestEmbeddedTx(types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1},
		[]int64{0}, []int64{0}, types.NewStdFee())

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc1 := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc1)
	acc2 := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	accountMapper.SetAccount(ctx, acc2)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort on invalid embedded tx")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")
}

func TestInvalidAccountNumberEmbeddedTx(t *testing.T) {
	// Create msg to be embedded
	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
	}

	// Create transaction with wrong account number
	tx := types.NewTestEmbeddedTx(types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1},
		[]int64{3}, []int64{0}, types.NewStdFee())

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	acc.SetAccountNumber(4)
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort on invalid embedded tx")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")
}

func TestInvalidSequenceEmbeddedTx(t *testing.T) {
	// Create msg to be embedded
	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
	}

	// Create transaction with wrong account number
	tx := types.NewTestEmbeddedTx(types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1},
		[]int64{4}, []int64{2}, types.NewStdFee())

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)

	// Set account in accountMapper
	acc := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	acc.SetAccountNumber(4)
	acc.SetSequence(3)
	accountMapper.SetAccount(ctx, acc)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	assert.True(t, abort, "Antehandler did not abort on invalid embedded tx")
	require.Equal(t, sdk.ABCICodeType(0x10004), res.Code, "Result is OK on bad tx")
}

func TestValidEmbeddedTx(t *testing.T) {
	// Create msg to be embedded
	msgs := []sdk.Msg{
		stake.NewMsgDelegate(types.TestAddr1[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
		stake.NewMsgDelegate(types.TestAddr2[:], types.TestAddr2[:], sdk.Coin{"steak", sdk.NewInt(50)}),
	}

	tx := types.NewTestEmbeddedTx(types.TestChainID, msgs, []*ecdsa.PrivateKey{types.TestPrivKey1, types.TestPrivKey2},
		 []int64{4, 5}, []int64{3, 1}, types.NewStdFee())

	ms, key := SetupMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{ChainID: types.TestChainID.String()}, false, nil)

	accountMapper := auth.NewAccountMapper(types.NewTestCodec(), key, auth.ProtoBaseAccount)
	
	// Set account in accountMapper
	acc1 := accountMapper.NewAccountWithAddress(ctx, types.TestAddr1[:])
	acc1.SetAccountNumber(4)
	acc1.SetSequence(3)
	accountMapper.SetAccount(ctx, acc1)
	acc2 := accountMapper.NewAccountWithAddress(ctx, types.TestAddr2[:])
	acc2.SetAccountNumber(5)
	acc2.SetSequence(1)
	accountMapper.SetAccount(ctx, acc2)

	handler := AnteHandler(accountMapper)

	_, res, abort := handler(ctx, tx)

	require.False(t, abort, "Antehandler abort on valid embedded tx")
	require.True(t, res.IsOK(), fmt.Sprintf("Result not OK on valid Tx: %s", res.Log))

	// Ensure account state updated correctly
	seq1, _ := accountMapper.GetSequence(ctx, types.TestAddr1[:])
	seq2, _ := accountMapper.GetSequence(ctx, types.TestAddr2[:])

	assert.Equal(t, int64(4), seq1, "Sequence1 did not increment correctly")
	assert.Equal(t, int64(2), seq2, "Sequence2 did not increment correctly")

}




package types

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestValidation(t *testing.T) {
	ethTxs := NewTestEthTxs(
		TestChainID,
		[]*ecdsa.PrivateKey{TestPrivKey1},
		[]ethcmn.Address{TestAddr1},
	)

	testCases := []struct {
		msg         sdk.Msg
		mutate      func(sdk.Msg) sdk.Msg
		expectedErr bool
	}{
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg { return msg }, false},
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg {
			tx := msg.(Transaction)
			tx.Data.Price = sdk.NewInt(-1)
			return tx
		}, true},
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg {
			tx := msg.(Transaction)
			tx.Data.Amount = sdk.NewInt(-1)
			return tx
		}, true},
	}

	for i, tc := range testCases {
		msg := tc.mutate(tc.msg)
		err := msg.ValidateBasic()

		if tc.expectedErr {
			require.NotEqual(t, sdk.CodeOK, err.Code(), fmt.Sprintf("expected error: test case #%d", i))
		} else {
			require.NoError(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		}
	}
}

func TestHasEmbeddedTx(t *testing.T) {
	testCodec := NewTestCodec()
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}

	sdkTxs := NewTestSDKTxs(
		testCodec, TestChainID, msgs, []*ecdsa.PrivateKey{TestPrivKey1},
		[]int64{0}, []int64{0}, NewStdFee(),
	)
	require.True(t, sdkTxs[0].HasEmbeddedTx(TestSDKAddress))

	ethTxs := NewTestEthTxs(
		TestChainID,
		[]*ecdsa.PrivateKey{TestPrivKey1},
		[]ethcmn.Address{TestAddr1},
	)
	require.False(t, ethTxs[0].HasEmbeddedTx(TestSDKAddress))
}

func TestGetEmbeddedTx(t *testing.T) {
	testCodec := NewTestCodec()
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}

	ethTxs := NewTestEthTxs(
		TestChainID,
		[]*ecdsa.PrivateKey{TestPrivKey1},
		[]ethcmn.Address{TestAddr1},
	)
	sdkTxs := NewTestSDKTxs(
		testCodec, TestChainID, msgs, []*ecdsa.PrivateKey{TestPrivKey1},
		[]int64{0}, []int64{0}, NewStdFee(),
	)

	etx, err := sdkTxs[0].GetEmbeddedTx(testCodec)
	require.NoError(t, err)
	require.NotEmpty(t, etx.Messages)

	etx, err = ethTxs[0].GetEmbeddedTx(testCodec)
	require.Error(t, err)
	require.Empty(t, etx.Messages)
}

func TestTransactionGetMsgs(t *testing.T) {
	ethTxs := NewTestEthTxs(
		TestChainID,
		[]*ecdsa.PrivateKey{TestPrivKey1},
		[]ethcmn.Address{TestAddr1},
	)

	msgs := ethTxs[0].GetMsgs()
	require.Len(t, msgs, 1)
	require.Equal(t, ethTxs[0], msgs[0])

	expectedMsgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}
	etx := NewTestEmbeddedTx(
		TestChainID, expectedMsgs, []*ecdsa.PrivateKey{TestPrivKey1},
		[]int64{0}, []int64{0}, NewStdFee(),
	)

	msgs = etx.GetMsgs()
	require.Len(t, msgs, len(expectedMsgs))
	require.Equal(t, expectedMsgs, msgs)
}

func TestGetRequiredSigners(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}
	etx := NewTestEmbeddedTx(
		TestChainID, msgs, []*ecdsa.PrivateKey{TestPrivKey1},
		[]int64{0}, []int64{0}, NewStdFee(),
	)

	signers := etx.(EmbeddedTx).GetRequiredSigners()
	require.Equal(t, []sdk.AccAddress{sdk.AccAddress(TestAddr1.Bytes())}, signers)
}

func TestVerifySig(t *testing.T) {
	ethTx := NewTestEthTxs(
		TestChainID,
		[]*ecdsa.PrivateKey{TestPrivKey1},
		[]ethcmn.Address{TestAddr1},
	)[0]

	addr, err := ethTx.VerifySig(TestChainID.BigInt())
	
	require.Nil(t, err, "Sig verification failed")
	require.Equal(t, TestAddr1, addr, "Address is not the same")
}

func TestTxDecoder(t *testing.T) {
	testCodec := NewTestCodec()
	txDecoder := TxDecoder(testCodec, TestSDKAddress)
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(TestAddr1.Bytes()))}

	// create a non-SDK Ethereum transaction
	emintTx := NewTransaction(
		uint64(0), TestAddr1, sdk.NewInt(10), 100, sdk.NewInt(100), nil,
	)
	emintTx.Sign(TestChainID, TestPrivKey1)

	// require the transaction to properly decode into a Transaction
	txBytes := testCodec.MustMarshalBinary(emintTx)
	tx, err := txDecoder(txBytes)
	require.NoError(t, err)
	require.Equal(t, emintTx, tx)

	// create embedded transaction and encode
	etx := NewTestEmbeddedTx(
		TestChainID, msgs, []*ecdsa.PrivateKey{TestPrivKey1},
		[]int64{0}, []int64{0}, NewStdFee(),
	)

	payload := testCodec.MustMarshalBinary(etx)

	expectedEtx := EmbeddedTx{}
	testCodec.UnmarshalBinary(payload, &expectedEtx)

	emintTx = NewTransaction(
		uint64(0), TestSDKAddress, sdk.NewInt(10), 100,
		sdk.NewInt(100), payload,
	)
	emintTx.Sign(TestChainID, TestPrivKey1)

	// require the transaction to properly decode into a Transaction
	txBytes = testCodec.MustMarshalBinary(emintTx)
	tx, err = txDecoder(txBytes)
	require.NoError(t, err)
	require.Equal(t, expectedEtx, tx)

	// require the decoding to fail when no transaction bytes are given
	tx, err = txDecoder([]byte{})
	require.Error(t, err)
	require.Nil(t, tx)

	// create a non-SDK Ethereum transaction with an SDK address and garbage payload
	emintTx = NewTransaction(
		uint64(0), TestSDKAddress, sdk.NewInt(10), 100, sdk.NewInt(100), []byte("garbage"),
	)
	emintTx.Sign(TestChainID, TestPrivKey1)

	// require the transaction to fail decoding as the payload is invalid
	txBytes = testCodec.MustMarshalBinary(emintTx)
	tx, err = txDecoder(txBytes)
	require.Error(t, err)
	require.Nil(t, tx)
}

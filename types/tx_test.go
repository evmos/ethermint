package types

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestTransactionRLPEncode(t *testing.T) {
	txs := NewTestEthTxs(TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1})
	gtxs := NewTestGethTxs(TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1})

	txRLP, err := rlp.EncodeToBytes(txs[0])
	require.NoError(t, err)

	gtxRLP, err := rlp.EncodeToBytes(gtxs[0])
	require.NoError(t, err)

	require.Equal(t, gtxRLP, txRLP)
}

func TestTransactionRLPDecode(t *testing.T) {
	txs := NewTestEthTxs(TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1})
	gtxs := NewTestGethTxs(TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1})

	txRLP, err := rlp.EncodeToBytes(txs[0])
	require.NoError(t, err)

	gtxRLP, err := rlp.EncodeToBytes(gtxs[0])
	require.NoError(t, err)

	var (
		decodedTx  Transaction
		decodedGtx ethtypes.Transaction
	)

	err = rlp.DecodeBytes(txRLP, &decodedTx)
	require.NoError(t, err)

	err = rlp.DecodeBytes(gtxRLP, &decodedGtx)
	require.NoError(t, err)

	require.Equal(t, decodedGtx.Hash(), decodedTx.Hash())
}

func TestValidation(t *testing.T) {
	ethTxs := NewTestEthTxs(
		TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1},
	)

	testCases := []struct {
		msg         sdk.Msg
		mutate      func(sdk.Msg) sdk.Msg
		expectedErr bool
	}{
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg { return msg }, false},
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg {
			tx := msg.(*Transaction)
			tx.data.Price = big.NewInt(-1)
			return tx
		}, true},
		{ethTxs[0], func(msg sdk.Msg) sdk.Msg {
			tx := msg.(*Transaction)
			tx.data.Amount = big.NewInt(-1)
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

func TestTransactionVerifySig(t *testing.T) {
	txs := NewTestEthTxs(
		TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1},
	)

	addr, err := txs[0].VerifySig(TestChainID)
	require.NoError(t, err)
	require.Equal(t, TestAddr1, addr)

	addr, err = txs[0].VerifySig(big.NewInt(100))
	require.Error(t, err)
	require.NotEqual(t, TestAddr1, addr)
}

func TestTxDecoder(t *testing.T) {
	testCodec := NewTestCodec()
	txDecoder := TxDecoder(testCodec, TestSDKAddr)
	msgs := []sdk.Msg{sdk.NewTestMsg()}

	// create a non-SDK Ethereum transaction
	txs := NewTestEthTxs(
		TestChainID, []int64{0}, []ethcmn.Address{TestAddr1}, []*ecdsa.PrivateKey{TestPrivKey1},
	)

	txBytes, err := rlp.EncodeToBytes(txs[0])
	require.NoError(t, err)

	// require the transaction to properly decode into a Transaction
	decodedTx, err := txDecoder(txBytes)
	require.NoError(t, err)
	require.IsType(t, Transaction{}, decodedTx)
	require.Equal(t, txs[0].data, (decodedTx.(Transaction)).data)

	// create a SDK (auth.StdTx) transaction and encode
	txs = NewTestSDKTxs(
		testCodec, TestChainID, TestSDKAddr, msgs, []int64{0}, []int64{0},
		[]*ecdsa.PrivateKey{TestPrivKey1}, NewTestStdFee(),
	)

	txBytes, err = rlp.EncodeToBytes(txs[0])
	require.NoError(t, err)

	// require the transaction to properly decode into a Transaction
	stdTx := NewTestStdTx(TestChainID, msgs, []int64{0}, []int64{0}, []*ecdsa.PrivateKey{TestPrivKey1}, NewTestStdFee())
	decodedTx, err = txDecoder(txBytes)
	require.NoError(t, err)
	require.IsType(t, auth.StdTx{}, decodedTx)
	require.Equal(t, stdTx, decodedTx)

	// require the decoding to fail when no transaction bytes are given
	decodedTx, err = txDecoder([]byte{})
	require.Error(t, err)
	require.Nil(t, decodedTx)
}

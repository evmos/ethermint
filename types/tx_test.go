package types

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var (
	testChainID = sdk.NewInt(3)

	testPrivKey1, _ = ethcrypto.GenerateKey()
	testPrivKey2, _ = ethcrypto.GenerateKey()

	testAddr1 = PrivKeyToEthAddress(testPrivKey1)
	testAddr2 = PrivKeyToEthAddress(testPrivKey2)

	testSDKAddress = GenerateEthAddress()
)

func newTestCodec() *wire.Codec {
	codec := wire.NewCodec()

	RegisterWire(codec)
	codec.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)

	return codec
}

func newStdFee() auth.StdFee {
	return auth.NewStdFee(5000, sdk.NewCoin("photon", 150))
}

func newTestEmbeddedTx(
	chainID sdk.Int, msgs []sdk.Msg, pKeys []*ecdsa.PrivateKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([][]byte, len(pKeys))

	for i, priv := range pKeys {
		signEtx := EmbeddedTxSign{chainID.String(), accNums[i], seqs[i], msgs, newStdFee()}

		signBytes, err := signEtx.Bytes()
		if err != nil {
			panic(err)
		}

		sig, err := ethcrypto.Sign(signBytes, priv)
		if err != nil {
			panic(err)
		}

		sigs[i] = sig
	}

	return EmbeddedTx{msgs, fee, sigs}
}

func newTestGethTxs(chainID sdk.Int, pKeys []*ecdsa.PrivateKey, addrs []ethcmn.Address) []ethtypes.Transaction {
	txs := make([]ethtypes.Transaction, len(pKeys))

	for i, priv := range pKeys {
		ethTx := ethtypes.NewTransaction(
			uint64(i), addrs[i], big.NewInt(10), 100, big.NewInt(100), nil,
		)

		signer := ethtypes.NewEIP155Signer(chainID.BigInt())
		ethTx, _ = ethtypes.SignTx(ethTx, signer, priv)

		txs[i] = *ethTx
	}

	return txs
}

func newTestEthTxs(chainID sdk.Int, pKeys []*ecdsa.PrivateKey, addrs []ethcmn.Address) []Transaction {
	txs := make([]Transaction, len(pKeys))

	for i, priv := range pKeys {
		emintTx := NewTransaction(
			uint64(i), addrs[i], sdk.NewInt(10), 100, sdk.NewInt(100), nil,
		)

		emintTx.Sign(chainID, priv)

		txs[i] = emintTx
	}

	return txs
}

func newTestSDKTxs(
	codec *wire.Codec, chainID sdk.Int, msgs []sdk.Msg, pKeys []*ecdsa.PrivateKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) []Transaction {

	txs := make([]Transaction, len(pKeys))
	etx := newTestEmbeddedTx(chainID, msgs, pKeys, accNums, seqs, fee)

	for i, priv := range pKeys {
		payload := codec.MustMarshalBinary(etx)

		emintTx := NewTransaction(
			uint64(i), testSDKAddress, sdk.NewInt(10), 100,
			sdk.NewInt(100), payload,
		)

		emintTx.Sign(testChainID, priv)

		txs[i] = emintTx
	}

	return txs
}

func TestConvertTx(t *testing.T) {
	gethTxs := newTestGethTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1, testPrivKey2},
		[]ethcmn.Address{testAddr1, testAddr2},
	)
	ethTxs := newTestEthTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1, testPrivKey2},
		[]ethcmn.Address{testAddr1, testAddr2},
	)

	testCases := []struct {
		ethTx      ethtypes.Transaction
		emintTx    Transaction
		expectedEq bool
	}{
		{gethTxs[0], ethTxs[0], true},
		{gethTxs[0], ethTxs[1], false},
		{gethTxs[1], ethTxs[0], false},
	}

	for i, tc := range testCases {
		convertedTx := tc.emintTx.ConvertTx(testChainID.BigInt())

		if tc.expectedEq {
			require.Equal(t, tc.ethTx, convertedTx, fmt.Sprintf("unexpected result: test case #%d", i))
		} else {
			require.NotEqual(t, tc.ethTx, convertedTx, fmt.Sprintf("unexpected result: test case #%d", i))
		}
	}
}

func TestValidation(t *testing.T) {
	ethTxs := newTestEthTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1},
		[]ethcmn.Address{testAddr1},
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
	testCodec := newTestCodec()
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}

	sdkTxs := newTestSDKTxs(
		testCodec, testChainID, msgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)
	require.True(t, sdkTxs[0].HasEmbeddedTx(testSDKAddress))

	ethTxs := newTestEthTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1},
		[]ethcmn.Address{testAddr1},
	)
	require.False(t, ethTxs[0].HasEmbeddedTx(testSDKAddress))
}

func TestGetEmbeddedTx(t *testing.T) {
	testCodec := newTestCodec()
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}

	ethTxs := newTestEthTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1},
		[]ethcmn.Address{testAddr1},
	)
	sdkTxs := newTestSDKTxs(
		testCodec, testChainID, msgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	etx, err := sdkTxs[0].GetEmbeddedTx(testCodec)
	require.NoError(t, err)
	require.NotEmpty(t, etx.Messages)

	etx, err = ethTxs[0].GetEmbeddedTx(testCodec)
	require.Error(t, err)
	require.Empty(t, etx.Messages)
}

func TestTransactionGetMsgs(t *testing.T) {
	ethTxs := newTestEthTxs(
		testChainID,
		[]*ecdsa.PrivateKey{testPrivKey1},
		[]ethcmn.Address{testAddr1},
	)

	msgs := ethTxs[0].GetMsgs()
	require.Len(t, msgs, 1)
	require.Equal(t, ethTxs[0], msgs[0])

	expectedMsgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}
	etx := newTestEmbeddedTx(
		testChainID, expectedMsgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	msgs = etx.GetMsgs()
	require.Len(t, msgs, len(expectedMsgs))
	require.Equal(t, expectedMsgs, msgs)
}

func TestGetRequiredSigners(t *testing.T) {
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}
	etx := newTestEmbeddedTx(
		testChainID, msgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	signers := etx.(EmbeddedTx).GetRequiredSigners()
	require.Equal(t, []sdk.AccAddress{sdk.AccAddress(testAddr1.Bytes())}, signers)
}

func TestTxDecoder(t *testing.T) {
	testCodec := newTestCodec()
	txDecoder := TxDecoder(testCodec, testSDKAddress)
	msgs := []sdk.Msg{sdk.NewTestMsg(sdk.AccAddress(testAddr1.Bytes()))}

	// create a non-SDK Ethereum transaction
	emintTx := NewTransaction(
		uint64(0), testAddr1, sdk.NewInt(10), 100, sdk.NewInt(100), nil,
	)
	emintTx.Sign(testChainID, testPrivKey1)

	// require the transaction to properly decode into a Transaction
	txBytes := testCodec.MustMarshalBinary(emintTx)
	tx, err := txDecoder(txBytes)
	require.NoError(t, err)
	require.Equal(t, emintTx, tx)

	// create embedded transaction and encode
	etx := newTestEmbeddedTx(
		testChainID, msgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	payload := testCodec.MustMarshalBinary(etx)

	expectedEtx := EmbeddedTx{}
	testCodec.UnmarshalBinary(payload, &expectedEtx)

	emintTx = NewTransaction(
		uint64(0), testSDKAddress, sdk.NewInt(10), 100,
		sdk.NewInt(100), payload,
	)
	emintTx.Sign(testChainID, testPrivKey1)

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
		uint64(0), testSDKAddress, sdk.NewInt(10), 100, sdk.NewInt(100), []byte("garbage"),
	)
	emintTx.Sign(testChainID, testPrivKey1)

	// require the transaction to fail decoding as the payload is invalid
	txBytes = testCodec.MustMarshalBinary(emintTx)
	tx, err = txDecoder(txBytes)
	require.Error(t, err)
	require.Nil(t, tx)
}

package types

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

var (
	testChainID = sdk.NewInt(3)

	testPrivKey1, _ = ethcrypto.GenerateKey()
	testPrivKey2, _ = ethcrypto.GenerateKey()

	testAddr1 = PrivKeyToEthAddress(testPrivKey1)
	testAddr2 = PrivKeyToEthAddress(testPrivKey2)
)

func newTestCodec() *wire.Codec {
	codec := wire.NewCodec()

	RegisterWire(codec)
	codec.RegisterConcrete(auth.StdTx{}, "test/StdTx", nil)
	codec.RegisterConcrete(&sdk.TestMsg{}, "test/TestMsg", nil)
	wire.RegisterCrypto(codec)

	return codec
}

func newStdFee() auth.StdFee {
	return auth.NewStdFee(5000, sdk.NewCoin("photon", sdk.NewInt(150)))
}

func newTestStdTx(
	chainID sdk.Int, msgs []sdk.Msg, pKeys []*ecdsa.PrivateKey,
	accNums []int64, seqs []int64, fee auth.StdFee,
) sdk.Tx {

	sigs := make([]auth.StdSignature, len(pKeys))

	for i, priv := range pKeys {
		signBytes := GetStdTxSignBytes(chainID.String(), accNums[i], seqs[i], newStdFee(), msgs, "")

		sig, err := ethcrypto.Sign(signBytes, priv)
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{Signature: sig, AccountNumber: accNums[i], Sequence: seqs[i]}
	}

	return auth.NewStdTx(msgs, fee, sigs, "")
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
	etx := newTestStdTx(
		testChainID, expectedMsgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	msgs = etx.GetMsgs()
	require.Len(t, msgs, len(expectedMsgs))
	require.Equal(t, expectedMsgs, msgs)
}

func TestTxDecoder(t *testing.T) {
	testCodec := newTestCodec()
	txDecoder := TxDecoder(testCodec)
	msgs := []sdk.Msg{sdk.NewTestMsg()}

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

	// create a SDK (auth.StdTx) transaction and encode
	stdTx := newTestStdTx(
		testChainID, msgs, []*ecdsa.PrivateKey{testPrivKey1},
		[]int64{0}, []int64{0}, newStdFee(),
	)

	// require the transaction to properly decode into a Transaction
	txBytes = testCodec.MustMarshalBinary(stdTx)
	tx, err = txDecoder(txBytes)
	require.NoError(t, err)
	require.Equal(t, stdTx, tx)

	// require the decoding to fail when no transaction bytes are given
	tx, err = txDecoder([]byte{})
	require.Error(t, err)
	require.Nil(t, tx)
}

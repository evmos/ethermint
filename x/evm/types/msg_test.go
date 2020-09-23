package types

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/crypto"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestMsgEthermint(t *testing.T) {
	addr := newSdkAddress()
	fromAddr := newSdkAddress()

	msg := NewMsgEthermint(0, &addr, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), fromAddr)
	require.NotNil(t, msg)
	require.Equal(t, msg.Recipient, &addr)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), TypeMsgEthermint)
}

func TestMsgEthermintValidation(t *testing.T) {
	testCases := []struct {
		nonce      uint64
		to         *sdk.AccAddress
		amount     sdk.Int
		gasLimit   uint64
		gasPrice   sdk.Int
		payload    []byte
		expectPass bool
		from       sdk.AccAddress
	}{
		{amount: sdk.NewInt(100), gasPrice: sdk.NewInt(100000), expectPass: true},
		{amount: sdk.NewInt(0), gasPrice: sdk.NewInt(100000), expectPass: true},
		{amount: sdk.NewInt(-1), gasPrice: sdk.NewInt(100000), expectPass: false},
		{amount: sdk.NewInt(100), gasPrice: sdk.NewInt(-1), expectPass: false},
		{amount: sdk.NewInt(100), gasPrice: sdk.NewInt(0), expectPass: false},
	}

	for i, tc := range testCases {
		msg := NewMsgEthermint(tc.nonce, tc.to, tc.amount, tc.gasLimit, tc.gasPrice, tc.payload, tc.from)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgEthermintEncodingAndDecoding(t *testing.T) {
	addr := newSdkAddress()
	fromAddr := newSdkAddress()

	msg := NewMsgEthermint(0, &addr, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), fromAddr)

	raw, err := ModuleCdc.MarshalBinaryBare(msg)
	require.NoError(t, err)

	var msg2 MsgEthermint
	err = ModuleCdc.UnmarshalBinaryBare(raw, &msg2)
	require.NoError(t, err)

	require.Equal(t, msg.AccountNonce, msg2.AccountNonce)
	require.Equal(t, msg.Recipient, msg2.Recipient)
	require.Equal(t, msg.Amount, msg2.Amount)
	require.Equal(t, msg.GasLimit, msg2.GasLimit)
	require.Equal(t, msg.Price, msg2.Price)
	require.Equal(t, msg.Payload, msg2.Payload)
	require.Equal(t, msg.From, msg2.From)
}

func newSdkAddress() sdk.AccAddress {
	tmpKey := secp256k1.GenPrivKey().PubKey()
	return sdk.AccAddress(tmpKey.Address().Bytes())
}

func TestMsgEthereumTx(t *testing.T) {
	addr := GenerateEthAddress()

	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg)
	require.Equal(t, *msg.Data.Recipient, addr)
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), TypeMsgEthereumTx)
	require.NotNil(t, msg.To())
	require.Equal(t, msg.GetMsgs(), []sdk.Msg{msg})
	require.Panics(t, func() { msg.GetSigners() })
	require.Panics(t, func() { msg.GetSignBytes() })

	msg = NewMsgEthereumTxContract(0, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg)
	require.Nil(t, msg.Data.Recipient)
	require.Nil(t, msg.To())
}

func TestMsgEthereumTxValidation(t *testing.T) {
	testCases := []struct {
		msg        string
		amount     *big.Int
		gasPrice   *big.Int
		expectPass bool
	}{
		{msg: "pass", amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{msg: "invalid amount", amount: big.NewInt(-1), gasPrice: big.NewInt(100000), expectPass: false},
		{msg: "invalid gas price", amount: big.NewInt(100), gasPrice: big.NewInt(-1), expectPass: false},
		{msg: "invalid gas price", amount: big.NewInt(100), gasPrice: big.NewInt(0), expectPass: false},
	}

	for i, tc := range testCases {
		msg := NewMsgEthereumTx(0, nil, tc.amount, 0, tc.gasPrice, nil)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "valid test %d failed: %s", i, tc.msg)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "invalid test %d passed: %s", i, tc.msg)
		}
	}
}

func TestMsgEthereumTxRLPSignBytes(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	chainID := big.NewInt(3)

	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))
	hash := msg.RLPSignBytes(chainID)
	require.Equal(t, "5BD30E35AD27449390B14C91E6BCFDCAADF8FE44EF33680E3BC200FC0DC083C7", fmt.Sprintf("%X", hash))
}

func TestMsgEthereumTxRLPEncode(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	msg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))

	raw, err := rlp.EncodeToBytes(&msg)
	require.NoError(t, err)
	require.Equal(t, ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080"), raw)
}

func TestMsgEthereumTxRLPDecode(t *testing.T) {
	var msg MsgEthereumTx

	raw := ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080")
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	expectedMsg := NewMsgEthereumTx(0, &addr, nil, 100000, nil, []byte("test"))

	err := rlp.Decode(bytes.NewReader(raw), &msg)
	require.NoError(t, err)
	require.Equal(t, expectedMsg.Data, msg.Data)
}

func TestMsgEthereumTxSig(t *testing.T) {
	chainID := big.NewInt(3)

	priv1, _ := crypto.GenerateKey()
	priv2, _ := crypto.GenerateKey()
	addr1 := ethcmn.BytesToAddress(priv1.PubKey().Address().Bytes())
	addr2 := ethcmn.BytesToAddress(priv2.PubKey().Address().Bytes())

	// require valid signature passes validation
	msg := NewMsgEthereumTx(0, &addr1, nil, 100000, nil, []byte("test"))
	err := msg.Sign(chainID, priv1.ToECDSA())
	require.Nil(t, err)

	signer, err := msg.VerifySig(chainID)
	require.NoError(t, err)
	require.Equal(t, addr1, signer)
	require.NotEqual(t, addr2, signer)

	// require invalid chain ID fail validation
	msg = NewMsgEthereumTx(0, &addr1, nil, 100000, nil, []byte("test"))
	err = msg.Sign(chainID, priv1.ToECDSA())
	require.Nil(t, err)

	signer, err = msg.VerifySig(big.NewInt(4))
	require.Error(t, err)
	require.Equal(t, ethcmn.Address{}, signer)
}

func TestMarshalAndUnmarshalLogs(t *testing.T) {
	var cdc = codec.New()

	logs := []*ethtypes.Log{
		{
			Address: ethcmn.BytesToAddress([]byte{0x11}),
			TxHash:  ethcmn.HexToHash("0x01"),
			// May need to find workaround since Topics is required to unmarshal from JSON
			Topics:  []ethcmn.Hash{},
			Removed: true,
		},
		{Address: ethcmn.BytesToAddress([]byte{0x01, 0x11}), Topics: []ethcmn.Hash{}},
	}

	raw, err := codec.MarshalJSONIndent(cdc, logs)
	require.NoError(t, err)

	var logs2 []*ethtypes.Log
	err = cdc.UnmarshalJSON(raw, &logs2)
	require.NoError(t, err)

	require.Len(t, logs2, 2)
	require.Equal(t, logs[0].Address, logs2[0].Address)
	require.Equal(t, logs[0].TxHash, logs2[0].TxHash)
	require.True(t, logs[0].Removed)

	emptyLogs := []*ethtypes.Log{}

	raw, err = codec.MarshalJSONIndent(cdc, emptyLogs)
	require.NoError(t, err)

	err = cdc.UnmarshalJSON(raw, &logs2)
	require.NoError(t, err)
}

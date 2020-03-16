package types

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/ethermint/crypto"
	"github.com/cosmos/ethermint/utils"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
)

func TestMsgEthereumTx(t *testing.T) {
	addr := GenerateEthAddress()

	msg1 := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg1)
	require.Equal(t, *msg1.Data.Recipient, addr)

	msg2 := NewEthereumTxMsgContract(0, nil, 100000, nil, []byte("test"))
	require.NotNil(t, msg2)
	require.Nil(t, msg2.Data.Recipient)

	msg3 := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))
	require.Equal(t, msg3.Route(), RouteEthereumTxMsg)
	require.Equal(t, msg3.Type(), TypeEthereumTxMsg)
	require.Panics(t, func() { msg3.GetSigners() })
	require.Panics(t, func() { msg3.GetSignBytes() })
}

func TestMsgEthereumTxValidation(t *testing.T) {
	testCases := []struct {
		payload    []byte
		amount     *big.Int
		gasPrice   *big.Int
		gasLimit   uint64
		nonce      uint64
		to         ethcmn.Address
		expectPass bool
	}{
		{amount: big.NewInt(100), gasPrice: big.NewInt(100000), expectPass: true},
		{amount: big.NewInt(-1), gasPrice: big.NewInt(100000), expectPass: false},
		{amount: big.NewInt(100), gasPrice: big.NewInt(-1), expectPass: false},
	}

	for i, tc := range testCases {
		msg := NewEthereumTxMsg(tc.nonce, &tc.to, tc.amount, tc.gasLimit, tc.gasPrice, tc.payload)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestMsgEthereumTxRLPSignBytes(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	chainID := big.NewInt(3)

	msg := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))
	hash := msg.RLPSignBytes(chainID)
	require.Equal(t, "5BD30E35AD27449390B14C91E6BCFDCAADF8FE44EF33680E3BC200FC0DC083C7", fmt.Sprintf("%X", hash))
}

func TestMsgEthereumTxRLPEncode(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	msg := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))

	raw, err := rlp.EncodeToBytes(msg)
	require.NoError(t, err)
	require.Equal(t, ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080"), raw)
}

func TestMsgEthereumTxRLPDecode(t *testing.T) {
	var msg EthereumTxMsg

	raw := ethcmn.FromHex("E48080830186A0940000000000000000746573745F61646472657373808474657374808080")
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	expectedMsg := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))

	err := rlp.Decode(bytes.NewReader(raw), &msg)
	require.NoError(t, err)
	require.Equal(t, expectedMsg.Data, msg.Data)
}

func TestMsgEthereumTxHash(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	msg := NewEthereumTxMsg(0, &addr, nil, 100000, nil, []byte("test"))

	hash := msg.Hash()
	require.Equal(t, "E2AA2E68E7586AE9700F1D3D643330866B6AC2B6CA4C804F7C85ECB11D0B0B29", fmt.Sprintf("%X", hash))
}

func TestMsgEthereumTxSig(t *testing.T) {
	chainID := big.NewInt(3)

	priv1, _ := crypto.GenerateKey()
	priv2, _ := crypto.GenerateKey()
	addr1 := ethcmn.BytesToAddress(priv1.PubKey().Address().Bytes())
	addr2 := ethcmn.BytesToAddress(priv2.PubKey().Address().Bytes())

	// require valid signature passes validation
	msg := NewEthereumTxMsg(0, &addr1, nil, 100000, nil, []byte("test"))
	msg.Sign(chainID, priv1.ToECDSA())

	signer, err := msg.VerifySig(chainID)
	require.NoError(t, err)
	require.Equal(t, addr1, signer)
	require.NotEqual(t, addr2, signer)

	// require invalid chain ID fail validation
	msg = NewEthereumTxMsg(0, &addr1, nil, 100000, nil, []byte("test"))
	msg.Sign(chainID, priv1.ToECDSA())

	signer, err = msg.VerifySig(big.NewInt(4))
	require.Error(t, err)
	require.Equal(t, ethcmn.Address{}, signer)
}

func TestMsgEthereumTxAmino(t *testing.T) {
	addr := GenerateEthAddress()
	msg := NewEthereumTxMsg(5, &addr, big.NewInt(1), 100000, big.NewInt(3), []byte("test"))

	msg.Data.V = big.NewInt(1)
	msg.Data.R = big.NewInt(2)
	msg.Data.S = big.NewInt(3)

	raw, err := ModuleCdc.MarshalBinaryBare(msg)
	require.NoError(t, err)

	var msg2 EthereumTxMsg

	err = ModuleCdc.UnmarshalBinaryBare(raw, &msg2)
	require.NoError(t, err)
	require.Equal(t, msg.Data, msg2.Data)
}

func TestMarshalAndUnmarshalInt(t *testing.T) {
	i := big.NewInt(3)
	m := utils.MarshalBigInt(i)
	i2, err := utils.UnmarshalBigInt(m)
	require.NoError(t, err)

	require.Equal(t, i, i2)
}

func TestMarshalAndUnmarshalData(t *testing.T) {
	addr := GenerateEthAddress()
	hash := ethcmn.BigToHash(big.NewInt(2))
	e := EncodableTxData{
		AccountNonce: 2,
		Price:        utils.MarshalBigInt(big.NewInt(3)),
		GasLimit:     1,
		Recipient:    &addr,
		Amount:       utils.MarshalBigInt(big.NewInt(4)),
		Payload:      []byte("test"),

		V: utils.MarshalBigInt(big.NewInt(5)),
		R: utils.MarshalBigInt(big.NewInt(6)),
		S: utils.MarshalBigInt(big.NewInt(7)),

		Hash: &hash,
	}
	str, err := marshalAmino(e)
	require.NoError(t, err)

	e2 := new(EncodableTxData)

	err = unmarshalAmino(e2, str)
	require.NoError(t, err)
	require.Equal(t, e, *e2)
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

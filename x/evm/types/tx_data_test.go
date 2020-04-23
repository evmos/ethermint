package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

func TestMarshalAndUnmarshalData(t *testing.T) {
	addr := GenerateEthAddress()
	hash := ethcmn.BigToHash(big.NewInt(2))

	txData := TxData{
		AccountNonce: 2,
		Price:        big.NewInt(3),
		GasLimit:     1,
		Recipient:    &addr,
		Amount:       big.NewInt(4),
		Payload:      []byte("test"),
		V:            big.NewInt(5),
		R:            big.NewInt(6),
		S:            big.NewInt(7),
		Hash:         &hash,
	}

	bz, err := txData.MarshalAmino()
	require.NoError(t, err)
	require.NotNil(t, bz)

	var txData2 TxData
	err = txData2.UnmarshalAmino(bz)
	require.NoError(t, err)

	require.Equal(t, txData, txData2)
}

func TestMsgEthereumTxAmino(t *testing.T) {
	addr := GenerateEthAddress()
	msg := NewMsgEthereumTx(5, &addr, big.NewInt(1), 100000, big.NewInt(3), []byte("test"))

	msg.Data.V = big.NewInt(1)
	msg.Data.R = big.NewInt(2)
	msg.Data.S = big.NewInt(3)

	raw, err := ModuleCdc.MarshalBinaryBare(msg)
	require.NoError(t, err)

	var msg2 MsgEthereumTx

	err = ModuleCdc.UnmarshalBinaryBare(raw, &msg2)
	require.NoError(t, err)
	require.Equal(t, msg, msg2)
}

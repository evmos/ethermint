package types

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestMsgEthereumTx(t *testing.T) {
	addr := GenerateEthAddress()

	msg := NewMsgEthereumTx(nil, 0, &addr, nil, 100000, nil, []byte("test"), nil)
	require.NotNil(t, msg)
	require.EqualValues(t, msg.Data.To, addr.Bytes())
	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), TypeMsgEthereumTx)
	require.NotNil(t, msg.To())
	require.Equal(t, msg.GetMsgs(), []sdk.Msg{msg})
	require.Panics(t, func() { msg.GetSigners() })
	require.Panics(t, func() { msg.GetSignBytes() })

	msg = NewMsgEthereumTxContract(nil, 0, nil, 100000, nil, []byte("test"), nil)
	require.NotNil(t, msg)
	require.Nil(t, msg.Data.To)
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
		msg := NewMsgEthereumTx(nil, 0, nil, tc.amount, 0, tc.gasPrice, nil, nil)

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

	msg := NewMsgEthereumTx(chainID, 0, &addr, nil, 100000, nil, []byte("test"), nil)
	hash := msg.RLPSignBytes(chainID)
	require.Equal(t, "5BD30E35AD27449390B14C91E6BCFDCAADF8FE44EF33680E3BC200FC0DC083C7", fmt.Sprintf("%X", hash))
}

func TestMsgEthereumTxRLPEncode(t *testing.T) {
	addr := ethcmn.BytesToAddress([]byte("test_address"))
	expMsg := NewMsgEthereumTx(big.NewInt(1), 0, &addr, nil, 100000, nil, []byte("test"), nil)

	raw, err := rlp.EncodeToBytes(&expMsg)
	require.NoError(t, err)

	msg := &MsgEthereumTx{}
	err = rlp.Decode(bytes.NewReader(raw), &msg)
	require.NoError(t, err)
	require.Equal(t, expMsg.Data, msg.Data)
}

// func TestMsgEthereumTxSig(t *testing.T) {
// 	chainID := big.NewInt(3)

// 	priv1, _ := ethsecp256k1.GenerateKey()
// 	priv2, _ := ethsecp256k1.GenerateKey()
// 	addr1 := ethcmn.BytesToAddress(priv1.PubKey().Address().Bytes())
// 	addr2 := ethcmn.BytesToAddress(priv2.PubKey().Address().Bytes())

// 	// require valid signature passes validation
// 	msg := NewMsgEthereumTx(nil, 0, &addr1, nil, 100000, nil, []byte("test"), nil)
// 	err := msg.Sign(chainID, priv1.ToECDSA())
// 	require.Nil(t, err)

// 	signer, err := msg.VerifySig(chainID)
// 	require.NoError(t, err)
// 	require.Equal(t, addr1, signer)
// 	require.NotEqual(t, addr2, signer)

// 	// require invalid chain ID fail validation
// 	msg = NewMsgEthereumTx(nil, 0, &addr1, nil, 100000, nil, []byte("test"), nil)
// 	err = msg.Sign(chainID, priv1.ToECDSA())
// 	require.Nil(t, err)

// 	signer, err = msg.VerifySig(big.NewInt(4))
// 	require.Error(t, err)
// 	require.Equal(t, ethcmn.Address{}, signer)
// }

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestEmintMsg(t *testing.T) {
	addr := newSdkAddress()
	fromAddr := newSdkAddress()

	msg := NewEmintMsg(0, &addr, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), fromAddr)
	require.NotNil(t, msg)
	require.Equal(t, msg.Recipient, &addr)

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), TypeEmintMsg)
}

func TestEmintMsgValidation(t *testing.T) {
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
	}

	for i, tc := range testCases {
		msg := NewEmintMsg(tc.nonce, tc.to, tc.amount, tc.gasLimit, tc.gasPrice, tc.payload, tc.from)

		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
		}
	}
}

func TestEmintEncodingAndDecoding(t *testing.T) {
	addr := newSdkAddress()
	fromAddr := newSdkAddress()

	msg := NewEmintMsg(0, &addr, sdk.NewInt(1), 100000, sdk.NewInt(2), []byte("test"), fromAddr)

	raw, err := cdc.MarshalBinaryBare(msg)
	require.NoError(t, err)

	var msg2 EmintMsg
	err = cdc.UnmarshalBinaryBare(raw, &msg2)
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

package encoding

import (
	"testing"

	"github.com/stretchr/testify/require"

	emintcrypto "github.com/cosmos/ethermint/crypto"
)

func TestKeyEncodingDecoding(t *testing.T) {
	// Priv Key encoding and decoding
	privKey, err := emintcrypto.GenerateKey()
	require.NoError(t, err)
	privBytes := privKey.Bytes()

	decodedPriv, err := PrivKeyFromBytes(privBytes)
	require.NoError(t, err)
	require.Equal(t, privKey, decodedPriv)

	// Pub key encoding and decoding
	pubKey := privKey.PubKey()
	pubBytes := pubKey.Bytes()

	decodedPub, err := PubKeyFromBytes(pubBytes)
	require.NoError(t, err)
	require.Equal(t, pubKey, decodedPub)
}

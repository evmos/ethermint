package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
	secp256k1 "github.com/tendermint/btcd/btcec"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestPrivKeyToSecp256k1(t *testing.T) {
	// require valid SECP256k1 key to convert
	secp256k1PrivKey := tmsecp256k1.GenPrivKey()
	convertedPriv, err := PrivKeyToSecp256k1(secp256k1PrivKey)
	require.NoError(t, err)
	require.Equal(t, secp256k1PrivKey[:], (*secp256k1.PrivateKey)(convertedPriv).Serialize())

	// require invalid ED25519 key not to convert
	ed25519PrivKey := tmed25519.GenPrivKey()
	convertedPriv, err = PrivKeyToSecp256k1(ed25519PrivKey)
	require.Error(t, err)
	require.Nil(t, convertedPriv)
}

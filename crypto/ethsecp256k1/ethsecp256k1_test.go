package ethsecp256k1

import (
	"encoding/base64"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func TestEthermintKeyAndCosmosKey(t *testing.T) {
	ethermintPrivKey, err := GenerateKey()
	require.NoError(t, err)
	require.True(t, ethermintPrivKey.Equals(ethermintPrivKey))
	require.Implements(t, (*cryptotypes.PrivKey)(nil), ethermintPrivKey)

	cosmosPrivKey := new(cosmossecp256k1.PrivKey)
	cosmosPrivKey.Key = ethermintPrivKey.GetKey()

	// Basically, cosmosKey and ethermintKey are same.
	cosmosPubKey := cosmosPrivKey.PubKey()
	ethermintPubKey := ethermintPrivKey.PubKey()
	require.Equal(t, cosmosPubKey.Bytes(), ethermintPubKey.Bytes())

	// Signatures from both keys using secp256k1 sign algorithm must be same.
	msg := []byte("hello world")
	hash := tmcrypto.Sha256(msg)
	sig1, err := secp256k1.Sign(hash, ethermintPrivKey.GetKey())
	require.NoError(t, err)
	sig2, err := secp256k1.Sign(hash, cosmosPrivKey.GetKey())
	require.NoError(t, err)
	require.Equal(t, sig1, sig2)

	// But signature mechanisms of platforms between Ethereum and Cosmos are different.
	// Ethermint signature is 65 bytes(RSV), Cosmos signature is 64 bytes(RS).
	res1, err := ethermintPrivKey.Sign(hash)
	require.NoError(t, err)
	res2, err := cosmosPrivKey.Sign(msg)
	require.NoError(t, err)
	// Signatures are same except v of signature from ethermintPrivKey.
	require.Equal(t, res1[:len(res1)-1], res2)
}

func TestPrivKey(t *testing.T) {
	// validate type and equality
	privKey, err := GenerateKey()
	require.NoError(t, err)
	require.True(t, privKey.Equals(privKey))
	require.Implements(t, (*cryptotypes.PrivKey)(nil), privKey)

	// validate inequality
	privKey2, err := GenerateKey()
	require.NoError(t, err)
	require.False(t, privKey.Equals(privKey2))

	// validate Ethereum address equality
	addr := privKey.PubKey().Address()
	key, err := privKey.ToECDSA()
	require.NoError(t, err)
	expectedAddr := crypto.PubkeyToAddress(key.PublicKey)
	require.Equal(t, expectedAddr.Bytes(), addr.Bytes())

	// validate we can sign some bytes
	msg := []byte("hello world")
	sigHash := crypto.Keccak256Hash(msg)
	expectedSig, err := secp256k1.Sign(sigHash.Bytes(), privKey.Bytes())
	require.NoError(t, err)

	sig, err := privKey.Sign(sigHash.Bytes())
	require.NoError(t, err)
	require.Equal(t, expectedSig, sig)
}

func TestPrivKey_PubKey(t *testing.T) {
	privKey, err := GenerateKey()
	require.NoError(t, err)

	// validate type and equality
	pubKey := &PubKey{
		Key: privKey.PubKey().Bytes(),
	}
	require.Implements(t, (*cryptotypes.PubKey)(nil), pubKey)

	// validate inequality
	privKey2, err := GenerateKey()
	require.NoError(t, err)
	require.False(t, pubKey.Equals(privKey2.PubKey()))

	// validate signature
	msg := []byte("hello world")
	sigHash := crypto.Keccak256Hash(msg)
	sig, err := privKey.Sign(sigHash.Bytes())
	require.NoError(t, err)

	res := pubKey.VerifySignature(msg, sig)
	require.True(t, res)
}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey, err := GenerateKey()
	require.NoError(t, err)

	pubKey := privKey.PubKey().(*PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"ethsecp256k1 private key",
			privKey,
			&PrivKey{},
			append([]byte{32}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"ethsecp256k1 public key",
			pubKey,
			&PubKey{},
			append([]byte{33}, pubKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(pubKey.Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.Marshal(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			err = aminoCdc.Unmarshal(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)

			// Do a round trip of encoding/decoding JSON.
			bz, err = aminoCdc.MarshalJSON(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expJSON, string(bz))

			err = aminoCdc.UnmarshalJSON(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)
		})
	}
}

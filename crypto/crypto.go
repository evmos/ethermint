package crypto

import (
	"crypto/ecdsa"
	"fmt"

	secp256k1 "github.com/tendermint/btcd/btcec"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

// PrivKeyToSecp256k1 accepts a Tendermint private key and attempts to convert
// it to a SECP256k1 ecdsa.PrivateKey.
func PrivKeyToSecp256k1(priv tmcrypto.PrivKey) (*ecdsa.PrivateKey, error) {
	secp256k1Key, ok := priv.(tmsecp256k1.PrivKeySecp256k1)
	if !ok {
		return nil, fmt.Errorf("invalid private key type: %T", priv)
	}

	ecdsaPrivKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), secp256k1Key[:])
	return ecdsaPrivKey.ToECDSA(), nil
}

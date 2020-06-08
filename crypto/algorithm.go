package crypto

import (
	"crypto/hmac"
	"crypto/sha512"

	"github.com/tyler-smith/go-bip39"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// EthSecp256k1Type uses the Ethereum secp256k1 ECDSA parameters.
const EthSecp256k1Type = hd.PubKeyType("ethsecp256k1")

var _ keyring.SignatureAlgo = ethSecp256k1{}

// EthSeckp256k1Option defines a keyring option for the ethereum Secp256k1 curve.
func EthSeckp256k1Option(options *keyring.Options) {
	options.SupportedAlgos = append(options.SupportedAlgos, Secp256k1)
	options.SupportedAlgosLedger = append(options.SupportedAlgosLedger, Secp256k1)
}

// Secp256k1 represents the Secp256k1 curve used in Ethereum.
var Secp256k1 = ethSecp256k1{}

type ethSecp256k1 struct{}

// Name returns the Secp256k1 PubKeyType.
func (s ethSecp256k1) Name() hd.PubKeyType {
	return EthSecp256k1Type
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s ethSecp256k1) Derive() hd.DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		// HMAC the seed to produce the private key and chain code
		mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
		_, err = mac.Write(seed)
		if err != nil {
			return nil, err
		}

		seed = mac.Sum(nil)

		priv, err := ethcrypto.ToECDSA(seed[:32])
		if err != nil {
			return nil, err
		}

		derivedKey := PrivKeySecp256k1(ethcrypto.FromECDSA(priv))

		return derivedKey, nil
	}
}

func (ethSecp256k1) Generate() hd.GenerateFn {
	return func(bz []byte) tmcrypto.PrivKey {
		var bzArr [32]byte
		copy(bzArr[:], bz)
		return PrivKeySecp256k1(bzArr[:])
	}
}

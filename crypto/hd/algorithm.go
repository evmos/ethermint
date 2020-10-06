package hd

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/tyler-smith/go-bip39"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
)

const (
	// EthSecp256k1 defines the ECDSA secp256k1 used on Ethereum
	EthSecp256k1 = keys.SigningAlgo(ethsecp256k1.KeyType)
)

// SupportedAlgorithms defines the list of signing algorithms used on Ethermint:
//  - eth_secp256k1 (Ethereum)
//  - secp256k1 (Tendermint)
var SupportedAlgorithms = []keys.SigningAlgo{EthSecp256k1, keys.Secp256k1}

// EthSecp256k1Options defines a keys options for the ethereum Secp256k1 curve.
func EthSecp256k1Options() []keys.KeybaseOption {
	return []keys.KeybaseOption{
		keys.WithKeygenFunc(EthermintKeygenFunc),
		keys.WithDeriveFunc(DeriveKey),
		keys.WithSupportedAlgos(SupportedAlgorithms),
		keys.WithSupportedAlgosLedger(SupportedAlgorithms),
	}
}

func DeriveKey(mnemonic, bip39Passphrase, hdPath string, algo keys.SigningAlgo) ([]byte, error) {
	switch algo {
	case keys.Secp256k1:
		return keys.StdDeriveKey(mnemonic, bip39Passphrase, hdPath, algo)
	case EthSecp256k1:
		return DeriveSecp256k1(mnemonic, bip39Passphrase, hdPath)
	default:
		return nil, errors.Wrap(keys.ErrUnsupportedSigningAlgo, string(algo))
	}
}

// EthermintKeygenFunc is the key generation function to generate secp256k1 ToECDSA
// from ethereum.
func EthermintKeygenFunc(bz []byte, algo keys.SigningAlgo) (tmcrypto.PrivKey, error) {
	if algo != EthSecp256k1 {
		return nil, fmt.Errorf("signing algorithm must be %s, got %s", EthSecp256k1, algo)
	}

	return ethsecp256k1.PrivKey(bz), nil
}

// DeriveSecp256k1 derives and returns the eth_secp256k1 private key for the given mnemonic and HD path.
func DeriveSecp256k1(mnemonic, bip39Passphrase, path string) ([]byte, error) {
	hdpath, err := ethaccounts.ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, err
	}

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	key := masterKey
	for _, n := range hdpath {
		key, err = key.Child(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}

	privateKeyECDSA := privateKey.ToECDSA()
	derivedKey := ethsecp256k1.PrivKey(ethcrypto.FromECDSA(privateKeyECDSA))
	return derivedKey, nil
}

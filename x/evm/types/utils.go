package types

import (
	"crypto/ecdsa"
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethsha "github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/pkg/errors"
)

// PrivKeyToEthAddress generates an Ethereum address given an ECDSA private key.
func PrivKeyToEthAddress(p *ecdsa.PrivateKey) ethcmn.Address {
	return ethcrypto.PubkeyToAddress(p.PublicKey)
}

// GenerateAddress generates an Ethereum address.
func GenerateEthAddress() ethcmn.Address {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	return PrivKeyToEthAddress(priv)
}

// ValidateSigner attempts to validate a signer for a given slice of bytes over
// which a signature and signer is given. An error is returned if address
// derived from the signature and bytes signed does not match the given signer.
func ValidateSigner(signBytes, sig []byte, signer ethcmn.Address) error {
	pk, err := ethcrypto.SigToPub(signBytes, sig)

	if err != nil {
		return errors.Wrap(err, "failed to derive public key from signature")
	} else if ethcrypto.PubkeyToAddress(*pk) != signer {
		return fmt.Errorf("invalid signature for signer: %s", signer)
	}

	return nil
}

func rlpHash(x interface{}) (hash ethcmn.Hash) {
	hasher := ethsha.NewKeccak256()

	rlp.Encode(hasher, x)
	hasher.Sum(hash[:0])

	return
}

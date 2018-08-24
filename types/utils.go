package types

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/pkg/errors"
)

// GenerateAddress generates an Ethereum address.
func GenerateEthAddress() ethcmn.Address {
	priv, err := ethcrypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	return PrivKeyToEthAddress(priv)
}

// PrivKeyToEthAddress generates an Ethereum address given an ECDSA private key.
func PrivKeyToEthAddress(p *ecdsa.PrivateKey) ethcmn.Address {
	return ethcrypto.PubkeyToAddress(ecdsa.PublicKey(p.PublicKey))
}

// ValidateSigner attempts to validate a signer for a given slice of bytes over
// which a signature and signer is given. An error is returned if address
// derived from the signature and bytes signed does not match the given signer.
func ValidateSigner(signBytes, sig []byte, signer ethcmn.Address) error {
	pk, err := ethcrypto.SigToPub(signBytes, sig)

	if err != nil {
		return errors.Wrap(err, "signature verification failed")
	} else if ethcrypto.PubkeyToAddress(*pk) != signer {
		return fmt.Errorf("invalid signature for signer: %s", signer)
	}

	return nil
}

// GetStdTxSignBytes returns the signature bytes for an auth.StdTx transaction
// that is compatible with Ethereum's signature mechanism.
func GetStdTxSignBytes(chainID string, accNum int64, seq int64, fee auth.StdFee, msgs []sdk.Msg, memo string) []byte {
	signBytes := auth.StdSignBytes(chainID, accNum, seq, fee, msgs, "")
	hash := sha256.Sum256(signBytes)
	return hash[:]
}

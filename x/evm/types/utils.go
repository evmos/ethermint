package types

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

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
	hasher := sha3.NewLegacyKeccak256()
	_ = rlp.Encode(hasher, x)
	_ = hasher.Sum(hash[:0])

	return hash
}

// EncodeTxResponse takes all of the necessary data from the EVM execution
// and returns the data as a byte slice encoded with protobuf.
func EncodeTxResponse(res *MsgEthereumTxResponse) ([]byte, error) {
	return proto.Marshal(res)
}

// DecodeTxResponse decodes an protobuf-encoded byte slice into TxResponse
func DecodeTxResponse(data []byte) (MsgEthereumTxResponse, error) {
	var txResponse MsgEthereumTxResponse
	err := proto.Unmarshal(data, &txResponse)
	if err != nil {
		return MsgEthereumTxResponse{}, err
	}
	return txResponse, nil
}

// ----------------------------------------------------------------------------
// Auxiliary

// recoverEthSig recovers a signature according to the Ethereum specification and
// returns the sender or an error.
//
// Ref: Ethereum Yellow Paper (BYZANTIUM VERSION 69351d5) Appendix F
// nolint: gocritic
func recoverEthSig(R, S, Vb *big.Int, sigHash ethcmn.Hash) (ethcmn.Address, error) {
	if Vb.BitLen() > 8 {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	V := byte(Vb.Uint64() - 27)
	if !ethcrypto.ValidateSignatureValues(V, R, S, true) {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V

	// recover the public key from the signature
	pub, err := ethcrypto.Ecrecover(sigHash[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}

	if len(pub) == 0 || pub[0] != 4 {
		return ethcmn.Address{}, errors.New("invalid public key")
	}

	var addr ethcmn.Address
	copy(addr[:], ethcrypto.Keccak256(pub[1:])[12:])

	return addr, nil
}

// IsEmptyHash returns true if the hash corresponds to an empty ethereum hex hash.
func IsEmptyHash(hash string) bool {
	return bytes.Equal(ethcmn.HexToHash(hash).Bytes(), ethcmn.Hash{}.Bytes())
}

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(ethcmn.HexToAddress(address).Bytes(), ethcmn.Address{}.Bytes())
}

package types

import (
	"bytes"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// IsEmptyHash returns true if the hash corresponds to an empty ethereum hex hash.
func IsEmptyHash(hash string) bool {
	return bytes.Equal(ethcmn.HexToHash(hash).Bytes(), ethcmn.Hash{}.Bytes())
}

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(ethcmn.HexToAddress(address).Bytes(), ethcmn.Address{}.Bytes())
}

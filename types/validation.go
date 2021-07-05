package types

import (
	"bytes"

	ethcmn "github.com/ethereum/go-ethereum/common"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// IsEmptyHash returns true if the hash corresponds to an empty ethereum hex hash.
func IsEmptyHash(hash string) bool {
	return bytes.Equal(ethcmn.HexToHash(hash).Bytes(), ethcmn.Hash{}.Bytes())
}

// IsZeroAddress returns true if the address corresponds to an empty ethereum hex address.
func IsZeroAddress(address string) bool {
	return bytes.Equal(ethcmn.HexToAddress(address).Bytes(), ethcmn.Address{}.Bytes())
}

// ValidateAddress returns an error if the provided string is either not a hex formatted string address
func ValidateAddress(address string) error {
	if !ethcmn.IsHexAddress(address) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidAddress, "address '%s' is not a valid ethereum hex address",
			address,
		)
	}
	return nil
}

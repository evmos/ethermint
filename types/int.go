package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// PowerReduction defines the default power reduction value for staking
var PowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

// MarshalBigInt marshals big int into text string for consistent encoding
func MarshalBigInt(i *big.Int) (string, error) {
	bz, err := i.MarshalText()
	if err != nil {
		return "", sdkerrors.Wrap(ErrMarshalBigInt, err.Error())
	}
	return string(bz), nil
}

// UnmarshalBigInt unmarshals string from *big.Int
func UnmarshalBigInt(s string) (*big.Int, error) {
	ret := new(big.Int)
	err := ret.UnmarshalText([]byte(s))
	if err != nil {
		return nil, sdkerrors.Wrap(ErrUnmarshalBigInt, err.Error())
	}
	return ret, nil
}

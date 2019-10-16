package utils

import "math/big"

// MarshalBigInt marshalls big int into text string for consistent encoding
func MarshalBigInt(i *big.Int) string {
	bz, err := i.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// UnmarshalBigInt unmarshalls string from *big.Int
func UnmarshalBigInt(s string) (*big.Int, error) {
	ret := new(big.Int)
	err := ret.UnmarshalText([]byte(s))
	return ret, err
}

// MustUnmarshalBigInt unmarshalls string from *big.Int
func MustUnmarshalBigInt(s string) *big.Int {
	ret := new(big.Int)
	err := ret.UnmarshalText([]byte(s))
	if err != nil {
		panic(err)
	}
	return ret
}

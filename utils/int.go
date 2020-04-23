package utils

import "math/big"

// MarshalBigInt marshalls big int into text string for consistent encoding
func MarshalBigInt(i *big.Int) (string, error) {
	bz, err := i.MarshalText()
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

// MustMarshalBigInt marshalls big int into text string for consistent encoding.
// It panics if an error is encountered.
func MustMarshalBigInt(i *big.Int) string {
	str, err := MarshalBigInt(i)
	if err != nil {
		panic(err)
	}
	return str
}

// UnmarshalBigInt unmarshalls string from *big.Int
func UnmarshalBigInt(s string) (*big.Int, error) {
	ret := new(big.Int)
	err := ret.UnmarshalText([]byte(s))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// MustUnmarshalBigInt unmarshalls string from *big.Int.
// It panics if an error is encountered.
func MustUnmarshalBigInt(s string) *big.Int {
	ret, err := UnmarshalBigInt(s)
	if err != nil {
		panic(err)
	}
	return ret
}

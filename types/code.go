package types

import (
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// ----------------------------------------------------------------------------
// Code & Storage
// ----------------------------------------------------------------------------

type (
	// Code is account Code type alias
	Code []byte
	// Storage is account storage type alias
	Storage map[ethcmn.Hash]ethcmn.Hash
)

func (c Code) String() string {
	return string(c)
}

func (c Storage) String() (str string) {
	for key, value := range c {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

// Copy returns a copy of storage.
func (c Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range c {
		cpy[key] = value
	}

	return cpy
}

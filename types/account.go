package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var _ auth.Account = (*Account)(nil)

const (
	// DenomDefault defines the single coin type/denomination supported in
	// Ethermint.
	DenomDefault = "photon"
)

// ----------------------------------------------------------------------------
// Main Ethermint account
// ----------------------------------------------------------------------------

// Account implements the auth.Account interface and embeds an
// auth.BaseAccount type. It is compatible with the auth.AccountMapper.
type Account struct {
	*auth.BaseAccount

	// merkle root of the storage trie
	//
	// TODO: good chance we may not need this
	Root ethcmn.Hash

	CodeHash []byte
}

// ProtoBaseAccount defines the prototype function for BaseAccount used for an
// account mapper.
func ProtoBaseAccount() auth.Account {
	return &Account{BaseAccount: &auth.BaseAccount{}}
}

// Balance returns the balance of an account.
func (acc Account) Balance() sdk.Int {
	return acc.GetCoins().AmountOf(DenomDefault)
}

// SetBalance sets an account's balance.
func (acc Account) SetBalance(amt sdk.Int) {
	// nolint:errcheck
	acc.SetCoins(sdk.Coins{sdk.NewCoin(DenomDefault, amt)})
}

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

package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var _ exported.Account = (*Account)(nil)
var _ exported.GenesisAccount = (*Account)(nil)

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
	// TODO: add back root if needed (marshalling is broken if not initializing)
	// Root ethcmn.Hash

	CodeHash []byte
}

// ProtoBaseAccount defines the prototype function for BaseAccount used for an
// account mapper.
func ProtoBaseAccount() exported.Account {
	return &Account{BaseAccount: &auth.BaseAccount{}}
}

// Balance returns the balance of an account.
func (acc Account) Balance() sdk.Int {
	return acc.GetCoins().AmountOf(DenomDefault)
}

// SetBalance sets an account's balance of photons
func (acc Account) SetBalance(amt sdk.Int) {
	coins := acc.GetCoins()
	diff := amt.Sub(coins.AmountOf(DenomDefault))
	if diff.IsZero() {
		return
	} else if diff.IsPositive() {
		// Increase coins to amount
		coins = coins.Add(sdk.Coins{sdk.NewCoin(DenomDefault, diff)})
	} else {
		// Decrease coins to amount
		coins = coins.Sub(sdk.Coins{sdk.NewCoin(DenomDefault, diff.Neg())})
	}
	if err := acc.SetCoins(coins); err != nil {
		panic(fmt.Sprintf("Could not set coins for address %s", acc.GetAddress()))
	}
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

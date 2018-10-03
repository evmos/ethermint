package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/ethermint/x/bank"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var _ auth.Account = (*Account)(nil)

// BaseAccount implements the auth.Account interface and embeds an
// auth.BaseAccount type. It is compatible with the auth.AccountMapper.
type Account struct {
	*auth.BaseAccount

	Root     ethcmn.Hash // merkle root of the storage trie
	CodeHash []byte
}

// ProtoBaseAccount defines the prototype function for BaseAccount used for an
// account mapper.
func ProtoBaseAccount() auth.Account {
	return &Account{BaseAccount: &auth.BaseAccount{}}
}

// Balance returns the balance of an account.
func (acc Account) Balance() sdk.Int {
	return acc.GetCoins().AmountOf(bank.DenomEthereum)
}

// SetBalance sets an account's balance.
func (acc Account) SetBalance(amt sdk.Int) {
	acc.SetCoins(sdk.Coins{sdk.NewCoin(bank.DenomEthereum, amt)})
}

// // NewAccount returns a reference to a new initialized account.
// func NewAccount(base auth.BaseAccount, code []byte) *Account {
// 	return &Account{
// 		BaseAccount: base,
// 		Code:        code,
// 		Storage:     storage,
// 	}
// }

// GetAccountDecoder returns the auth.AccountDecoder function for the custom
// Account type.
// func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
// 	return func(accBytes []byte) (auth.Account, error) {
// 		if len(accBytes) == 0 {
// 			return nil, sdk.ErrTxDecode("account bytes are empty")
// 		}

// 		acc := new(Account)

// 		err := cdc.UnmarshalBinaryBare(accBytes, &acc)
// 		if err != nil {
// 			return nil, sdk.ErrTxDecode("failed to decode account bytes")
// 		}

// 		return acc, err
// 	}
// }

// Account code and storage type aliases.
type (
	Code    []byte
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

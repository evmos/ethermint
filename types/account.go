package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var _ auth.Account = (*Account)(nil)

type (
	// Storage defines account storage
	Storage map[ethcmn.Hash]ethcmn.Hash

	// Account defines an auth.BaseAccount extension for Ethermint. It is
	// compatible with the auth.AccountMapper.
	Account struct {
		auth.BaseAccount

		Code    []byte
		Storage Storage
	}
)

// NewAccount returns a reference to a new initialized account.
func NewAccount(base auth.BaseAccount, code []byte, storage Storage) *Account {
	return &Account{
		BaseAccount: base,
		Code:        code,
		Storage:     storage,
	}
}

// GetAccountDecoder returns the auth.AccountDecoder function for the custom
// Account type.
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (auth.Account, error) {
		if len(accBytes) == 0 {
			return nil, sdk.ErrTxDecode("account bytes are empty")
		}

		acc := new(Account)

		err := cdc.UnmarshalBinaryBare(accBytes, &acc)
		if err != nil {
			return nil, sdk.ErrTxDecode("failed to decode account bytes")
		}

		return acc, err
	}
}

package types

import (
	"bytes"
	"errors"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var zeroAddrBytes = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

type (
	// GenesisState defines the application's genesis state. It contains all the
	// information required and accounts to initialize the blockchain.
	GenesisState struct {
		Accounts []GenesisAccount `json:"accounts"`
	}

	// GenesisStorage represents the GenesisAccount Storage map as single key value
	// pairs. This is to prevent non determinism at genesis initialization or export.
	GenesisStorage struct {
		Key   ethcmn.Hash `json:"key"`
		Value ethcmn.Hash `json:"value"`
	}

	// GenesisAccount defines an account to be initialized in the genesis state.
	// Its main difference between with Geth's GenesisAccount is that it uses a custom
	// storage type and that it doesn't contain the private key field.
	GenesisAccount struct {
		Address ethcmn.Address   `json:"address"`
		Balance *big.Int         `json:"balance"`
		Code    []byte           `json:"code,omitempty"`
		Storage []GenesisStorage `json:"storage,omitempty"`
	}
)

// NewGenesisStorage creates a new GenesisStorage instance
func NewGenesisStorage(key, value ethcmn.Hash) GenesisStorage {
	return GenesisStorage{
		Key:   key,
		Value: value,
	}
}

// DefaultGenesisState sets default evm genesis state with empty accounts.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts: []GenesisAccount{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, acc := range gs.Accounts {
		if bytes.Equal(acc.Address.Bytes(), zeroAddrBytes) {
			return errors.New("invalid GenesisAccount: address cannot be empty")
		}
		if acc.Balance == nil {
			return errors.New("invalid GenesisAccount: balance cannot be empty")
		}
	}
	return nil
}

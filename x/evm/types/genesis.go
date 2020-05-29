package types

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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
		Code    hexutil.Bytes    `json:"code,omitempty"`
		Storage []GenesisStorage `json:"storage,omitempty"`
	}
)

// Validate performs a basic validation of a GenesisAccount fields.
func (ga GenesisAccount) Validate() error {
	if bytes.Equal(ga.Address.Bytes(), ethcmn.Address{}.Bytes()) {
		return fmt.Errorf("address cannot be the zero address %s", ga.Address.String())
	}
	if ga.Balance == nil {
		return errors.New("balance cannot be nil")
	}
	if ga.Balance.Sign() == -1 {
		return errors.New("balance cannot be negative")
	}
	if ga.Code != nil && len(ga.Code) == 0 {
		return errors.New("code bytes cannot be empty")
	}

	seenStorage := make(map[string]bool)
	for i, state := range ga.Storage {
		if seenStorage[state.Key.String()] {
			return fmt.Errorf("duplicate state key %d", i)
		}
		if bytes.Equal(state.Key.Bytes(), ethcmn.Hash{}.Bytes()) {
			return fmt.Errorf("state %d key hash cannot be empty", i)
		}
		// NOTE: state value can be empty
		seenStorage[state.Key.String()] = true
	}
	return nil
}

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
	seenAccounts := make(map[string]bool)
	for _, acc := range gs.Accounts {
		if seenAccounts[acc.Address.String()] {
			return fmt.Errorf("duplicated genesis account %s", acc.Address.String())
		}
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid genesis account %s: %w", acc.Address.String(), err)
		}
		seenAccounts[acc.Address.String()] = true
	}
	return nil
}

package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type (
	// GenesisState defines the evm module genesis state
	GenesisState struct {
		Accounts    []GenesisAccount  `json:"accounts"`
		TxsLogs     []TransactionLogs `json:"txs_logs"`
		ChainConfig ChainConfig       `json:"chain_config"`
		Params      Params            `json:"params"`
	}

	// GenesisAccount defines an account to be initialized in the genesis state.
	// Its main difference between with Geth's GenesisAccount is that it uses a custom
	// storage type and that it doesn't contain the private key field.
	GenesisAccount struct {
		Address string        `json:"address"`
		Balance sdk.Int       `json:"balance"`
		Code    hexutil.Bytes `json:"code,omitempty"`
		Storage Storage       `json:"storage,omitempty"`
	}
)

// Validate performs a basic validation of a GenesisAccount fields.
func (ga GenesisAccount) Validate() error {
	if ga.Address == (ethcmn.Address{}.String()) {
		return fmt.Errorf("address cannot be the zero address %s", ga.Address)
	}
	if ga.Balance.IsNil() {
		return errors.New("balance cannot be nil")
	}
	if ga.Balance.IsNegative() {
		return errors.New("balance cannot be negative")
	}
	if ga.Code != nil && len(ga.Code) == 0 {
		return errors.New("code bytes cannot be empty")
	}

	return ga.Storage.Validate()
}

// DefaultGenesisState sets default evm genesis state with empty accounts and default params and
// chain config values.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts:    []GenesisAccount{},
		TxsLogs:     []TransactionLogs{},
		ChainConfig: DefaultChainConfig(),
		Params:      DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	seenAccounts := make(map[string]bool)
	seenTxs := make(map[string]bool)
	for _, acc := range gs.Accounts {
		if seenAccounts[acc.Address] {
			return fmt.Errorf("duplicated genesis account %s", acc.Address)
		}
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid genesis account %s: %w", acc.Address, err)
		}
		seenAccounts[acc.Address] = true
	}

	for _, tx := range gs.TxsLogs {
		if seenTxs[tx.Hash.String()] {
			return fmt.Errorf("duplicated logs from transaction %s", tx.Hash.String())
		}

		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid logs from transaction %s: %w", tx.Hash.String(), err)
		}

		seenTxs[tx.Hash.String()] = true
	}

	if err := gs.ChainConfig.Validate(); err != nil {
		return err
	}

	return gs.Params.Validate()
}

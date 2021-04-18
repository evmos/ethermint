package types

import (
	"errors"
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// Validate performs a basic validation of a GenesisAccount fields.
func (ga GenesisAccount) Validate() error {
	if IsZeroAddress(ga.Address) {
		return fmt.Errorf("address cannot be the zero address %s", ga.Address)
	}
	if len(ethcmn.Hex2Bytes(ga.Code)) == 0 {
		return errors.New("code cannot be empty")
	}

	return ga.Storage.Validate()
}

// DefaultGenesisState sets default evm genesis state with empty accounts and default params and
// chain config values.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
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
		if seenTxs[tx.Hash] {
			return fmt.Errorf("duplicated logs from transaction %s", tx.Hash)
		}

		if err := tx.Validate(); err != nil {
			return fmt.Errorf("invalid logs from transaction %s: %w", tx.Hash, err)
		}

		seenTxs[tx.Hash] = true
	}

	if err := gs.ChainConfig.Validate(); err != nil {
		return err
	}

	return gs.Params.Validate()
}

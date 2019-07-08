package evm

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/types"
)

type (
	// GenesisState defines the application's genesis state. It contains all the
	// information required and accounts to initialize the blockchain.
	GenesisState struct {
		Accounts []GenesisAccount `json:"accounts"`
	}

	// GenesisAccount defines an account to be initialized in the genesis state.
	GenesisAccount struct {
		Address sdk.AccAddress `json:"address"`
		Coins   sdk.Coins      `json:"coins"`
		Code    []byte         `json:"code,omitempty"`
		Storage types.Storage  `json:"storage,omitempty"`
	}
)

func ValidateGenesis(data GenesisState) error {
	for _, acct := range data.Accounts {
		if acct.Address == nil {
			return fmt.Errorf("Invalid GenesisAccount Error: Missing Address")
		}
		if acct.Coins == nil {
			return fmt.Errorf("Invalid GenesisAccount Error: Missing Coins")
		}
	}
	return nil
}

func DefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts: []GenesisAccount{},
	}
}

// TODO: Implement these once keeper is established
//func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) []abci.ValidatorUpdate {
//	for _, record := range data.Accounts {
//		// TODO: Add to keeper
//	}
//	return []abci.ValidatorUpdate{}
//}
//
//func ExportGenesis(ctx sdk.Context, k Keeper) GenesisState {
//	return GenesisState{Accounts: nil}
//}

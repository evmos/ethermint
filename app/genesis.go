package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/ethermint/types"
)

type (
	// GenesisState defines the application's genesis state. It contains all the
	// information required and accounts to initialize the blockchain.
	GenesisState struct {
		Accounts  []GenesisAccount   `json:"accounts"`
		StakeData stake.GenesisState `json:"stake"`
		GovData   gov.GenesisState   `json:"gov"`
	}

	// GenesisAccount defines an account to be initialized in the genesis state.
	GenesisAccount struct {
		Address sdk.AccAddress `json:"address"`
		Coins   sdk.Coins      `json:"coins"`
		Code    []byte         `json:"code,omitempty"`
		Storage types.Storage  `json:"storage,omitempty"`
	}
)

// NewGenesisAccount returns a reference to a new initialized genesis account.
func NewGenesisAccount(acc *types.Account) GenesisAccount {
	return GenesisAccount{
		Address: acc.GetAddress(),
		Coins:   acc.GetCoins(),
		Code:    acc.Code,
		Storage: acc.Storage,
	}
}

// ToAccount converts a genesis account to an initialized Ethermint account.
func (ga *GenesisAccount) ToAccount() (acc *types.Account) {
	base := auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}

	return types.NewAccount(base, ga.Code, ga.Storage)
}

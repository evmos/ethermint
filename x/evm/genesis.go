package evm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) []abci.ValidatorUpdate {
	for _, record := range data.Accounts {
		k.SetCode(ctx, record.Address, record.Code)
		k.CreateGenesisAccount(ctx, record)
	}
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state
func ExportGenesis(ctx sdk.Context, _ Keeper) GenesisState {
	return GenesisState{Accounts: nil}
}

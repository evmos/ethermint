package faucet

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/x/faucet/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// InitGenesis initializes genesis state based on exported genesis
func InitGenesis(ctx sdk.Context, k Keeper, data types.GenesisState) []abci.ValidatorUpdate {
	if acc := k.GetFaucetAccount(ctx); acc == nil {
		panic(fmt.Sprintf("%s module account has not been set", ModuleName))
	}

	k.SetEnabled(ctx, data.EnableFaucet)
	k.SetTimout(ctx, data.Timeout)
	k.SetCap(ctx, data.FaucetCap)
	k.SetMaxPerRequest(ctx, data.MaxAmountPerRequest)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports genesis state
func ExportGenesis(ctx sdk.Context, k Keeper) types.GenesisState {
	return types.GenesisState{
		EnableFaucet:        k.IsEnabled(ctx),
		Timeout:             k.GetTimeout(ctx),
		FaucetCap:           k.GetCap(ctx),
		MaxAmountPerRequest: k.GetMaxPerRequest(ctx),
	}
}

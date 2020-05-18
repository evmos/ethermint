package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// SupplyKeeper is required for mining coin
type SupplyKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(
		ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
	) error
	GetModuleAccount(ctx sdk.Context, moduleName string) supplyexported.ModuleAccountI
}

// StakingKeeper is required for getting Denom
type StakingKeeper interface {
	BondDenom(ctx sdk.Context) string
}

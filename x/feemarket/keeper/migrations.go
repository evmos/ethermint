package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v0_12 "github.com/tharsis/ethermint/x/feemarket/v0_12"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{
		keeper: keeper,
	}
}

// Migrate1to2 migrates the store from consensus version v1 to v2
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v0_12.MigrateStore(ctx, m.keeper, m.keeper.storeKey)
}

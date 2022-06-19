package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v2 "github.com/evmos/ethermint/x/evm/migrations/v2"
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
	return v2.MigrateStore(ctx, &m.keeper.paramSpace)
}

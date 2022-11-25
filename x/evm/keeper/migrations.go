package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4 "github.com/evmos/ethermint/x/evm/migrations/v4"
	"github.com/evmos/ethermint/x/evm/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace types.Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper, legacySubspace types.Subspace) Migrator {
	return Migrator{
		keeper:         keeper,
		legacySubspace: legacySubspace,
	}
}

// TODO: Figure out if these will be deleted
// Migrate1to2 migrates the store from consensus version v1 to v2
// NOTE: This migration handler is no longer valid as it's missing deprecated Cosmos SDK params module
// func (m Migrator) Migrate1to2(ctx sdk.Context) error {
//	return v2.MigrateStore(ctx, &m.keeper.paramSpace)
// }
// TODO: Figure out if these will be deleted
// Migrate2to3 migrates the store from consensus version v2 to v3
// NOTE: This migration handler is no longer valid as it's missing deprecated Cosmos SDK params module
// func (m Migrator) Migrate2to3(ctx sdk.Context) error {
//	return v3.MigrateStore(ctx, &m.keeper.paramSpace)
// }

// Migrate3to4 migrates the store from consensus version v3 to v4
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	return v4.MigrateStore(ctx, ctx.KVStore(m.keeper.storeKey), m.legacySubspace, m.keeper.cdc)
}

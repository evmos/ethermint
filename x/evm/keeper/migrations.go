package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/x/evm/types"
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
	paramstore := m.keeper.paramSpace
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	// add RejectUnprotected
	paramstore.Set(ctx, types.ParamStoreKeyRejectUnprotected, types.DefaultParams().RejectUnprotected)
	return nil
}

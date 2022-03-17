package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	store := ctx.KVStore(m.keeper.storeKey)
	baseFeeKeyPrefix := []byte{2}
	bz := store.Get(baseFeeKeyPrefix)
	if len(bz) > 0 {
		baseFee := new(big.Int).SetBytes(bz)
		m.keeper.SetBaseFee(ctx, baseFee)
	}
	store.Delete(baseFeeKeyPrefix)
	return nil
}

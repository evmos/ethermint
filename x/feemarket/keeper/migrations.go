package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeyPrefixBaseFeeV1 is the base fee key prefix used in version 1
var KeyPrefixBaseFeeV1 = []byte{2}

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
	bz := store.Get(KeyPrefixBaseFeeV1)
	if len(bz) > 0 {
		baseFee := new(big.Int).SetBytes(bz)
		m.keeper.SetBaseFee(ctx, baseFee)
	}
	store.Delete(KeyPrefixBaseFeeV1)
	return nil
}

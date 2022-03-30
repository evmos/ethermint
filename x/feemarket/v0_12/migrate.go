package v0_12

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type FeeMarketKeeper interface {
	SetBaseFee(sdk.Context, *big.Int)
}

// KeyPrefixBaseFeeV1 is the base fee key prefix used in version 1
var KeyPrefixBaseFeeV1 = []byte{2}

// MigrateStore migrates the BaseFee value from the store to the params for
// In-Place Store migration logic.
func MigrateStore(ctx sdk.Context, k FeeMarketKeeper, storeKey sdk.StoreKey) error {
	store := ctx.KVStore(storeKey)
	bz := store.Get(KeyPrefixBaseFeeV1)
	if len(bz) > 0 {
		baseFee := new(big.Int).SetBytes(bz)
		k.SetBaseFee(ctx, baseFee)
	}
	store.Delete(KeyPrefixBaseFeeV1)
	return nil
}

// TODO: update
func MigrateJSON(ctx sdk.Context, k FeeMarketKeeper, storeKey sdk.StoreKey) error {
	return nil
}

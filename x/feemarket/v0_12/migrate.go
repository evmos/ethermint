package v0_12

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tharsis/ethermint/x/feemarket/types"
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

// MigrateJSON accepts exported v0.11 x/feemarket genesis state and migrates it to
// v0.12 x/feemarket genesis state. The migration includes:
// - Migrate BaseFee to Params
func MigrateJSON(oldState types.GenesisState) types.GenesisState {
	return types.GenesisState{
		Params: types.Params{
			NoBaseFee:                oldState.Params.NoBaseFee,
			BaseFeeChangeDenominator: oldState.Params.BaseFeeChangeDenominator,
			ElasticityMultiplier:     oldState.Params.ElasticityMultiplier,
			EnableHeight:             oldState.Params.EnableHeight,
			// BaseFee:                  oldState.BaseFee, FIXME: import
		},
		BlockGas: oldState.BlockGas,
	}
}

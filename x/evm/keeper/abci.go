package keeper

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

// BeginBlock sets the block hash -> block height map for the previous block height
// and resets the Bloom filter and the transaction count to 0.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.WithContext(ctx)
	k.WithChainID(ctx)
}

// EndBlock updates the accounts and commits state objects to the KV Store, while
// deleting the empty ones. It also sets the bloom filers for the request block to
// the store. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	k.WithContext(infCtx)

	k.SetBlockBloom(infCtx, req.Height, ethtypes.BytesToBloom(k.GetBlockBloomTransient().Bytes()))
	k.WithContext(ctx)

	return []abci.ValidatorUpdate{}
}

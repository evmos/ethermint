package keeper

import (
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

// BeginBlock sets the sdk Context and EIP155 chain id to the Keeper.
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	k.WithContext(ctx)
	k.WithChainID(ctx)
}

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	// Gas costs are handled within msg handler so costs should be ignored
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	k.WithContext(infCtx)

	baseFee := k.CalculateBaseFee(ctx)

	// only set base fee if the NoBaseFee param is false
	if baseFee != nil {
		k.SetBaseFee(ctx, baseFee)
		k.SetBlockGasUsed(ctx)

		k.Ctx().EventManager().EmitEvent(sdk.NewEvent(
			"block_gas",
			sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("amount", fmt.Sprintf("%d", ctx.BlockGasMeter().GasConsumedToLimit())),
		))
	}

	bloom := ethtypes.BytesToBloom(k.GetBlockBloomTransient().Bytes())
	k.SetBlockBloom(infCtx, req.Height, bloom)

	k.WithContext(ctx)

	return []abci.ValidatorUpdate{}
}

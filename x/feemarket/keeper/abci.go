package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlock updates base fee
func (k *Keeper) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	baseFee := k.CalculateBaseFee(ctx)

	// return immediately if base fee is nil
	if baseFee == nil {
		return
	}

	k.SetBaseFee(ctx, baseFee)

	defer func() {
		telemetry.SetGauge(float32(baseFee.Int64()), "feemarket", "base_fee")
	}()

	// Store current base fee in event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFeeMarket,
			sdk.NewAttribute(types.AttributeKeyBaseFee, baseFee.String()),
		),
	})
}

// EndBlock update block gas used.
// The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) {
	if ctx.BlockGasMeter() == nil {
		k.Logger(ctx).Error("block gas meter is nil when setting block gas used")
		return
	}

	gasUsed := ctx.BlockGasMeter().GasConsumedToLimit()

	k.SetBlockGasUsed(ctx, gasUsed)

	defer func() {
		telemetry.SetGauge(float32(gasUsed), "feemarket", "block_gas")
	}()

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"block_gas",
		sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
		sdk.NewAttribute("amount", fmt.Sprintf("%d", ctx.BlockGasMeter().GasConsumedToLimit())),
	))
}

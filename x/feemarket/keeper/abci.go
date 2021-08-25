package keeper

// EndBlock also retrieves the bloom filter value from the transient store and commits it to the
// KVStore. The EVM end block logic doesn't update the validator set, thus it returns
// an empty slice.
func (k *Keeper) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	baseFee := k.CalculateBaseFee(ctx)
	if baseFee == nil {
		return
	}

	// only set base fee if the NoBaseFee param is false
	k.SetBaseFee(ctx, baseFee)
	k.SetBlockGasUsed(ctx)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"block_gas",
		sdk.NewAttribute("height", fmt.Sprintf("%d", ctx.BlockHeight())),
		sdk.NewAttribute("amount", fmt.Sprintf("%d", ctx.BlockGasMeter().GasConsumedToLimit())),
	))
}
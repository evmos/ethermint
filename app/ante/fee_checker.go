package ante

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	ethermint "github.com/evmos/ethermint/types"
	"github.com/evmos/ethermint/x/evm/types"
)

// NewDynamicFeeChecker returns a `TxFeeChecker` that applies a dynamic fee to
// Cosmos txs using the EIP-1559 fee market logic.
// This can be called in both CheckTx and deliverTx modes.
// a) feeCap = tx.fees / tx.gas
// b) tipFeeCap = tx.MaxPriorityPrice (default) or MaxInt64
// - when `ExtensionOptionDynamicFeeTx` is omitted, `tipFeeCap` defaults to `MaxInt64`.
// - when london hardfork is not enabled, it fallbacks to SDK default behavior (validator min-gas-prices).
// - Tx priority is set to `effectiveGasPrice / DefaultPriorityReduction`.
func NewDynamicFeeChecker(k DynamicFeeEVMKeeper) authante.TxFeeChecker {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
		feeTx, ok := tx.(sdk.FeeTx)
		if !ok {
			return nil, 0, fmt.Errorf("tx must be a FeeTx")
		}

		if ctx.BlockHeight() == 0 {
			// genesis transactions: fallback to min-gas-price logic
			return checkTxFeeWithValidatorMinGasPrices(ctx, feeTx)
		}

		params := k.GetParams(ctx)
		denom := params.EvmDenom
		ethCfg := params.ChainConfig.EthereumConfig(k.ChainID())

		baseFee := k.GetBaseFee(ctx, ethCfg)
		if baseFee == nil {
			// london hardfork is not enabled: fallback to min-gas-prices logic
			return checkTxFeeWithValidatorMinGasPrices(ctx, feeTx)
		}

		// default to `MaxInt64` when there's no extension option.
		maxPriorityPrice := sdkmath.NewInt(math.MaxInt64)

		// get the priority tip cap from the extension option.
		if hasExtOptsTx, ok := tx.(authante.HasExtensionOptionsTx); ok {
			for _, opt := range hasExtOptsTx.GetExtensionOptions() {
				if extOpt, ok := opt.GetCachedValue().(*ethermint.ExtensionOptionDynamicFeeTx); ok {
					maxPriorityPrice = extOpt.MaxPriorityPrice
					break
				}
			}
		}

		gas := feeTx.GetGas()
		feeCoins := feeTx.GetFee()
		fee := feeCoins.AmountOfNoDenomValidation(denom)

		feeCap := fee.Quo(sdkmath.NewIntFromUint64(gas))
		baseFeeInt := sdkmath.NewIntFromBigInt(baseFee)

		if feeCap.LT(baseFeeInt) {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient gas prices; got: %s required: %s", feeCap, baseFeeInt)
		}

		// calculate the effective gas price using the EIP-1559 logic.
		effectivePrice := sdkmath.NewIntFromBigInt(types.EffectiveGasPrice(baseFeeInt.BigInt(), feeCap.BigInt(), maxPriorityPrice.BigInt()))

		// NOTE: create a new coins slice without having to validate the denom
		effectiveFee := sdk.Coins{
			{
				Denom:  denom,
				Amount: effectivePrice.Mul(sdkmath.NewIntFromUint64(gas)),
			},
		}

		bigPriority := effectivePrice.Sub(baseFeeInt).Quo(types.DefaultPriorityReduction)
		priority := int64(math.MaxInt64)

		if bigPriority.IsInt64() {
			priority = bigPriority.Int64()
		}

		return effectiveFee, priority, nil
	}
}

// checkTxFeeWithValidatorMinGasPrices implements the default fee logic, where the minimum price per
// unit of gas is fixed and set by each validator, and the tx priority is computed from the gas price.
func checkTxFeeWithValidatorMinGasPrices(ctx sdk.Context, tx sdk.FeeTx) (sdk.Coins, int64, error) {
	feeCoins := tx.GetFee()
	gas := tx.GetGas()
	minGasPrices := ctx.MinGasPrices()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))

		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
		}
	}

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
func getTxPriority(fees sdk.Coins, gas int64) int64 {
	var priority int64

	for _, fee := range fees {
		gasPrice := fee.Amount.QuoRaw(gas)
		amt := gasPrice.Quo(types.DefaultPriorityReduction)
		p := int64(math.MaxInt64)

		if amt.IsInt64() {
			p = amt.Int64()
		}

		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}

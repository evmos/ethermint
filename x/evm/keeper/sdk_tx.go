package keeper

import (
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/evmos/ethermint/x/evm/types"
)

// CheckSDKTxFee implements the `TxFeeChecker` for ante handler,
// could be called in both checkTx and deliverTx modes.
func (k Keeper) CheckSDKTxFee(ctx sdk.Context, tx sdk.Tx) (effectiveFee sdk.Coins, priority int64, err error) {
	params := k.GetParams(ctx)
	denom := params.EvmDenom
	ethCfg := params.ChainConfig.EthereumConfig(k.ChainID())
	baseFee := k.GetBaseFee(ctx, ethCfg)

	if baseFee == nil {
		// fallback to default sdk logic if london hardfork is not enabled
		return authante.CheckTxFeeWithValidatorMinGasPrices(ctx, tx)
	}

	// default to `0` when there's no extension option.
	prioPriceCap := sdkmath.NewInt(0)
	if hasExtOptsTx, ok := tx.(authante.HasExtensionOptionsTx); ok {
		for _, opt := range hasExtOptsTx.GetExtensionOptions() {
			if extOpt, ok := opt.GetCachedValue().(*types.ExtensionOptionDynamicFeeTx); ok {
				prioPriceCap = extOpt.MaxPriorityPrice
				break
			}
		}
	}
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, fmt.Errorf("Tx must be a FeeTx")
	}

	gas := feeTx.GetGas()
	feeCoins := feeTx.GetFee()
	fee := feeCoins.AmountOfNoDenomValidation(denom)
	priceCap := fee.Quo(sdkmath.NewIntFromUint64(gas))

	basePrice := sdkmath.NewIntFromBigInt(baseFee)
	if priceCap.LT(basePrice) {
		return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient gas price; got: %s required: %s", priceCap, basePrice)
	}

	effectivePrice := sdkmath.NewIntFromBigInt(types.EffectiveGasPrice(basePrice.BigInt(), priceCap.BigInt(), prioPriceCap.BigInt()))
	effectiveFee = sdk.NewCoins(sdk.NewCoin(denom, effectivePrice.Mul(sdkmath.NewIntFromUint64(gas))))
	bigPriority := effectivePrice.Sub(basePrice).Quo(DefaultPriorityReduction)
	if !bigPriority.IsInt64() {
		priority = math.MaxInt64
	} else {
		priority = bigPriority.Int64()
	}
	return
}

package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"math/big"
)

// GasWantedDecorator keeps track of the gasWanted amount on the current block in transient store
// for BaseFee calculation.
// NOTE: This decorator does not perform any validation
type GasWantedDecorator struct {
	evmKeeper EVMKeeper
	feeMaker  FeeMarketKeeper
}

// NewGasWantedDecorator creates a new NewGasWantedDecorator
func NewGasWantedDecorator(
	evmKeeper EVMKeeper,
	feeMarket FeeMarketKeeper,
) GasWantedDecorator {
	return GasWantedDecorator{
		evmKeeper,
		feeMarket,
	}
}

func (gwd GasWantedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	params := gwd.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(gwd.evmKeeper.ChainID())

	blockHeight := big.NewInt(ctx.BlockHeight())
	london := ethCfg.IsLondon(blockHeight)

	feeTx, ok := tx.(sdk.FeeTx)
	if ok {
		gasWanted := feeTx.GetGas()

		// Add total gasWanted to cumulative in block transientStore in FeeMarket module
		if london && !gwd.feeMaker.GetParams(ctx).NoBaseFee {
			if _, err := gwd.feeMaker.AddTransientGasWanted(ctx, gasWanted); err != nil {
				return ctx, sdkerrors.Wrapf(err, "failed to add gas wanted to transient store")
			}
		}
	}

	return next(ctx, tx, simulate)
}

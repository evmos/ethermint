package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// MinGasPriceDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies for both CheckTx and DeliverTx
// If fee is high enough, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MinGasPriceDecorator
type MinGasPriceDecorator struct {
	feesKeeper FeeMarketKeeper
	evmKeeper  EVMKeeper
}

func NewMinGasPriceDecorator(fk FeeMarketKeeper, ek EVMKeeper) MinGasPriceDecorator {
	return MinGasPriceDecorator{feesKeeper: fk, evmKeeper: ek}
}

func (mpd MinGasPriceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	minGasPrice := mpd.feesKeeper.GetParams(ctx).MinGasPrice
	minGasPrices := sdk.DecCoins{sdk.DecCoin{
		Denom:  mpd.evmKeeper.GetParams(ctx).EvmDenom,
		Amount: minGasPrice,
	}}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		gasLimit := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(gasLimit)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !feeCoins.IsAnyGTE(requiredFees) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "provided fee < minimum global fee (%s < %s). Please increase the gas price.", feeCoins, requiredFees)
		}
	}

	return next(ctx, tx, simulate)
}

// EthMinGasPriceDecorator will check if the transaction's fee is at least as large
// as the MinGasPrices param. If fee is too low, decorator returns error and tx
// is rejected. This applies to both CheckTx and DeliverTx and regardless
// if London hard fork or fee market params (EIP-1559) are enabled.
// If fee is high enough, then call next AnteHandler
type EthMinGasPriceDecorator struct {
	feesKeeper FeeMarketKeeper
	evmKeeper  EVMKeeper
}

func NewEthMinGasPriceDecorator(fk FeeMarketKeeper, ek EVMKeeper) EthMinGasPriceDecorator {
	return EthMinGasPriceDecorator{feesKeeper: fk, evmKeeper: ek}
}

func (empd EthMinGasPriceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	minGasPrice := empd.feesKeeper.GetParams(ctx).MinGasPrice

	if !minGasPrice.IsZero() {
		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
			}

			feeAmt := ethMsg.GetFee()

			// For dynamic transactions, GetFee() uses the GasFeeCap value, which
			// is the maximum gas price that the signer can pay. In practice, the
			// signer can pay less, if the block's BaseFee is lower. So, in this case,
			// we use the EffectiveFee. If the feemarket formula results in a BaseFee
			// that lowers EffectivePrice until it is < MinGasPrices, the users must
			// increase the GasTipCap (priority fee) until EffectivePrice > MinGasPrices.
			// Transactions with MinGasPrices * gasUsed < tx fees < EffectiveFee are rejected
			// by the feemarket AnteHandle
			txData, err := evmtypes.UnpackTxData(ethMsg.Data)
			if err != nil {
				return ctx, sdkerrors.Wrapf(err, "failed to unpack tx data %s", ethMsg.Hash)
			}
			if txData.TxType() != ethtypes.LegacyTxType {
				paramsEvm := empd.evmKeeper.GetParams(ctx)
				ethCfg := paramsEvm.ChainConfig.EthereumConfig(empd.evmKeeper.ChainID())
				baseFee := empd.evmKeeper.GetBaseFee(ctx, ethCfg)
				feeAmt = ethMsg.GetEffectiveFee(baseFee)
			}

			gasLimit := sdk.NewDec(int64(ethMsg.GetGas()))
			requiredFee := minGasPrice.Mul(gasLimit)

			if sdk.NewDecFromBigInt(feeAmt).LT(requiredFee) {
				return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "provided fee < minimum global fee (%s < %s). Please increase the priority tip (for EIP-1559 txs) or the gas prices (for access list or legacy txs)", feeAmt, requiredFee)
			}
		}
	}

	return next(ctx, tx, simulate)
}

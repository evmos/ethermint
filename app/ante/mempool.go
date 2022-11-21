package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// EthMempoolFeeDecorator will check if the transaction's effective fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator
type EthMempoolFeeDecorator struct {
	evmKeeper EVMKeeper
}

func NewEthMempoolFeeDecorator(ek EVMKeeper) EthMempoolFeeDecorator {
	return EthMempoolFeeDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle ensures that the provided fees meet a minimum threshold for the validator,
// if this is a CheckTx. This is only for local mempool purposes, and thus
// is only ran on check tx.
// It only do the check if london hardfork not enabled or feemarket not enabled, because in that case feemarket will take over the task.
func (mfd EthMempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.IsCheckTx() && !simulate {
		chainCfg := mfd.evmKeeper.GetChainConfig(ctx)
		ethCfg := chainCfg.EthereumConfig(mfd.evmKeeper.ChainID())
		baseFee := mfd.evmKeeper.GetBaseFee(ctx, ethCfg)
		evmDenom := mfd.evmKeeper.GetEVMDenom(ctx)

		if baseFee == nil {
			for _, msg := range tx.GetMsgs() {
				ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
				if !ok {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid message type %T, expected %T", msg, (*evmtypes.MsgEthereumTx)(nil))
				}

				feeAmt := ethMsg.GetFee()
				glDec := sdk.NewDec(int64(ethMsg.GetGas()))
				requiredFee := ctx.MinGasPrices().AmountOf(evmDenom).Mul(glDec)
				if sdk.NewDecFromBigInt(feeAmt).LT(requiredFee) {
					return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeAmt, requiredFee)
				}
			}
		}
	}

	return next(ctx, tx, simulate)
}

package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/palantir/stacktrace"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	cmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// DeductTxCostsFromUserBalance it calculates the tx costs and deducts the fees
func (k Keeper) DeductTxCostsFromUserBalance(
	ctx sdk.Context,
	msgEthTx evmtypes.MsgEthereumTx,
	txData evmtypes.TxData,
	denom string,
	homestead, istanbul, london bool,
) (sdk.Coins, error) {
	isContractCreation := txData.GetTo() == nil

	// fetch sender account from signature
	signerAcc, err := authante.GetSignerAcc(ctx, k.accountKeeper, msgEthTx.GetFrom())
	if err != nil {
		return nil, stacktrace.Propagate(err, "account not found for sender %s", msgEthTx.From)
	}

	gasLimit := txData.GetGas()

	var accessList ethtypes.AccessList
	if txData.GetAccessList() != nil {
		accessList = txData.GetAccessList()
	}

	intrinsicGas, err := core.IntrinsicGas(txData.GetData(), accessList, isContractCreation, homestead, istanbul)
	if err != nil {
		return nil, stacktrace.Propagate(sdkerrors.Wrap(
			err,
			"failed to compute intrinsic gas cost"), "failed to retrieve intrinsic gas, contract creation = %t; homestead = %t, istanbul = %t",
			isContractCreation, homestead, istanbul,
		)
	}

	// intrinsic gas verification during CheckTx
	if ctx.IsCheckTx() && gasLimit < intrinsicGas {
		return nil, sdkerrors.Wrapf(
			sdkerrors.ErrOutOfGas,
			"gas limit too low: %d (gas limit) < %d (intrinsic gas)", gasLimit, intrinsicGas,
		)
	}

	// calculate the fees paid to validators based on the effective tip and price
	effectiveTip := txData.GetGasPrice()

	feeMktParams := k.feeMarketKeeper.GetParams(ctx)

	if london && !feeMktParams.NoBaseFee && txData.TxType() == ethtypes.DynamicFeeTxType {
		baseFee := k.feeMarketKeeper.GetBaseFee(ctx)
		effectiveTip = cmath.BigMin(txData.GetGasTipCap(), new(big.Int).Sub(txData.GetGasFeeCap(), baseFee))
	}

	gasUsed := new(big.Int).SetUint64(txData.GetGas())
	feeAmt := new(big.Int).Mul(gasUsed, effectiveTip)

	fees := sdk.Coins{sdk.NewCoin(denom, sdk.NewIntFromBigInt(feeAmt))}

	// deduct the full gas cost from the user balance
	if err := authante.DeductFees(k.bankKeeper, ctx, signerAcc, fees); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"failed to deduct full gas cost %s from the user %s balance",
			fees, msgEthTx.From,
		)
	}
	return fees, nil
}

// CheckSenderBalance validates that the tx cost value is positive and that the
// sender has enough funds to pay for the fees and value of the transaction.
func CheckSenderBalance(
	ctx sdk.Context,
	bankKeeper evmtypes.BankKeeper,
	sender sdk.AccAddress,
	txData evmtypes.TxData,
	denom string,
) error {
	balance := bankKeeper.GetBalance(ctx, sender, denom)
	cost := txData.Cost()

	if cost.Sign() < 0 {
		return stacktrace.Propagate(
			sdkerrors.Wrapf(
				sdkerrors.ErrInvalidCoins,
				"tx cost (%s%s) is negative and invalid", cost, denom,
			),
			"tx cost amount should never be negative")
	}

	if balance.IsNegative() || balance.Amount.BigInt().Cmp(cost) < 0 {
		return stacktrace.Propagate(
			sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"sender balance < tx cost (%s < %s%s)", balance, txData.Cost(), denom,
			),
			"sender should have had enough funds to pay for tx cost = fee + amount (%s = %s + %s)",
			cost, txData.Fee(), txData.GetValue(),
		)
	}
	return nil
}

package keeper

import (
	"math"
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// DefaultPriorityReduction is the default amount of price values required for 1 unit of priority.
// Because priority is `int64` while price is `big.Int`, it's necessary to scale down the range to keep it more pratical.
// The default value is the same as the `sdk.DefaultPowerReduction`.
var DefaultPriorityReduction = sdk.DefaultPowerReduction

// DeductTxCostsFromUserBalance it calculates the tx costs and deducts the fees
// returns (effectiveFee, priority, error)
func (k Keeper) DeductTxCostsFromUserBalance(
	ctx sdk.Context,
	msgEthTx evmtypes.MsgEthereumTx,
	txData evmtypes.TxData,
	denom string,
	homestead, istanbul, london bool,
) (fees sdk.Coins, priority int64, err error) {
	isContractCreation := txData.GetTo() == nil

	// fetch sender account from signature
	signerAcc, err := authante.GetSignerAcc(ctx, k.accountKeeper, msgEthTx.GetFrom())
	if err != nil {
		return nil, 0, sdkerrors.Wrapf(err, "account not found for sender %s", msgEthTx.From)
	}

	gasLimit := txData.GetGas()

	var accessList ethtypes.AccessList
	if txData.GetAccessList() != nil {
		accessList = txData.GetAccessList()
	}

	intrinsicGas, err := core.IntrinsicGas(txData.GetData(), accessList, isContractCreation, homestead, istanbul)
	if err != nil {
		return nil, 0, sdkerrors.Wrapf(
			err,
			"failed to retrieve intrinsic gas, contract creation = %t; homestead = %t, istanbul = %t",
			isContractCreation, homestead, istanbul,
		)
	}

	// intrinsic gas verification during CheckTx
	if ctx.IsCheckTx() && gasLimit < intrinsicGas {
		return nil, 0, sdkerrors.Wrapf(
			sdkerrors.ErrOutOfGas,
			"gas limit too low: %d (gas limit) < %d (intrinsic gas)", gasLimit, intrinsicGas,
		)
	}

	var feeAmt *big.Int

	baseFee := k.getBaseFee(ctx, london)
	if baseFee != nil && txData.GetGasFeeCap().Cmp(baseFee) < 0 {
		return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "the tx gasfeecap is lower than the tx baseFee: %s (gasfeecap), %s (basefee) ", txData.GetGasFeeCap(), baseFee)
	}

	feeAmt = txData.EffectiveFee(baseFee)
	if feeAmt.Sign() == 0 {
		// zero fee, no need to deduct
		return sdk.Coins{}, 0, nil
	}

	fees = sdk.Coins{sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(feeAmt))}

	// deduct the full gas cost from the user balance
	if err := authante.DeductFees(k.bankKeeper, ctx, signerAcc, fees); err != nil {
		return nil, 0, sdkerrors.Wrapf(
			err,
			"failed to deduct full gas cost %s from the user %s balance",
			fees, msgEthTx.From,
		)
	}

	// calculate priority based on effective gas price
	tipPrice := txData.EffectiveGasPrice(baseFee)
	// if london hardfork is not enabled, tipPrice is the gasPrice
	if baseFee != nil {
		tipPrice = new(big.Int).Sub(tipPrice, baseFee)
	}
	priorityBig := new(big.Int).Quo(tipPrice, DefaultPriorityReduction.BigInt())
	if !priorityBig.IsInt64() {
		priority = math.MaxInt64
	} else {
		priority = priorityBig.Int64()
	}

	return fees, priority, nil
}

// CheckSenderBalance validates that the tx cost value is positive and that the
// sender has enough funds to pay for the fees and value of the transaction.
func CheckSenderBalance(
	balance sdkmath.Int,
	txData evmtypes.TxData,
) error {
	cost := txData.Cost()

	if cost.Sign() < 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidCoins,
			"tx cost (%s) is negative and invalid", cost,
		)
	}

	if balance.IsNegative() || balance.BigInt().Cmp(cost) < 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"sender balance < tx cost (%s < %s)", balance, txData.Cost(),
		)
	}
	return nil
}

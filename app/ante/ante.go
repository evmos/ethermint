package ante

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/crypto"
	emint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	ethcore "github.com/ethereum/go-ethereum/core"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

const (
	// TODO: Use this cost per byte through parameter or overriding NewConsumeGasForTxSizeDecorator
	// which currently defaults at 10, if intended
	// memoCostPerByte     sdk.Gas = 3
	secp256k1VerifyCost uint64 = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(ak auth.AccountKeeper, sk types.SupplyKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {

		switch castTx := tx.(type) {
		case auth.StdTx:
			stdAnte := sdk.ChainAnteDecorators(
				authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
				authante.NewMempoolFeeDecorator(),
				authante.NewValidateBasicDecorator(),
				authante.NewValidateMemoDecorator(ak),
				authante.NewConsumeGasForTxSizeDecorator(ak),
				authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
				authante.NewValidateSigCountDecorator(ak),
				authante.NewDeductFeeDecorator(ak, sk),
				authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
				authante.NewSigVerificationDecorator(ak),
				authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
			)

			return stdAnte(ctx, tx, sim)

		case evmtypes.MsgEthereumTx:
			return ethAnteHandler(ctx, ak, sk, &castTx, sim)

		default:
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}
	}
}

// sigGasConsumer overrides the DefaultSigVerificationGasConsumer from the x/auth
// module on the SDK. It doesn't allow ed25519 nor multisig thresholds.
func sigGasConsumer(
	meter sdk.GasMeter, sig []byte, pubkey tmcrypto.PubKey, params types.Params,
) error {
	switch pubkey.(type) {
	case crypto.PubKeySecp256k1:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
		return nil
	case tmcrypto.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: tendermint secp256k1")
		return nil
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

// ----------------------------------------------------------------------------
// Ethereum Ante Handler

// ethAnteHandler defines an internal ante handler for an Ethereum transaction
// ethTxMsg. During CheckTx, the transaction is passed through a series of
// pre-message execution validation checks such as signature and account
// verification in addition to minimum fees being checked. Otherwise, during
// DeliverTx, the transaction is simply passed to the EVM which will also
// perform the same series of checks. The distinction is made in CheckTx to
// prevent spam and DoS attacks.
func ethAnteHandler(
	ctx sdk.Context, ak auth.AccountKeeper, sk types.SupplyKeeper,
	ethTxMsg *evmtypes.MsgEthereumTx, sim bool,
) (newCtx sdk.Context, err error) {

	var senderAddr sdk.AccAddress

	// This is done to ignore costs in Ante handler checks
	ctx = ctx.WithBlockGasMeter(sdk.NewInfiniteGasMeter())

	if ctx.IsCheckTx() {
		// Only perform pre-message (Ethereum transaction) execution validation
		// during CheckTx. Otherwise, during DeliverTx the EVM will handle them.
		if senderAddr, err = validateEthTxCheckTx(ctx, ak, ethTxMsg); err != nil {
			return ctx, err
		}
	} else {
		// This is still currently needed to retrieve the sender address
		if senderAddr, err = validateSignature(ctx, ethTxMsg); err != nil {
			return ctx, err
		}

		// Explicit nonce check is also needed in case of multiple txs with same nonce not being handled
		if err := checkNonce(ctx, ak, ethTxMsg, senderAddr); err != nil {
			return ctx, err
		}
	}

	// Recover and catch out of gas error
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				err = sdkerrors.Wrapf(
					sdkerrors.ErrOutOfGas,
					"out of gas in location: %v; gasUsed: %d",
					rType.Descriptor, ctx.GasMeter().GasConsumed(),
				)
			default:
				panic(r)
			}
		}
	}()

	// Fetch sender account from signature
	senderAcc, err := auth.GetSignerAcc(ctx, ak, senderAddr)
	if err != nil {
		return ctx, err
	}

	// Charge sender for gas up to limit
	if ethTxMsg.Data.GasLimit != 0 {
		// Cost calculates the fees paid to validators based on gas limit and price
		cost := new(big.Int).Mul(ethTxMsg.Data.Price, new(big.Int).SetUint64(ethTxMsg.Data.GasLimit))

		feeAmt := sdk.NewCoins(
			sdk.NewCoin(emint.DenomDefault, sdk.NewIntFromBigInt(cost)),
		)

		err = auth.DeductFees(sk, ctx, senderAcc, feeAmt)
		if err != nil {
			return ctx, err
		}
	}

	// Set gas meter after ante handler to ignore gaskv costs
	newCtx = auth.SetGasMeter(sim, ctx, ethTxMsg.Data.GasLimit)

	gas, _ := ethcore.IntrinsicGas(ethTxMsg.Data.Payload, ethTxMsg.To() == nil, true)
	newCtx.GasMeter().ConsumeGas(gas, "eth intrinsic gas")

	// Increment sequence of sender
	acc := ak.GetAccount(ctx, senderAddr)
	if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
		panic(err)
	}
	ak.SetAccount(ctx, acc)

	return newCtx, nil
}

func validateEthTxCheckTx(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.MsgEthereumTx,
) (sdk.AccAddress, error) {
	// Validate sufficient fees have been provided that meet a minimum threshold
	// defined by the proposer (for mempool purposes during CheckTx).
	if err := ensureSufficientMempoolFees(ctx, ethTxMsg); err != nil {
		return nil, err
	}

	// validate enough intrinsic gas
	if err := validateIntrinsicGas(ethTxMsg); err != nil {
		return nil, err
	}

	signer, err := validateSignature(ctx, ethTxMsg)
	if err != nil {
		return nil, err
	}

	// validate account (nonce and balance checks)
	if err := validateAccount(ctx, ak, ethTxMsg, signer); err != nil {
		return nil, err
	}

	return sdk.AccAddress(signer.Bytes()), nil
}

// Validates signature and returns sender address
func validateSignature(ctx sdk.Context, ethTxMsg *evmtypes.MsgEthereumTx) (sdk.AccAddress, error) {
	// parse the chainID from a string to a base-10 integer
	chainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return nil, sdkerrors.Wrap(emint.ErrInvalidChainID, ctx.ChainID())
	}

	// validate sender/signature
	signer, err := ethTxMsg.VerifySig(chainID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "signature verification failed")
	}

	return sdk.AccAddress(signer.Bytes()), nil
}

// validateIntrinsicGas validates that the Ethereum tx message has enough to
// cover intrinsic gas. Intrinsic gas for a transaction is the amount of gas
// that the transaction uses before the transaction is executed. The gas is a
// constant value of 21000 plus any cost inccured by additional bytes of data
// supplied with the transaction.
func validateIntrinsicGas(ethTxMsg *evmtypes.MsgEthereumTx) error {
	gas, err := ethcore.IntrinsicGas(ethTxMsg.Data.Payload, ethTxMsg.To() == nil, true)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to compute intrinsic gas cost")
	}

	if ethTxMsg.Data.GasLimit < gas {
		return fmt.Errorf(
			"intrinsic gas too low: %d < %d", ethTxMsg.Data.GasLimit, gas,
		)
	}

	return nil
}

// validateAccount validates the account nonce and that the account has enough
// funds to cover the tx cost.
func validateAccount(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.MsgEthereumTx, signer sdk.AccAddress,
) error {

	acc := ak.GetAccount(ctx, signer)

	// on InitChain make sure account number == 0
	if ctx.BlockHeight() == 0 && acc.GetAccountNumber() != 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"invalid account number for height zero (got %d)", acc.GetAccountNumber(),
		)
	}

	// Validate nonce is correct
	if err := checkNonce(ctx, ak, ethTxMsg, signer); err != nil {
		return err
	}

	// validate sender has enough funds
	balance := acc.GetCoins().AmountOf(emint.DenomDefault)
	if balance.BigInt().Cmp(ethTxMsg.Cost()) < 0 {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"%s < %s%s", balance.String(), ethTxMsg.Cost().String(), emint.DenomDefault,
		)
	}

	return nil
}

func checkNonce(
	ctx sdk.Context, ak auth.AccountKeeper, ethTxMsg *evmtypes.MsgEthereumTx, signer sdk.AccAddress,
) error {
	acc := ak.GetAccount(ctx, signer)
	// Validate the transaction nonce is valid (equivalent to the sender account’s
	// current nonce).
	seq := acc.GetSequence()
	if ethTxMsg.Data.AccountNonce != seq {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"got nonce %d, expected %d", ethTxMsg.Data.AccountNonce, seq,
		)
	}

	return nil
}

// ensureSufficientMempoolFees verifies that enough fees have been provided by the
// Ethereum transaction that meet the minimum threshold set by the block
// proposer.
//
// NOTE: This should only be ran during a CheckTx mode.
func ensureSufficientMempoolFees(ctx sdk.Context, ethTxMsg *evmtypes.MsgEthereumTx) error {
	// fee = GP * GL
	fee := sdk.NewDecCoinFromCoin(sdk.NewInt64Coin(emint.DenomDefault, ethTxMsg.Fee().Int64()))

	minGasPrices := ctx.MinGasPrices()
	allGTE := true
	for _, v := range minGasPrices {
		if !fee.IsGTE(v) {
			allGTE = false
		}
	}

	// it is assumed that the minimum fees will only include the single valid denom
	if !ctx.MinGasPrices().IsZero() && !allGTE {
		// reject the transaction that does not meet the minimum fee
		return sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFee,
			"got: %q required: %q", fee, ctx.MinGasPrices(),
		)
	}

	return nil
}

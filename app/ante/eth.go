package ante

import (
	"fmt"
	"math/big"

	log "github.com/xlab/suplog"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	sidechain "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
)

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	GetParams(ctx sdk.Context) evmtypes.Params
}

// EthSetupContextDecorator sets the infinite GasMeter in the Context and wraps
// the next AnteHandler with a defer clause to recover from any downstream
// OutOfGas panics in the AnteHandler chain to return an error with information
// on gas provided and gas used.
// CONTRACT: Must be first decorator in the chain
// CONTRACT: Tx must implement GasTx interface
type EthSetupContextDecorator struct{}

// NewEthSetupContextDecorator creates a new EthSetupContextDecorator
func NewEthSetupContextDecorator() EthSetupContextDecorator {
	return EthSetupContextDecorator{}
}

// AnteHandle sets the infinite gas meter to done to ignore costs in AnteHandler checks.
// This is undone at the EthGasConsumeDecorator, where the context is set with the
// ethereum tx GasLimit.
func (escd EthSetupContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// all transactions must implement GasTx
	gasTx, ok := tx.(authante.GasTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be GasTx")
	}

	// Decorator will catch an OutOfGasPanic caused in the next antehandler
	// AnteHandlers must have their own defer/recover in order for the BaseApp
	// to know how much gas was used! This is because the GasMeter is created in
	// the AnteHandler, but if it panics the context won't be set properly in
	// runTx's recover call.
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				log := fmt.Sprintf(
					"out of gas in location: %v; gasLimit: %d, gasUsed: %d",
					rType.Descriptor, gasTx.GetGas(), ctx.GasMeter().GasConsumed(),
				)
				err = sdkerrors.Wrap(sdkerrors.ErrOutOfGas, log)
			default:
				log.Errorln(r)
				panic(r)
			}
		}
	}()

	return next(ctx, tx, simulate)
}

// EthMempoolFeeDecorator validates that sufficient fees have been provided that
// meet a minimum threshold defined by the proposer (for mempool purposes during CheckTx).
type EthMempoolFeeDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthMempoolFeeDecorator creates a new EthMempoolFeeDecorator
func NewEthMempoolFeeDecorator(ek EVMKeeper) EthMempoolFeeDecorator {
	return EthMempoolFeeDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle verifies that enough fees have been provided by the
// Ethereum transaction that meet the minimum threshold set by the block
// proposer.
//
// NOTE: This should only be run during a CheckTx mode.
func (emfd EthMempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgEthTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type, not implements sdk.FeeTx: %T", tx)
	}

	evmDenom := emfd.evmKeeper.GetParams(ctx).EvmDenom
	txFee := msgEthTx.GetFee().AmountOf(evmDenom).Int64()
	if txFee < 0 {
		return ctx, sdkerrors.Wrap(
			sdkerrors.ErrInsufficientFee,
			"negative fee not allowed",
		)
	}

	// txFee = GP * GL
	fee := sdk.NewInt64DecCoin(evmDenom, txFee)

	minGasPrices := ctx.MinGasPrices()

	// check that fee provided is greater than the minimum
	// NOTE: we only check if injs are present in min gas prices. It is up to the
	// sender if they want to send additional fees in other denominations.
	var hasEnoughFees bool
	if fee.Amount.GTE(minGasPrices.AmountOf(evmDenom)) {
		hasEnoughFees = true
	}

	// reject transaction if minimum gas price is positive and the transaction does not
	// meet the minimum fee
	if !ctx.MinGasPrices().IsZero() && !hasEnoughFees {
		return ctx, sdkerrors.Wrap(
			sdkerrors.ErrInsufficientFee,
			fmt.Sprintf("insufficient fee, got: %q required: %q", fee, ctx.MinGasPrices()),
		)
	}

	return next(ctx, tx, simulate)
}

// EthValidateBasicDecorator will call tx.ValidateBasic and return any non-nil error.
// If ValidateBasic passes, decorator calls next AnteHandler in chain. Note,
// EthValidateBasicDecorator decorator will not get executed on ReCheckTx since it
// is not dependent on application state.
type EthValidateBasicDecorator struct{}

func NewEthValidateBasicDecorator() EthValidateBasicDecorator {
	return EthValidateBasicDecorator{}
}

func (vbd EthValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to validate basic on recheck tx, call next antehandler
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	if err := msgEthTx.ValidateBasic(); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// EthSigVerificationDecorator validates an ethereum signature
type EthSigVerificationDecorator struct {
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator() EthSigVerificationDecorator {
	return EthSigVerificationDecorator{}
}

// AnteHandle validates the signature and returns sender address
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if simulate {
		// when simulating, no signatures required and the from address is explicitly set
		return next(ctx, tx, simulate)
	}

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := sidechain.ParseChainID(ctx.ChainID())
	if err != nil {
		fmt.Println("chain id parsing failed")

		return ctx, err
	}

	// validate sender/signature
	_, eip155Err := msgEthTx.VerifySig(chainIDEpoch)
	if eip155Err != nil {
		_, homesteadErr := msgEthTx.VerifySigHomestead()
		if homesteadErr != nil {
			errMsg := fmt.Sprintf("signature verification failed for both EIP155 and Homestead signers: (%s, %s)",
				eip155Err.Error(), homesteadErr.Error())
			err := sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)
			return ctx, err
		}
	}

	// NOTE: when signature verification succeeds, a non-empty signer address can be
	// retrieved from the transaction on the next AnteDecorators.

	return next(ctx, tx, simulate)
}

type noMessages struct{}

func getTxMsg(tx sdk.Tx) interface{} {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return &noMessages{}
	}

	return msgs[0]
}

// EthAccountVerificationDecorator validates an account balance checks
type EthAccountVerificationDecorator struct {
	ak         AccountKeeper
	bankKeeper BankKeeper
	evmKeeper  EVMKeeper
}

// NewEthAccountVerificationDecorator creates a new EthAccountVerificationDecorator
func NewEthAccountVerificationDecorator(ak AccountKeeper, bankKeeper BankKeeper, ek EVMKeeper) EthAccountVerificationDecorator {
	return EthAccountVerificationDecorator{
		ak:         ak,
		bankKeeper: bankKeeper,
		evmKeeper:  ek,
	}
}

// AnteHandle validates the signature and returns sender address
func (avd EthAccountVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// sender address should be in the tx cache from the previous AnteHandle call
	address := msgEthTx.GetFrom()
	if address.Empty() {
		log.Panicln("sender address cannot be empty")
	}

	acc := avd.ak.GetAccount(ctx, address)
	if acc == nil {
		acc = avd.ak.NewAccountWithAddress(ctx, address)
		avd.ak.SetAccount(ctx, acc)
	}

	// on InitChain make sure account number == 0
	if ctx.BlockHeight() == 0 && acc.GetAccountNumber() != 0 {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"invalid account number for height zero (got %d)", acc.GetAccountNumber(),
		)
	}

	evmDenom := avd.evmKeeper.GetParams(ctx).EvmDenom

	// validate sender has enough funds to pay for gas cost
	balance := avd.bankKeeper.GetBalance(ctx, address, evmDenom)
	if balance.Amount.BigInt().Cmp(msgEthTx.Cost()) < 0 {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"sender balance < tx gas cost (%s%s < %s%s)", balance.String(), evmDenom, msgEthTx.Cost().String(), evmDenom,
		)
	}

	return next(ctx, tx, simulate)
}

// EthNonceVerificationDecorator checks that the account nonce from the transaction matches
// the sender account sequence.
type EthNonceVerificationDecorator struct {
	ak AccountKeeper
}

// NewEthNonceVerificationDecorator creates a new EthNonceVerificationDecorator
func NewEthNonceVerificationDecorator(ak AccountKeeper) EthNonceVerificationDecorator {
	return EthNonceVerificationDecorator{
		ak: ak,
	}
}

// AnteHandle validates that the transaction nonce is valid (equivalent to the sender account’s
// current nonce).
func (nvd EthNonceVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// sender address should be in the tx cache from the previous AnteHandle call
	address := msgEthTx.GetFrom()
	if address.Empty() {
		log.Panicln("sender address cannot be empty")
	}

	acc := nvd.ak.GetAccount(ctx, address)
	if acc == nil {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrUnknownAddress,
			"account %s (%s) is nil", common.BytesToAddress(address.Bytes()), address,
		)
	}

	seq := acc.GetSequence()
	// if multiple transactions are submitted in succession with increasing nonces,
	// all will be rejected except the first, since the first needs to be included in a block
	// before the sequence increments
	if msgEthTx.Data.AccountNonce != seq {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"invalid nonce; got %d, expected %d", msgEthTx.Data.AccountNonce, seq,
		)
	}

	return next(ctx, tx, simulate)
}

// EthGasConsumeDecorator validates enough intrinsic gas for the transaction and
// gas consumption.
type EthGasConsumeDecorator struct {
	ak         AccountKeeper
	bankKeeper BankKeeper
	evmKeeper  EVMKeeper
}

// NewEthGasConsumeDecorator creates a new EthGasConsumeDecorator
func NewEthGasConsumeDecorator(ak AccountKeeper, bankKeeper BankKeeper, ek EVMKeeper) EthGasConsumeDecorator {
	return EthGasConsumeDecorator{
		ak:         ak,
		bankKeeper: bankKeeper,
		evmKeeper:  ek,
	}
}

// AnteHandle validates that the Ethereum tx message has enough to cover intrinsic gas
// (during CheckTx only) and that the sender has enough balance to pay for the gas cost.
//
// Intrinsic gas for a transaction is the amount of gas
// that the transaction uses before the transaction is executed. The gas is a
// constant value of 21000 plus any cost inccured by additional bytes of data
// supplied with the transaction.
func (egcd EthGasConsumeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// sender address should be in the tx cache from the previous AnteHandle call
	address := msgEthTx.GetFrom()
	if address.Empty() {
		log.Panicln("sender address cannot be empty")
	}

	// fetch sender account from signature
	senderAcc, err := authante.GetSignerAcc(ctx, egcd.ak, address)
	if err != nil {
		return ctx, err
	}

	if senderAcc == nil {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrUnknownAddress,
			"sender account %s (%s) is nil", common.BytesToAddress(address.Bytes()), address,
		)
	}

	gasLimit := msgEthTx.GetGas()
	gas, err := ethcore.IntrinsicGas(msgEthTx.Data.Payload, msgEthTx.To() == nil, true, false)
	if err != nil {
		return ctx, sdkerrors.Wrap(err, "failed to compute intrinsic gas cost")
	}

	// intrinsic gas verification during CheckTx
	if ctx.IsCheckTx() && gasLimit < gas {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "intrinsic gas too low: %d < %d", gasLimit, gas)
	}

	// Charge sender for gas up to limit
	if gasLimit != 0 {
		// Cost calculates the fees paid to validators based on gas limit and price
		cost := new(big.Int).Mul(new(big.Int).SetBytes(msgEthTx.Data.Price), new(big.Int).SetUint64(gasLimit))

		evmDenom := egcd.evmKeeper.GetParams(ctx).EvmDenom

		feeAmt := sdk.NewCoins(
			sdk.NewCoin(evmDenom, sdk.NewIntFromBigInt(cost)),
		)

		err = authante.DeductFees(egcd.bankKeeper, ctx, senderAcc, feeAmt)
		if err != nil {
			return ctx, err
		}
	}

	// Set gas meter after ante handler to ignore gaskv costs
	newCtx = authante.SetGasMeter(simulate, ctx, gasLimit)
	return next(newCtx, tx, simulate)
}

// EthIncrementSenderSequenceDecorator increments the sequence of the signers. The
// main difference with the SDK's IncrementSequenceDecorator is that the MsgEthereumTx
// doesn't implement the SigVerifiableTx interface.
//
// CONTRACT: must be called after msg.VerifySig in order to cache the sender address.
type EthIncrementSenderSequenceDecorator struct {
	ak AccountKeeper
}

// NewEthIncrementSenderSequenceDecorator creates a new EthIncrementSenderSequenceDecorator.
func NewEthIncrementSenderSequenceDecorator(ak AccountKeeper) EthIncrementSenderSequenceDecorator {
	return EthIncrementSenderSequenceDecorator{
		ak: ak,
	}
}

// AnteHandle handles incrementing the sequence of the sender.
func (issd EthIncrementSenderSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		ctx = ctx.WithGasMeter(gasMeter)
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// increment sequence of all signers
	for _, addr := range msgEthTx.GetSigners() {
		acc := issd.ak.GetAccount(ctx, addr)
		if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
			log.WithError(err).Panicln("failed to set acc sequence")
		}
		issd.ak.SetAccount(ctx, acc)
	}

	// set the original gas meter
	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
}

// EthAccountSetupDecorator sets an account to state if it's not stored already. This only applies for MsgEthermint.
type EthAccountSetupDecorator struct {
	ak AccountKeeper
}

// NewEthAccountSetupDecorator creates a new EthAccountSetupDecorator instance
func NewEthAccountSetupDecorator(ak AccountKeeper) EthAccountSetupDecorator {
	return EthAccountSetupDecorator{
		ak: ak,
	}
}

// AnteHandle sets an account for MsgEthereumTx (evm) if the sender is registered.
func (asd EthAccountSetupDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		ctx = ctx.WithGasMeter(gasMeter)
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	setupAccount(asd.ak, ctx, msgEthTx.GetFrom())

	// set the original gas meter
	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
}

func setupAccount(ak AccountKeeper, ctx sdk.Context, addr sdk.AccAddress) {
	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		return
	}

	acc = ak.NewAccountWithAddress(ctx, addr)
	ak.SetAccount(ctx, acc)
}

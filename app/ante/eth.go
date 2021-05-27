package ante

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	vm.StateDB

	ChainID() *big.Int
	GetParams(ctx sdk.Context) evmtypes.Params
	GetChainConfig(ctx sdk.Context) (evmtypes.ChainConfig, bool)
	WithContext(ctx sdk.Context)
	ResetRefundTransient(ctx sdk.Context)
	PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList)
	NewEVM(msg core.Message, config *params.ChainConfig) *vm.EVM
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
	if !ctx.IsCheckTx() || simulate {
		return next(ctx, tx, simulate)
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type, not implements sdk.FeeTx: %T", tx)
	}

	evmDenom := emfd.evmKeeper.GetParams(ctx).EvmDenom

	// fee cost = gas price * gas limit
	// NOTE: this panics if the cost's BitLen is > 255
	fee := sdk.NewDecCoin(evmDenom, sdk.NewIntFromBigInt(msgEthTx.Cost()))

	minGasPrices := ctx.MinGasPrices()

	// check that fee provided is greater than the minimum
	//
	// NOTE: we only check if the evm tokens are present in min gas prices.
	// It is up to the sender if they want to send additional fees in other denominations.
	var hasEnoughFees bool
	if fee.Amount.GTE(minGasPrices.AmountOf(evmDenom)) {
		hasEnoughFees = true
	}

	// reject transaction if minimum gas price is positive and the transaction does not
	// meet the minimum fee

	// NOTE: here we are supporting 0 fee txs
	if !ctx.MinGasPrices().IsZero() && !hasEnoughFees {
		return ctx, sdkerrors.Wrap(
			sdkerrors.ErrInsufficientFee,
			fmt.Sprintf("insufficient fee, got: %q required: %q", fee, ctx.MinGasPrices()),
		)
	}

	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
}

// EthSigVerificationDecorator validates an ethereum signature
type EthSigVerificationDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator(ek EVMKeeper) EthSigVerificationDecorator {
	return EthSigVerificationDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle validates checks that the registered chain id is the same as the one on the message, and
// that the signer address matches the one defined on the message.
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	chainID := esvd.evmKeeper.ChainID()

	config, found := esvd.evmKeeper.GetChainConfig(ctx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(chainID)

	blockNum := big.NewInt(ctx.BlockHeight())
	signer := ethtypes.MakeSigner(ethCfg, blockNum)

	sender, err := signer.Sender(msgEthTx.AsTransaction())
	if err != nil {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error())
	}

	// set the sender
	msgEthTx.From = sender.String()

	ctx = ctx.WithGasMeter(gasMeter)

	// NOTE: when signature verification succeeds, a non-empty signer address can be
	return next(ctx, msgEthTx, simulate)
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

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// sender address should be in the tx cache from the previous AnteHandle call
	from := msgEthTx.GetFrom()
	if from.Empty() {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "from address cannot be empty")
	}

	acc := avd.ak.GetAccount(ctx, from)
	if acc == nil {
		_ = avd.ak.NewAccountWithAddress(ctx, from)
	}

	evmDenom := avd.evmKeeper.GetParams(ctx).EvmDenom

	// validate sender has enough funds to pay for gas cost
	balance := avd.bankKeeper.GetBalance(ctx, from, evmDenom)
	if balance.Amount.BigInt().Cmp(msgEthTx.Cost()) < 0 {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"sender balance < tx gas cost (%s < %s%s)", balance.String(), msgEthTx.Cost().String(), evmDenom,
		)
	}

	ctx = ctx.WithGasMeter(gasMeter)
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
	// no need to check the nonce on ReCheckTx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// sender address should be in the tx cache from the previous AnteHandle call
	seq, err := nvd.ak.GetSequence(ctx, msgEthTx.GetFrom())
	if err != nil {
		return ctx, err
	}

	// if multiple transactions are submitted in succession with increasing nonces,
	// all will be rejected except the first, since the first needs to be included in a block
	// before the sequence increments
	if msgEthTx.Data.Nonce != seq {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrInvalidSequence,
			"invalid nonce; got %d, expected %d", msgEthTx.Data.Nonce, seq,
		)
	}

	ctx = ctx.WithGasMeter(gasMeter)
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
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// reset the refund gas value for the current transaction
	egcd.evmKeeper.ResetRefundTransient(ctx)

	config, found := egcd.evmKeeper.GetChainConfig(ctx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(egcd.evmKeeper.ChainID())

	blockHeight := big.NewInt(ctx.BlockHeight())
	homestead := ethCfg.IsHomestead(blockHeight)
	istanbul := ethCfg.IsIstanbul(blockHeight)
	isContractCreation := msgEthTx.To() == nil

	// fetch sender account from signature
	signerAcc, err := authante.GetSignerAcc(ctx, egcd.ak, msgEthTx.GetFrom())
	if err != nil {
		return ctx, err
	}

	gasLimit := msgEthTx.GetGas()

	var accessList ethtypes.AccessList
	if msgEthTx.Data.Accesses != nil {
		accessList = *msgEthTx.Data.Accesses.ToEthAccessList()
	}

	gas, err := core.IntrinsicGas(msgEthTx.Data.Input, accessList, isContractCreation, homestead, istanbul)
	if err != nil {
		return ctx, sdkerrors.Wrap(err, "failed to compute intrinsic gas cost")
	}

	// intrinsic gas verification during CheckTx
	if ctx.IsCheckTx() && gasLimit < gas {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "intrinsic gas too low: %d < %d", gasLimit, gas)
	}

	// Cost calculates the fees paid to validators based on gas limit and price
	cost := msgEthTx.Cost()

	evmDenom := egcd.evmKeeper.GetParams(ctx).EvmDenom
	feeAmt := sdk.Coins{sdk.NewCoin(evmDenom, sdk.NewIntFromBigInt(cost))}

	err = authante.DeductFees(egcd.bankKeeper, ctx, signerAcc, feeAmt)
	if err != nil {
		return ctx, err
	}

	ctx = ctx.WithGasMeter(gasMeter) // set the original gas meter limit

	// consume gas for the current transaction. After runTx is executed on Baseapp, the application will consume gas
	// from the block gas pool.
	ctx.GasMeter().ConsumeGas(gas, "intrinsic gas")

	// generate a copy of the gas pool (i.e block gas meter) to see if we've run out of gas for this block
	// if current gas consumed is greater than the limit, this funcion panics and the error is recovered on the Baseapp
	gasPool := sdk.NewGasMeter(ctx.BlockGasMeter().Limit())
	gasPool.ConsumeGas(ctx.GasMeter().GasConsumedToLimit(), "gas pool check")

	// we know that we have enough gas on the pool to cover the intrinsic gas
	// set up the updated context to the evm Keeper
	egcd.evmKeeper.WithContext(ctx)
	return next(ctx, tx, simulate)
}

type CanTransferDecorator struct {
	evmKeeper EVMKeeper
}

// NewCanTransferDecorator creates a new CanTransferDecorator.
func NewCanTransferDecorator(evmKeeper EVMKeeper) CanTransferDecorator {
	return CanTransferDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle creates an EVM from the message and calls the BlockContext CanTransfer function to
// see if the address can execute the transaction.
func (ctd CanTransferDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	msg, err := msgEthTx.AsMessage()
	if err != nil {
		return ctx, err
	}

	config, found := ctd.evmKeeper.GetChainConfig(ctx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(ctd.evmKeeper.ChainID())

	evm := ctd.evmKeeper.NewEVM(msg, ethCfg)

	// check that caller has enough balance to cover asset transfer for **topmost** call
	// NOTE: here the gas consumed is from the original context
	if msg.Value().Sign() > 0 && !evm.Context.CanTransfer(ctd.evmKeeper, msg.From(), msg.Value()) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "address %s", msg.From().Hex())
	}

	// set the original gas meter
	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
}

type AccessListDecorator struct {
	evmKeeper EVMKeeper
}

// NewAccessListDecorator creates a new AccessListDecorator.
func NewAccessListDecorator(evmKeeper EVMKeeper) AccessListDecorator {
	return AccessListDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle handles the preparatory steps for executing an EVM state transition with
// regards to both EIP-2929 and EIP-2930:
//
// 	- Add sender to access list (2929)
// 	- Add destination to access list (2929)
// 	- Add precompiles to access list (2929)
// 	- Add the contents of the optional tx access list (2930)
//
// This method should only be called if Yolov3/Berlin/2929+2930 is applicable at the current number.
func (ald AccessListDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	gasMeter := ctx.GasMeter()
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	config, found := ald.evmKeeper.GetChainConfig(ctx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	msgEthTx, ok := getTxMsg(tx).(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	msg, err := msgEthTx.AsMessage()
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "tx cannot be expressed as core.Message: %s", err.Error())
	}

	ethCfg := config.EthereumConfig(ald.evmKeeper.ChainID())

	// setup the keeper context before setting the access list
	ald.evmKeeper.WithContext(ctx)

	if rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight())); rules.IsBerlin {
		ald.evmKeeper.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	}

	// set the original gas meter
	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
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
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", getTxMsg(tx))
	}

	// increment sequence of all signers
	for _, addr := range msgEthTx.GetSigners() {
		acc := issd.ak.GetAccount(ctx, addr)

		if acc == nil {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownAddress,
				"account %s (%s) is nil", common.BytesToAddress(addr.Bytes()), addr,
			)
		}

		if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
			return ctx, err
		}

		issd.ak.SetAccount(ctx, acc)
	}

	// set the original gas meter
	ctx = ctx.WithGasMeter(gasMeter)
	return next(ctx, tx, simulate)
}

package ante

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	tx "github.com/cosmos/cosmos-sdk/types/tx"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"

	ethermint "github.com/tharsis/ethermint/types"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// EVMKeeper defines the expected keeper interface used on the Eth AnteHandler
type EVMKeeper interface {
	vm.StateDB

	ChainID() *big.Int
	GetParams(ctx sdk.Context) evmtypes.Params
	WithContext(ctx sdk.Context)
	ResetRefundTransient(ctx sdk.Context)
	NewEVM(msg core.Message, cfg *evmtypes.EVMConfig, tracer vm.Tracer) *vm.EVM
	GetCodeHash(addr common.Address) common.Hash
	DeductTxCostsFromUserBalance(
		ctx sdk.Context, msgEthTx evmtypes.MsgEthereumTx, txData evmtypes.TxData, denom string, homestead, istanbul, london bool,
	) (sdk.Coins, error)
}

type protoTxProvider interface {
	GetProtoTx() *tx.Tx
}

// EthSigVerificationDecorator validates an ethereum signatures
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
// It's not skipped for RecheckTx, because it set `From` address which is critical from other ante handler to work.
// Failure in RecheckTx will prevent tx to be included into block, especially when CheckTx succeed, in which case user
// won't see the error message.
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if tx == nil || len(tx.GetMsgs()) != 1 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only 1 ethereum msg supported per tx")
	}

	chainID := esvd.evmKeeper.ChainID()

	params := esvd.evmKeeper.GetParams(ctx)

	ethCfg := params.ChainConfig.EthereumConfig(chainID)
	blockNum := big.NewInt(ctx.BlockHeight())
	signer := ethtypes.MakeSigner(ethCfg, blockNum)

	msg := tx.GetMsgs()[0]
	msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
	}

	sender, err := signer.Sender(msgEthTx.AsTransaction())
	if err != nil {
		return ctx, sdkerrors.Wrapf(
			sdkerrors.ErrorInvalidSigner,
			"couldn't retrieve sender address ('%s') from the ethereum transaction: %s",
			msgEthTx.From,
			err.Error(),
		)
	}

	// set up the sender to the transaction field if not already
	msgEthTx.From = sender.Hex()

	return next(ctx, msgEthTx, simulate)
}

// EthAccountVerificationDecorator validates an account balance checks
type EthAccountVerificationDecorator struct {
	ak         evmtypes.AccountKeeper
	bankKeeper evmtypes.BankKeeper
	evmKeeper  EVMKeeper
}

// NewEthAccountVerificationDecorator creates a new EthAccountVerificationDecorator
func NewEthAccountVerificationDecorator(ak evmtypes.AccountKeeper, bankKeeper evmtypes.BankKeeper, ek EVMKeeper) EthAccountVerificationDecorator {
	return EthAccountVerificationDecorator{
		ak:         ak,
		bankKeeper: bankKeeper,
		evmKeeper:  ek,
	}
}

// AnteHandle validates checks that the sender balance is greater than the total transaction cost.
// The account will be set to store if it doesn't exis, i.e cannot be found on store.
// This AnteHandler decorator will fail if:
// - any of the msgs is not a MsgEthereumTx
// - from address is empty
// - account balance is lower than the transaction cost
func (avd EthAccountVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	avd.evmKeeper.WithContext(ctx)
	evmDenom := avd.evmKeeper.GetParams(ctx).EvmDenom

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "failed to unpack tx data any for tx %d", i)
		}

		// sender address should be in the tx cache from the previous AnteHandle call
		from := msgEthTx.GetFrom()
		if from.Empty() {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "from address cannot be empty")
		}

		// check whether the sender address is EOA
		fromAddr := common.BytesToAddress(from)
		codeHash := avd.evmKeeper.GetCodeHash(fromAddr)
		if codeHash != common.BytesToHash(evmtypes.EmptyCodeHash) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidType,
				"the sender is not EOA: address <%v>, codeHash <%s>", fromAddr, codeHash)
		}

		acc := avd.ak.GetAccount(ctx, from)
		if acc == nil {
			acc = avd.ak.NewAccountWithAddress(ctx, from)
			avd.ak.SetAccount(ctx, acc)
		}

		if err := evmkeeper.CheckSenderBalance(ctx, avd.bankKeeper, from, txData, evmDenom); err != nil {
			return ctx, sdkerrors.Wrap(err, "failed to check sender balance")
		}

	}
	// recover  the original gas meter
	avd.evmKeeper.WithContext(ctx)
	return next(ctx, tx, simulate)
}

// EthNonceVerificationDecorator checks that the account nonce from the transaction matches
// the sender account sequence.
type EthNonceVerificationDecorator struct {
	ak evmtypes.AccountKeeper
}

// NewEthNonceVerificationDecorator creates a new EthNonceVerificationDecorator
func NewEthNonceVerificationDecorator(ak evmtypes.AccountKeeper) EthNonceVerificationDecorator {
	return EthNonceVerificationDecorator{
		ak: ak,
	}
}

// AnteHandle validates that the transaction nonces are valid and equivalent to the sender account’s
// current nonce.
func (nvd EthNonceVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to check the nonce on ReCheckTx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}

		// sender address should be in the tx cache from the previous AnteHandle call
		seq, err := nvd.ak.GetSequence(ctx, msgEthTx.GetFrom())
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "sequence not found for address %s", msgEthTx.From)
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, sdkerrors.Wrap(err, "failed to unpack tx data")
		}

		// if multiple transactions are submitted in succession with increasing nonces,
		// all will be rejected except the first, since the first needs to be included in a block
		// before the sequence increments
		if txData.GetNonce() != seq {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidSequence,
				"invalid nonce; got %d, expected %d", txData.GetNonce(), seq,
			)
		}
	}

	return next(ctx, tx, simulate)
}

// EthGasConsumeDecorator validates enough intrinsic gas for the transaction and
// gas consumption.
type EthGasConsumeDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthGasConsumeDecorator creates a new EthGasConsumeDecorator
func NewEthGasConsumeDecorator(
	evmKeeper EVMKeeper,
) EthGasConsumeDecorator {
	return EthGasConsumeDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle validates that the Ethereum tx message has enough to cover intrinsic gas
// (during CheckTx only) and that the sender has enough balance to pay for the gas cost.
//
// Intrinsic gas for a transaction is the amount of gas that the transaction uses before the
// transaction is executed. The gas is a constant value plus any cost inccured by additional bytes
// of data supplied with the transaction.
//
// This AnteHandler decorator will fail if:
// - the transaction contains more than one message
// - the message is not a MsgEthereumTx
// - sender account cannot be found
// - transaction's gas limit is lower than the intrinsic gas
// - user doesn't have enough balance to deduct the transaction fees (gas_limit * gas_price)
// - transaction or block gas meter runs out of gas
func (egcd EthGasConsumeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// reset the refund gas value in the keeper for the current transaction
	egcd.evmKeeper.ResetRefundTransient(ctx)

	params := egcd.evmKeeper.GetParams(ctx)

	ethCfg := params.ChainConfig.EthereumConfig(egcd.evmKeeper.ChainID())

	blockHeight := big.NewInt(ctx.BlockHeight())
	homestead := ethCfg.IsHomestead(blockHeight)
	istanbul := ethCfg.IsIstanbul(blockHeight)
	london := ethCfg.IsLondon(blockHeight)
	evmDenom := params.EvmDenom

	var events sdk.Events

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, sdkerrors.Wrap(err, "failed to unpack tx data")
		}

		fees, err := egcd.evmKeeper.DeductTxCostsFromUserBalance(
			ctx,
			*msgEthTx,
			txData,
			evmDenom,
			homestead,
			istanbul,
			london,
		)
		if err != nil {
			return ctx, sdkerrors.Wrapf(err, "failed to deduct transaction costs from user balance")
		}

		events = append(events, sdk.NewEvent(sdk.EventTypeTx, sdk.NewAttribute(sdk.AttributeKeyFee, fees.String())))
	}

	// TODO: change to typed events
	ctx.EventManager().EmitEvents(events)

	// TODO: deprecate after https://github.com/cosmos/cosmos-sdk/issues/9514  is fixed on SDK
	blockGasLimit := ethermint.BlockGasLimit(ctx)

	// NOTE: safety check
	if blockGasLimit > 0 {
		// generate a copy of the gas pool (i.e block gas meter) to see if we've run out of gas for this block
		// if current gas consumed is greater than the limit, this funcion panics and the error is recovered on the Baseapp
		gasPool := sdk.NewGasMeter(blockGasLimit)
		gasPool.ConsumeGas(ctx.GasMeter().GasConsumedToLimit(), "gas pool check")
	}

	// we know that we have enough gas on the pool to cover the intrinsic gas
	// set up the updated context to the evm Keeper
	egcd.evmKeeper.WithContext(ctx)
	return next(ctx, tx, simulate)
}

// CanTransferDecorator checks if the sender is allowed to transfer funds according to the EVM block
// context rules.
type CanTransferDecorator struct {
	evmKeeper       EVMKeeper
	feemarketKeeper evmtypes.FeeMarketKeeper
}

// NewCanTransferDecorator creates a new CanTransferDecorator instance.
func NewCanTransferDecorator(evmKeeper EVMKeeper, fmk evmtypes.FeeMarketKeeper) CanTransferDecorator {
	return CanTransferDecorator{
		evmKeeper:       evmKeeper,
		feemarketKeeper: fmk,
	}
}

// AnteHandle creates an EVM from the message and calls the BlockContext CanTransfer function to
// see if the address can execute the transaction.
func (ctd CanTransferDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	ctd.evmKeeper.WithContext(ctx)

	params := ctd.evmKeeper.GetParams(ctx)
	feeMktParams := ctd.feemarketKeeper.GetParams(ctx)

	ethCfg := params.ChainConfig.EthereumConfig(ctd.evmKeeper.ChainID())
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}

		var baseFee *big.Int
		if evmtypes.IsLondon(ethCfg, ctx.BlockHeight()) && !feeMktParams.NoBaseFee {
			baseFee = ctd.feemarketKeeper.GetBaseFee(ctx)
		}

		coreMsg, err := msgEthTx.AsMessage(signer, baseFee)
		if err != nil {
			return ctx, sdkerrors.Wrapf(
				err,
				"failed to create an ethereum core.Message from signer %T", signer,
			)
		}

		// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
		cfg := &evmtypes.EVMConfig{
			ChainConfig: ethCfg,
			Params:      params,
			CoinBase:    common.Address{},
			BaseFee:     baseFee,
		}
		evm := ctd.evmKeeper.NewEVM(coreMsg, cfg, evmtypes.NewNoOpTracer())

		// check that caller has enough balance to cover asset transfer for **topmost** call
		// NOTE: here the gas consumed is from the context with the infinite gas meter
		if coreMsg.Value().Sign() > 0 && !evm.Context.CanTransfer(ctd.evmKeeper, coreMsg.From(), coreMsg.Value()) {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrInsufficientFunds,
				"failed to transfer %s from address %s using the EVM block context transfer function",
				coreMsg.Value(),
				coreMsg.From(),
			)
		}

		if evmtypes.IsLondon(ethCfg, ctx.BlockHeight()) && !feeMktParams.NoBaseFee && baseFee == nil {
			return ctx, sdkerrors.Wrap(evmtypes.ErrInvalidBaseFee, "base fee is supported but evm block context value is nil")
		}

		if evmtypes.IsLondon(ethCfg, ctx.BlockHeight()) && !feeMktParams.NoBaseFee && baseFee != nil && coreMsg.GasFeeCap().Cmp(baseFee) < 0 {
			return ctx, sdkerrors.Wrapf(evmtypes.ErrInvalidBaseFee, "max fee per gas less than block base fee (%s < %s)", coreMsg.GasFeeCap(), baseFee)
		}
	}

	ctd.evmKeeper.WithContext(ctx)

	// set the original gas meter
	return next(ctx, tx, simulate)
}

// EthIncrementSenderSequenceDecorator increments the sequence of the signers.
type EthIncrementSenderSequenceDecorator struct {
	ak evmtypes.AccountKeeper
}

// NewEthIncrementSenderSequenceDecorator creates a new EthIncrementSenderSequenceDecorator.
func NewEthIncrementSenderSequenceDecorator(ak evmtypes.AccountKeeper) EthIncrementSenderSequenceDecorator {
	return EthIncrementSenderSequenceDecorator{
		ak: ak,
	}
}

// AnteHandle handles incrementing the sequence of the signer (i.e sender). If the transaction is a
// contract creation, the nonce will be incremented during the transaction execution and not within
// this AnteHandler decorator.
func (issd EthIncrementSenderSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		// increment sequence of all signers
		for _, addr := range msg.GetSigners() {
			acc := issd.ak.GetAccount(ctx, addr)

			if acc == nil {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrUnknownAddress,
					"account %s (%s) is nil", common.BytesToAddress(addr.Bytes()), addr,
				)
			}

			if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
				return ctx, sdkerrors.Wrapf(err, "failed to set sequence to %d", acc.GetSequence()+1)
			}

			issd.ak.SetAccount(ctx, acc)
		}
	}

	return next(ctx, tx, simulate)
}

// EthValidateBasicDecorator is adapted from ValidateBasicDecorator from cosmos-sdk, it ignores ErrNoSignatures
type EthValidateBasicDecorator struct {
	evmKeeper EVMKeeper
}

// NewEthValidateBasicDecorator creates a new EthValidateBasicDecorator
func NewEthValidateBasicDecorator(ek EVMKeeper) EthValidateBasicDecorator {
	return EthValidateBasicDecorator{
		evmKeeper: ek,
	}
}

// AnteHandle handles basic validation of tx
func (vbd EthValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to validate basic on recheck tx, call next antehandler
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	err := tx.ValidateBasic()
	// ErrNoSignatures is fine with eth tx
	if err != nil && !errors.Is(err, sdkerrors.ErrNoSignatures) {
		return ctx, sdkerrors.Wrap(err, "tx basic validation failed")
	}

	// For eth type cosmos tx, some fields should be veified as zero values,
	// since we will only verify the signature against the hash of the MsgEthereumTx.Data
	if wrapperTx, ok := tx.(protoTxProvider); ok {
		protoTx := wrapperTx.GetProtoTx()
		body := protoTx.Body
		if body.Memo != "" || body.TimeoutHeight != uint64(0) || len(body.NonCriticalExtensionOptions) > 0 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest,
				"for eth tx body Memo TimeoutHeight NonCriticalExtensionOptions should be empty")
		}

		if len(body.ExtensionOptions) != 1 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "for eth tx length of ExtensionOptions should be 1")
		}

		if len(protoTx.GetMsgs()) != 1 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only 1 ethereum msg supported per tx")
		}
		msg := protoTx.GetMsgs()[0]
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}
		ethGasLimit := msgEthTx.GetGas()

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, sdkerrors.Wrap(err, "failed to unpack MsgEthereumTx Data")
		}
		params := vbd.evmKeeper.GetParams(ctx)
		ethFeeAmount := sdk.Coins{sdk.NewCoin(params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}

		authInfo := protoTx.AuthInfo
		if len(authInfo.SignerInfos) > 0 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "for eth tx AuthInfo SignerInfos should be empty")
		}

		if authInfo.Fee.Payer != "" || authInfo.Fee.Granter != "" {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "for eth tx AuthInfo Fee payer and granter should be empty")
		}

		if !authInfo.Fee.Amount.IsEqual(ethFeeAmount) {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid eth tx AuthInfo Fee Amount")
		}

		if authInfo.Fee.GasLimit != ethGasLimit {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid eth tx AuthInfo Fee GasLimit")
		}

		sigs := protoTx.Signatures
		if len(sigs) > 0 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "for eth tx Signatures should be empty")
		}
	}

	return next(ctx, tx, simulate)
}

// EthSetupContextDecorator is adapted from SetUpContextDecorator from cosmos-sdk, it ignores gas consumption
// by setting the gas meter to infinite
type EthSetupContextDecorator struct{}

func NewEthSetUpContextDecorator() EthSetupContextDecorator {
	return EthSetupContextDecorator{}
}

func (esc EthSetupContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// all transactions must implement GasTx
	_, ok := tx.(authante.GasTx)
	if !ok {
		return newCtx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be GasTx")
	}

	newCtx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	return next(newCtx, tx, simulate)
}

// EthMempoolFeeDecorator will check if the transaction's effective fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator
type EthMempoolFeeDecorator struct {
	feemarketKeeper evmtypes.FeeMarketKeeper
	evmKeeper       EVMKeeper
}

func NewEthMempoolFeeDecorator(ek EVMKeeper, fmk evmtypes.FeeMarketKeeper) EthMempoolFeeDecorator {
	return EthMempoolFeeDecorator{
		feemarketKeeper: fmk,
		evmKeeper:       ek,
	}
}

// AnteHandle ensures that the provided fees meet a minimum threshold for the validator,
// if this is a CheckTx. This is only for local mempool purposes, and thus
// is only ran on check tx.
func (mfd EthMempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.IsCheckTx() && !simulate {
		if len(tx.GetMsgs()) != 1 {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only 1 ethereum msg supported per tx")
		}
		msg, ok := tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, (*evmtypes.MsgEthereumTx)(nil))
		}

		var feeAmt *big.Int

		feeMktParams := mfd.feemarketKeeper.GetParams(ctx)
		params := mfd.evmKeeper.GetParams(ctx)
		chainID := mfd.evmKeeper.ChainID()
		ethCfg := params.ChainConfig.EthereumConfig(chainID)
		evmDenom := params.EvmDenom
		if evmtypes.IsLondon(ethCfg, ctx.BlockHeight()) && !feeMktParams.NoBaseFee {
			baseFee := mfd.feemarketKeeper.GetBaseFee(ctx)
			feeAmt = msg.GetEffectiveFee(baseFee)
		} else {
			feeAmt = msg.GetFee()
		}

		glDec := sdk.NewDec(int64(msg.GetGas()))
		requiredFee := ctx.MinGasPrices().AmountOf(evmDenom).Mul(glDec)
		if sdk.NewDecFromBigInt(feeAmt).LT(requiredFee) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeAmt, requiredFee)
		}
	}

	return next(ctx, tx, simulate)
}

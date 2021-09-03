package ante

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/palantir/stacktrace"

	ethermint "github.com/tharsis/ethermint/types"
	evmkeeper "github.com/tharsis/ethermint/x/evm/keeper"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"

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
	WithContext(ctx sdk.Context)
	ResetRefundTransient(ctx sdk.Context)
	NewEVM(msg core.Message, config *params.ChainConfig, params evmtypes.Params, coinbase common.Address, tracer vm.Tracer) *vm.EVM
	GetCodeHash(addr common.Address) common.Hash
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
		return ctx, stacktrace.Propagate(
			sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "only 1 ethereum msg supported per tx"),
			"",
		)
	}

	chainID := esvd.evmKeeper.ChainID()

	params := esvd.evmKeeper.GetParams(ctx)

	ethCfg := params.ChainConfig.EthereumConfig(chainID)
	blockNum := big.NewInt(ctx.BlockHeight())
	signer := ethtypes.MakeSigner(ethCfg, blockNum)

	msg := tx.GetMsgs()[0]
	msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
	if !ok {
		return ctx, stacktrace.Propagate(
			sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
			"failed to cast transaction",
		)
	}

	sender, err := signer.Sender(msgEthTx.AsTransaction())
	if err != nil {
		return ctx, stacktrace.Propagate(
			sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error()),
			"couldn't retrieve sender address ('%s') from the ethereum transaction",
			msgEthTx.From,
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
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to unpack tx data any for tx %d", i)
		}

		// sender address should be in the tx cache from the previous AnteHandle call
		from := msgEthTx.GetFrom()
		if from.Empty() {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "from address cannot be empty"),
				"sender address should have been in the tx field from the previous AnteHandle call",
			)
		}

		// check whether the sender address is EOA
		fromAddr := common.BytesToAddress(from)
		codeHash := avd.evmKeeper.GetCodeHash(fromAddr)
		if codeHash != common.BytesToHash(evmtypes.EmptyCodeHash) {
			return ctx, stacktrace.Propagate(sdkerrors.Wrapf(sdkerrors.ErrInvalidType,
				"the sender is not EOA: address <%v>, codeHash <%s>", fromAddr, codeHash), "")
		}

		acc := avd.ak.GetAccount(ctx, from)
		if acc == nil {
			acc = avd.ak.NewAccountWithAddress(ctx, from)
			avd.ak.SetAccount(ctx, acc)
		}

		if err := evmkeeper.CheckSenderBalance(ctx, avd.bankKeeper, from, txData, evmDenom); err != nil {
			return ctx, stacktrace.Propagate(err, "failed to check sender balance")
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

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		// sender address should be in the tx cache from the previous AnteHandle call
		seq, err := nvd.ak.GetSequence(ctx, msgEthTx.GetFrom())
		if err != nil {
			return ctx, stacktrace.Propagate(err, "sequence not found for address %s", msgEthTx.From)
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to unpack tx data")
		}

		// if multiple transactions are submitted in succession with increasing nonces,
		// all will be rejected except the first, since the first needs to be included in a block
		// before the sequence increments
		if txData.GetNonce() != seq {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(
					sdkerrors.ErrInvalidSequence,
					"invalid nonce; got %d, expected %d", txData.GetNonce(), seq,
				),
				"",
			)
		}
	}

	return next(ctx, tx, simulate)
}

// EthGasConsumeDecorator validates enough intrinsic gas for the transaction and
// gas consumption.
type EthGasConsumeDecorator struct {
	ak         evmtypes.AccountKeeper
	bankKeeper evmtypes.BankKeeper
	evmKeeper  EVMKeeper
}

// NewEthGasConsumeDecorator creates a new EthGasConsumeDecorator
func NewEthGasConsumeDecorator(ak evmtypes.AccountKeeper, bankKeeper evmtypes.BankKeeper, ek EVMKeeper) EthGasConsumeDecorator {
	return EthGasConsumeDecorator{
		ak:         ak,
		bankKeeper: bankKeeper,
		evmKeeper:  ek,
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
	evmDenom := params.EvmDenom

	var events sdk.Events

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to unpack tx data")
		}

		fees, err := evmkeeper.DeductTxCostsFromUserBalance(
			ctx,
			egcd.bankKeeper,
			egcd.ak,
			*msgEthTx,
			txData,
			evmDenom,
			homestead,
			istanbul,
		)

		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to deduct transaction costs from user balance")
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
	evmKeeper EVMKeeper
}

// NewCanTransferDecorator creates a new CanTransferDecorator instance.
func NewCanTransferDecorator(evmKeeper EVMKeeper) CanTransferDecorator {
	return CanTransferDecorator{
		evmKeeper: evmKeeper,
	}
}

// AnteHandle creates an EVM from the message and calls the BlockContext CanTransfer function to
// see if the address can execute the transaction.
func (ctd CanTransferDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	ctd.evmKeeper.WithContext(ctx)

	params := ctd.evmKeeper.GetParams(ctx)

	ethCfg := params.ChainConfig.EthereumConfig(ctd.evmKeeper.ChainID())
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		coreMsg, err := msgEthTx.AsMessage(signer)
		if err != nil {
			return ctx, stacktrace.Propagate(
				err,
				"failed to create an ethereum core.Message from signer %T", signer,
			)
		}

		// NOTE: pass in an empty coinbase address and nil tracer as we don't need them for the check below
		evm := ctd.evmKeeper.NewEVM(coreMsg, ethCfg, params, common.Address{}, nil)

		// check that caller has enough balance to cover asset transfer for **topmost** call
		// NOTE: here the gas consumed is from the context with the infinite gas meter
		if coreMsg.Value().Sign() > 0 && !evm.Context.CanTransfer(ctd.evmKeeper, coreMsg.From(), coreMsg.Value()) {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "address %s", coreMsg.From()),
				"failed to transfer %s using the EVM block context transfer function", coreMsg.Value(),
			)
		}
	}

	ctd.evmKeeper.WithContext(ctx)

	// set the original gas meter
	return next(ctx, tx, simulate)
}

// AccessListDecorator prepare an access list for the sender if Yolov3/Berlin/EIPs 2929 and 2930 are
// applicable at the current block number.
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
// The AnteHandler will only prepare the access list if Yolov3/Berlin/EIPs 2929 and 2930 are applicable at the current number.
func (ald AccessListDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {

	params := ald.evmKeeper.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(ald.evmKeeper.ChainID())

	rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight()))

	// we don't need to prepare the access list if the chain is not currently on the Berlin upgrade
	if !rules.IsBerlin {
		return next(ctx, tx, simulate)
	}

	// setup the keeper context before setting the access list
	ald.evmKeeper.WithContext(ctx)

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		sender := common.BytesToAddress(msgEthTx.GetFrom())

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to unpack tx data")
		}

		ald.evmKeeper.PrepareAccessList(sender, txData.GetTo(), vm.ActivePrecompiles(rules), txData.GetAccessList())
	}

	// set the original gas meter
	ald.evmKeeper.WithContext(ctx)
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

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		txData, err := evmtypes.UnpackTxData(msgEthTx.Data)
		if err != nil {
			return ctx, stacktrace.Propagate(err, "failed to unpack tx data")
		}

		// NOTE: on contract creation, the nonce is incremented within the EVM Create function during tx execution
		// and not previous to the state transition ¯\_(ツ)_/¯
		if txData.GetTo() == nil {
			// contract creation, don't increment sequence on AnteHandler but on tx execution
			// continue to the next item
			continue
		}

		// increment sequence of all signers
		for _, addr := range msg.GetSigners() {
			acc := issd.ak.GetAccount(ctx, addr)

			if acc == nil {
				return ctx, stacktrace.Propagate(
					sdkerrors.Wrapf(
						sdkerrors.ErrUnknownAddress,
						"account %s (%s) is nil", common.BytesToAddress(addr.Bytes()), addr,
					),
					"signer account not found",
				)
			}

			if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
				return ctx, stacktrace.Propagate(err, "failed to set sequence to %d", acc.GetSequence()+1)
			}

			issd.ak.SetAccount(ctx, acc)
		}
	}

	// set the original gas meter
	return next(ctx, tx, simulate)
}

// EthValidateBasicDecorator is adapted from ValidateBasicDecorator from cosmos-sdk, it ignores ErrNoSignatures
type EthValidateBasicDecorator struct{}

// NewEthValidateBasicDecorator creates a new EthValidateBasicDecorator
func NewEthValidateBasicDecorator() EthValidateBasicDecorator {
	return EthValidateBasicDecorator{}
}

// AnteHandle handles basic validation of tx
func (vbd EthValidateBasicDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to validate basic on recheck tx, call next antehandler
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	err := tx.ValidateBasic()
	// ErrNoSignatures is fine with eth tx
	if err != nil && err != sdkerrors.ErrNoSignatures {
		return ctx, stacktrace.Propagate(err, "tx basic validation failed")
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

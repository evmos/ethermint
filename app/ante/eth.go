package ante

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/palantir/stacktrace"

	ethermint "github.com/cosmos/ethermint/types"
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
	NewEVM(msg core.Message, config *params.ChainConfig) *vm.EVM
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
func (esvd EthSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if len(tx.GetMsgs()) != 1 {
		return ctx, stacktrace.Propagate(
			sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "only 1 ethereum msg supported per tx, got %d", len(tx.GetMsgs())),
			"",
		)
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	chainID := esvd.evmKeeper.ChainID()

	config, found := esvd.evmKeeper.GetChainConfig(infCtx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(chainID)
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

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	avd.evmKeeper.WithContext(infCtx)
	evmDenom := avd.evmKeeper.GetParams(infCtx).EvmDenom

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
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

		acc := avd.ak.GetAccount(infCtx, from)
		if acc == nil {
			acc = avd.ak.NewAccountWithAddress(infCtx, from)
			avd.ak.SetAccount(infCtx, acc)
		}

		// validate sender has enough funds to pay for tx cost
		balance := avd.bankKeeper.GetBalance(infCtx, from, evmDenom)
		if balance.Amount.BigInt().Cmp(msgEthTx.Cost()) < 0 {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(
					sdkerrors.ErrInsufficientFunds,
					"sender balance < tx cost (%s < %s%s)", balance, msgEthTx.Cost(), evmDenom,
				),
				"sender should have had enough funds to pay for tx cost = fee + amount (%s = %s + amount)", msgEthTx.Cost(), msgEthTx.Fee(),
			)
		}

	}
	// recover  the original gas meter
	avd.evmKeeper.WithContext(ctx)
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

// AnteHandle validates that the transaction nonces are valid and equivalent to the sender account’s
// current nonce.
func (nvd EthNonceVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to check the nonce on ReCheckTx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		// sender address should be in the tx cache from the previous AnteHandle call
		seq, err := nvd.ak.GetSequence(infCtx, msgEthTx.GetFrom())
		if err != nil {
			return ctx, stacktrace.Propagate(err, "sequence not found for address %s", msgEthTx.From)
		}

		// if multiple transactions are submitted in succession with increasing nonces,
		// all will be rejected except the first, since the first needs to be included in a block
		// before the sequence increments
		if msgEthTx.Data.Nonce != seq {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(
					sdkerrors.ErrInvalidSequence,
					"invalid nonce; got %d, expected %d", msgEthTx.Data.Nonce, seq,
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
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	// reset the refund gas value in the keeper for the current transaction
	egcd.evmKeeper.ResetRefundTransient(infCtx)

	config, found := egcd.evmKeeper.GetChainConfig(infCtx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(egcd.evmKeeper.ChainID())

	blockHeight := big.NewInt(ctx.BlockHeight())
	homestead := ethCfg.IsHomestead(blockHeight)
	istanbul := ethCfg.IsIstanbul(blockHeight)

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		isContractCreation := msgEthTx.To() == nil

		// fetch sender account from signature
		signerAcc, err := authante.GetSignerAcc(infCtx, egcd.ak, msgEthTx.GetFrom())
		if err != nil {
			return ctx, stacktrace.Propagate(err, "account not found for sender %s", msgEthTx.From)
		}

		gasLimit := msgEthTx.GetGas()

		var accessList ethtypes.AccessList
		if msgEthTx.Data.Accesses != nil {
			accessList = *msgEthTx.Data.Accesses.ToEthAccessList()
		}

		intrinsicGas, err := core.IntrinsicGas(msgEthTx.Data.Input, accessList, isContractCreation, homestead, istanbul)
		if err != nil {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrap(err, "failed to compute intrinsic gas cost"),
				"failed to retrieve intrinsic gas, contract creation = %t; homestead = %t, istanbul = %t", isContractCreation, homestead, istanbul)
		}

		// intrinsic gas verification during CheckTx
		if ctx.IsCheckTx() && gasLimit < intrinsicGas {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrOutOfGas, "gas limit too low: %d (gas limit) < %d (intrinsic gas)", gasLimit, intrinsicGas)
		}

		// calculate the fees paid to validators based on gas limit and price
		feeAmt := msgEthTx.Fee() // fee = gas limit * gas price

		evmDenom := egcd.evmKeeper.GetParams(infCtx).EvmDenom
		fees := sdk.Coins{sdk.NewCoin(evmDenom, sdk.NewIntFromBigInt(feeAmt))}

		// deduct the full gas cost from the user balance
		if err := authante.DeductFees(egcd.bankKeeper, infCtx, signerAcc, fees); err != nil {
			return ctx, stacktrace.Propagate(
				err,
				"failed to deduct full gas cost %s from the user %s balance", fees, msgEthTx.From,
			)
		}

		// consume intrinsic gas for the current transaction. After runTx is executed on Baseapp, the
		// application will consume gas from the block gas pool.
		ctx.GasMeter().ConsumeGas(intrinsicGas, "intrinsic gas")
	}

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
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	ctd.evmKeeper.WithContext(infCtx)

	config, found := ctd.evmKeeper.GetChainConfig(infCtx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(ctd.evmKeeper.ChainID())
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

		evm := ctd.evmKeeper.NewEVM(coreMsg, ethCfg)

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
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.

	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	config, found := ald.evmKeeper.GetChainConfig(infCtx)
	if !found {
		return ctx, evmtypes.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(ald.evmKeeper.ChainID())

	rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight()))

	// we don't need to prepare the access list if the chain is not currently on the Berlin upgrade
	if !rules.IsBerlin {
		return next(ctx, tx, simulate)
	}

	// setup the keeper context before setting the access list
	ald.evmKeeper.WithContext(infCtx)

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		sender := common.BytesToAddress(msgEthTx.GetFrom())

		ald.evmKeeper.PrepareAccessList(sender, msgEthTx.To(), vm.ActivePrecompiles(rules), *msgEthTx.Data.Accesses.ToEthAccessList())
	}

	// set the original gas meter
	ald.evmKeeper.WithContext(ctx)
	return next(ctx, tx, simulate)
}

// EthIncrementSenderSequenceDecorator increments the sequence of the signers.
type EthIncrementSenderSequenceDecorator struct {
	ak AccountKeeper
}

// NewEthIncrementSenderSequenceDecorator creates a new EthIncrementSenderSequenceDecorator.
func NewEthIncrementSenderSequenceDecorator(ak AccountKeeper) EthIncrementSenderSequenceDecorator {
	return EthIncrementSenderSequenceDecorator{
		ak: ak,
	}
}

// AnteHandle handles incrementing the sequence of the signer (i.e sender). If the transaction is a
// contract creation, the nonce will be incremented during the transaction execution and not within
// this AnteHandler decorator.
func (issd EthIncrementSenderSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// get and set account must be called with an infinite gas meter in order to prevent
	// additional gas from being deducted.
	infCtx := ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

	for i, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return ctx, stacktrace.Propagate(
				sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type %T, expected %T", tx, &evmtypes.MsgEthereumTx{}),
				"failed to cast transaction %d", i,
			)
		}

		// NOTE: on contract creation, the nonce is incremented within the EVM Create function during tx execution
		// and not previous to the state transition ¯\_(ツ)_/¯
		if msgEthTx.To() == nil {
			// contract creation, don't increment sequence on AnteHandler but on tx execution
			// continue to the next item
			continue
		}

		// increment sequence of all signers
		for _, addr := range msg.GetSigners() {
			acc := issd.ak.GetAccount(infCtx, addr)

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

			issd.ak.SetAccount(infCtx, acc)
		}
	}

	// set the original gas meter
	return next(ctx, tx, simulate)
}

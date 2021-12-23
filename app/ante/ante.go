package ante

import (
	"fmt"
	"runtime/debug"

	tmlog "github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	channelkeeper "github.com/cosmos/ibc-go/v2/modules/core/04-channel/keeper"
	ibcante "github.com/cosmos/ibc-go/v2/modules/core/ante"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

const (
	secp256k1VerifyCost uint64 = 21000
)

// NewAnteHandler returns an ante handler responsible for attempting to route an
// Ethereum or SDK transaction to an internal ante handler for performing
// transaction-level processing (e.g. fee payment, signature verification) before
// being passed onto it's respective handler.
func NewAnteHandler(
	ak evmtypes.AccountKeeper,
	bankKeeper evmtypes.BankKeeper,
	evmKeeper EVMKeeper,
	feeGrantKeeper authante.FeegrantKeeper,
	channelKeeper channelkeeper.Keeper,
	feeMarketKeeper evmtypes.FeeMarketKeeper,
	signModeHandler authsigning.SignModeHandler,
) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		defer Recover(ctx.Logger(), &err)

		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/ethermint.evm.v1.ExtensionOptionsEthereumTx":
					// handle as *evmtypes.MsgEthereumTx

					anteHandler = sdk.ChainAnteDecorators(
						NewEthSetUpContextDecorator(),                         // outermost AnteDecorator. SetUpContext must be called first
						NewEthMempoolFeeDecorator(evmKeeper, feeMarketKeeper), // Check eth effective gas price against minimal-gas-prices
						NewEthValidateBasicDecorator(evmKeeper),
						NewEthSigVerificationDecorator(evmKeeper),
						NewEthAccountVerificationDecorator(ak, bankKeeper, evmKeeper),
						NewEthNonceVerificationDecorator(ak),
						NewEthGasConsumeDecorator(evmKeeper),
						NewCanTransferDecorator(evmKeeper, feeMarketKeeper),
						NewEthIncrementSenderSequenceDecorator(ak), // innermost AnteDecorator.
					)

				default:
					return ctx, sdkerrors.Wrapf(
						sdkerrors.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s",
						typeURL,
					)
				}

				return anteHandler(ctx, tx, sim)
			}
		}

		// Reject messages that requires specific authentication here.
		// For example `MsgEthereumTx` requires fee to be deducted in the antehandler in order to perform the refund.
		for _, msg := range tx.GetMsgs() {
			if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
				return ctx, sdkerrors.Wrapf(
					sdkerrors.ErrInvalidType,
					"MsgEthereumTx needs to be contained within a tx with ExtensionOptionsEthereumTx option",
				)
			}
		}

		// handle as totally normal Cosmos SDK tx

		switch tx.(type) {
		case sdk.Tx:
			anteHandler = sdk.ChainAnteDecorators(
				authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
				authante.NewRejectExtensionOptionsDecorator(),
				authante.NewMempoolFeeDecorator(),
				authante.NewValidateBasicDecorator(),
				authante.NewTxTimeoutHeightDecorator(),
				authante.NewValidateMemoDecorator(ak),
				ibcante.NewAnteDecorator(channelKeeper),
				authante.NewConsumeGasForTxSizeDecorator(ak),
				authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
				authante.NewValidateSigCountDecorator(ak),
				authante.NewDeductFeeDecorator(ak, bankKeeper, feeGrantKeeper),
				authante.NewSigGasConsumeDecorator(ak, DefaultSigVerificationGasConsumer),
				authante.NewSigVerificationDecorator(ak, signModeHandler),
				authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
			)
		default:
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
	}
}

func Recover(logger tmlog.Logger, err *error) {
	if r := recover(); r != nil {
		*err = sdkerrors.Wrapf(sdkerrors.ErrPanic, "%v", r)

		if e, ok := r.(error); ok {
			logger.Error(
				"ante handler panicked",
				"error", e,
				"stack trace", string(debug.Stack()),
			)
		} else {
			logger.Error(
				"ante handler panicked",
				"recover", fmt.Sprintf("%v", r),
			)
		}
	}
}

var _ authante.SignatureVerificationGasConsumer = DefaultSigVerificationGasConsumer

// DefaultSigVerificationGasConsumer is the default implementation of SignatureVerificationGasConsumer. It consumes gas
// for signature verification based upon the public key type. The cost is fetched from the given params and is matched
// by the concrete type.
func DefaultSigVerificationGasConsumer(
	meter sdk.GasMeter, sig signing.SignatureV2, params authtypes.Params,
) error {
	// support for ethereum ECDSA secp256k1 keys
	_, ok := sig.PubKey.(*ethsecp256k1.PubKey)
	if ok {
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: eth_secp256k1")
		return nil
	}

	return authante.DefaultSigVerificationGasConsumer(meter, sig, params)
}

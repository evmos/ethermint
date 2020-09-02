package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/crypto"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

func init() {
	crypto.RegisterCodec(types.ModuleCdc)
}

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
func NewAnteHandler(ak auth.AccountKeeper, evmKeeper EVMKeeper, sk types.SupplyKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler
		switch tx.(type) {
		case auth.StdTx:
			anteHandler = sdk.ChainAnteDecorators(
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

		case evmtypes.MsgEthereumTx:
			anteHandler = sdk.ChainAnteDecorators(
				NewEthSetupContextDecorator(), // outermost AnteDecorator. EthSetUpContext must be called first
				NewEthMempoolFeeDecorator(evmKeeper),
				NewEthSigVerificationDecorator(),
				NewAccountVerificationDecorator(ak, evmKeeper),
				NewNonceVerificationDecorator(ak),
				NewEthGasConsumeDecorator(ak, sk, evmKeeper),
				NewIncrementSenderSequenceDecorator(ak), // innermost AnteDecorator.
			)
		default:
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "invalid transaction type: %T", tx)
		}

		return anteHandler(ctx, tx, sim)
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

// IncrementSequenceDecorator handles incrementing sequences of all signers.
// Use the IncrementSequenceDecorator decorator to prevent replay attacks. Note,
// there is no need to execute IncrementSequenceDecorator on RecheckTX since
// CheckTx would already bump the sequence number.
//
// NOTE: Since CheckTx and DeliverTx state are managed separately, subsequent and
// sequential txs orginating from the same account cannot be handled correctly in
// a reliable way unless sequence numbers are managed and tracked manually by a
// client. It is recommended to instead use multiple messages in a tx.
type IncrementSequenceDecorator struct {
	ak auth.AccountKeeper
}

func NewIncrementSequenceDecorator(ak auth.AccountKeeper) IncrementSequenceDecorator {
	return IncrementSequenceDecorator{
		ak: ak,
	}
}

func (isd IncrementSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// no need to increment sequence on RecheckTx
	if ctx.IsReCheckTx() && !simulate {
		return next(ctx, tx, simulate)
	}

	sigTx, ok := tx.(authante.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// increment sequence of all signers
	for _, addr := range sigTx.GetSigners() {
		acc := isd.ak.GetAccount(ctx, addr)
		if err := acc.SetSequence(acc.GetSequence() + 1); err != nil {
			panic(err)
		}

		isd.ak.SetAccount(ctx, acc)
	}

	return next(ctx, tx, simulate)
}

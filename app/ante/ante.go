package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	tmcrypto "github.com/tendermint/tendermint/crypto"
)

func init() {
	ethsecp256k1.RegisterCodec(types.ModuleCdc)
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
				NewAccountSetupDecorator(ak),
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
				authante.NewValidateBasicDecorator(),
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
	meter sdk.GasMeter, _ []byte, pubkey tmcrypto.PubKey, _ types.Params,
) error {
	switch pubkey.(type) {
	case ethsecp256k1.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: secp256k1")
		return nil
	case tmcrypto.PubKey:
		meter.ConsumeGas(secp256k1VerifyCost, "ante verify: tendermint secp256k1")
		return nil
	default:
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "unrecognized public key type: %T", pubkey)
	}
}

// AccountSetupDecorator sets an account to state if it's not stored already. This only applies for MsgEthermint.
type AccountSetupDecorator struct {
	ak auth.AccountKeeper
}

// NewAccountSetupDecorator creates a new AccountSetupDecorator instance
func NewAccountSetupDecorator(ak auth.AccountKeeper) AccountSetupDecorator {
	return AccountSetupDecorator{
		ak: ak,
	}
}

// AnteHandle sets an account for MsgEthermint (evm) if the sender is registered.
// NOTE: Since the account is set without any funds, the message execution will
// fail if the validator requires a minimum fee > 0.
func (asd AccountSetupDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "no messages included in transaction")
	}

	for _, msg := range msgs {
		if msgEthermint, ok := msg.(evmtypes.MsgEthermint); ok {
			setupAccount(asd.ak, ctx, msgEthermint.From)
		}
	}

	return next(ctx, tx, simulate)
}

func setupAccount(ak keeper.AccountKeeper, ctx sdk.Context, addr sdk.AccAddress) {
	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		return
	}

	acc = ak.NewAccountWithAddress(ctx, addr)
	ak.SetAccount(ctx, acc)
}

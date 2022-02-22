package ante

import (
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"github.com/tharsis/ethermint/ethereum/eip712"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Verify all signatures for a tx and return an error if any are invalid. Note,
// the Eip712SigVerificationDecorator decorator will not get executed on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type Eip712SigVerificationDecorator struct {
	ak              evmtypes.AccountKeeper
	signModeHandler authsigning.SignModeHandler
}

func NewEip712SigVerificationDecorator(ak evmtypes.AccountKeeper, signModeHandler authsigning.SignModeHandler) Eip712SigVerificationDecorator {
	return Eip712SigVerificationDecorator{
		ak:              ak,
		signModeHandler: signModeHandler,
	}
}

func (svd Eip712SigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// no need to verify signatures on recheck tx
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	signerAddrs := sigTx.GetSigners()

	// check that signer length and signature length are the same
	if len(sigs) != len(signerAddrs) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signerAddrs), len(sigs))
	}

	for i, sig := range sigs {
		acc, err := authante.GetSignerAcc(ctx, svd.ak, signerAddrs[i])
		if err != nil {
			return ctx, err
		}

		// retrieve pubkey
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		if sig.Sequence != acc.GetSequence() {
			return ctx, sdkerrors.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
		genesis := ctx.BlockHeight() == 0
		chainID := ctx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}
		signerData := authsigning.SignerData{
			ChainID:       chainID,
			AccountNumber: accNum,
			Sequence:      acc.GetSequence(),
		}

		if !simulate {
			err := VerifySignature(pubKey, signerData, sig.Data, svd.signModeHandler, tx.(authsigning.Tx))
			if err != nil {
				errMsg := fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s)", accNum, chainID)
				return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, errMsg)

			}
		}
	}

	return next(ctx, tx, simulate)
}

var ethermintCodec codec.ProtoCodecMarshaler

func init() {
	registry := codectypes.NewInterfaceRegistry()
	ethermint.RegisterInterfaces(registry)
	ethermintCodec = codec.NewProtoCodec(registry)
}

// VerifySignature verifies a transaction signature contained in SignatureData abstracting over different signing modes
// and single vs multi-signatures.
func VerifySignature(
	pubKey cryptotypes.PubKey,
	signerData authsigning.SignerData,
	sigData signing.SignatureData,
	_ authsigning.SignModeHandler,
	tx authsigning.Tx,
) error {
	switch data := sigData.(type) {
	case *signing.SingleSignatureData:
		if data.SignMode != signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
			return fmt.Errorf("unexpected SignatureData %T: wrong SignMode", sigData)
		}

		// @contract: this code is reached only when Msg has Web3Tx extension (so this custom Ante handler flow),
		// and the signature is SIGN_MODE_LEGACY_AMINO_JSON which is supported for EIP712 for now

		msgs := tx.GetMsgs()
		txBytes := legacytx.StdSignBytes(
			signerData.ChainID,
			signerData.AccountNumber,
			signerData.Sequence,
			tx.GetTimeoutHeight(),
			legacytx.StdFee{
				Amount: tx.GetFee(),
				Gas:    tx.GetGas(),
			},
			msgs, tx.GetMemo(),
		)

		var chainID uint64
		var err error

		var (
			feePayer     sdk.AccAddress
			feePayerSig  []byte
			feeDelegated bool
		)

		if txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx); ok {
			if opts := txWithExtensions.GetExtensionOptions(); len(opts) > 0 {
				var optIface ethermint.ExtensionOptionsWeb3TxI

				if err := ethermintCodec.UnpackAny(opts[0], &optIface); err != nil {
					return errors.Wrap(err, "failed to proto-unpack ExtensionOptionsWeb3Tx")
				}

				if extOpt, ok := optIface.(*ethermint.ExtensionOptionsWeb3Tx); ok {
					// chainID in EIP712 typed data is allowed to not match signerData.ChainID,
					// but limited to certain options: 1 (mainnet), 42 (Kovan), thus Metamask will
					// be able to submit signatures without switching networks.

					if extOpt.TypedDataChainID == 1 || extOpt.TypedDataChainID == 42 || extOpt.TypedDataChainID == 9000 {
						chainID = extOpt.TypedDataChainID
					}

					if len(extOpt.FeePayer) > 0 {
						feePayer, err = sdk.AccAddressFromBech32(extOpt.FeePayer)
						if err != nil {
							return errors.Wrap(err, "failed to parse feePayer from ExtensionOptionsWeb3Tx")
						}

						feePayerSig = extOpt.FeePayerSig
						if len(feePayerSig) == 0 {
							return errors.Wrap(err, "no feePayerSig provided in ExtensionOptionsWeb3Tx")
						}

						feeDelegated = true
					}
				}
			}
		}

		if chainID == 0 {
			chainID, err = strconv.ParseUint(signerData.ChainID, 10, 64)
			if err != nil {
				return errors.Wrapf(err, "failed to parse chainID: %s", signerData.ChainID)
			}
		}

		var typedData apitypes.TypedData
		var sigHash []byte

		if feeDelegated {
			feeDelegation := &eip712.FeeDelegationOptions{
				FeePayer: feePayer,
			}

			typedData, err = eip712.WrapTxToTypedData(ethermintCodec, chainID, msgs[0], txBytes, feeDelegation)
			if err != nil {
				return errors.Wrap(err, "failed to pack tx data in EIP712 object")
			}

			sigHash, err = eip712.ComputeTypedDataHash(typedData)
			if err != nil {
				return err
			}

			if len(feePayerSig) != 65 {
				return fmt.Errorf("signature length doesn't match typical [R||S||V] signature 65 bytes")
			}
			if feePayerSig[64] > 4 {
				// Remove the recovery offset if needed (ie. Metamask eip712 signature)
				feePayerSig[64] -= 27
			}

			feePayerPubkey, err := secp256k1.RecoverPubkey(sigHash, feePayerSig)
			if err != nil {
				return errors.Wrap(err, "failed to recover delegated fee payer from sig")
			}

			ecPubKey, err := ethcrypto.UnmarshalPubkey(feePayerPubkey)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal recovered fee payer pubkey")
			}

			pk := &ethsecp256k1.PubKey{
				Key: ethcrypto.CompressPubkey(ecPubKey),
			}

			recoveredFeePayerAcc := sdk.AccAddress(pk.Address().Bytes())

			if !recoveredFeePayerAcc.Equals(feePayer) {
				return errors.New("failed to verify delegated fee payer sig")
			}

			// Overwrite the transaction signature because we are using EIP712
			// TODO: should we allow only feePayerSignature and return error
			// if the transaction signatures != []?
			data.Signature = feePayerSig
		} else {
			typedData, err = eip712.WrapTxToTypedData(ethermintCodec, chainID, msgs[0], txBytes, nil)
			if err != nil {
				return errors.Wrap(err, "failed to pack tx data in EIP712 object")
			}

			sigHash, err = eip712.ComputeTypedDataHash(typedData)
			if err != nil {
				return err
			}
		}

		if len(data.Signature) != 65 {
			return fmt.Errorf("signature length doesn't match typical [R||S||V] signature 65 bytes")
		}

		// VerifySignature of ethsecp256k1 accepts 64 byte signature [R||S]
		// WARNING! Under NO CIRCUMSTANCES try to use pubKey.VerifySignature there
		if !secp256k1.VerifySignature(pubKey.Bytes(), sigHash, data.Signature[:len(data.Signature)-1]) {
			return fmt.Errorf("unable to verify signer signature of EIP712 typed data")
		}

		return nil
	default:
		return fmt.Errorf("unexpected SignatureData %T", sigData)
	}
}

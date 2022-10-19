package eip712_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/evmos/ethermint/ethereum/eip712"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"

	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Tests single-signer EIP-712 signature verification. Multi-signer verification tests are included
// in ante/integration test files.

var config = encoding.MakeConfig(app.ModuleBasics)
var clientCtx = client.Context{}.WithTxConfig(config.TxConfig)

// Set up test env to replicate prod. environment
func setupTestEnv(t *testing.T) {
	t.Helper()

	sdk.GetConfig().SetBech32PrefixForAccount("ethm", "")
	eip712.SetEncodingConfig(config)
}

// Helper to create random test addresses for messages
func createTestAddress(t *testing.T) sdk.AccAddress {
	t.Helper()

	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	require.NoError(t, err)

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return addr.Bytes()
}

// Helper to create random keypair for signing + verification
func createTestKeyPair(t *testing.T) (*ethsecp256k1.PrivKey, *ethsecp256k1.PubKey) {
	t.Helper()

	privKey, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)

	pubKey := &ethsecp256k1.PubKey{
		Key: privKey.PubKey().Bytes(),
	}
	require.Implements(t, (*cryptotypes.PubKey)(nil), pubKey)

	return privKey, pubKey
}

// Helper to create instance of sdk.Coins[] with single coin
func makeCoins(t *testing.T, denom string, amount math.Int) sdk.Coins {
	t.Helper()

	return sdk.NewCoins(
		sdk.NewCoin(
			denom,
			amount,
		),
	)
}

func TestEIP712SignatureVerification(t *testing.T) {
	setupTestEnv(t)

	signModes := []signing.SignMode{
		signing.SignMode_SIGN_MODE_DIRECT,
		signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}

	testCases := []struct {
		title         string
		fee           txtypes.Fee
		memo          string
		msgs          []sdk.Msg
		accountNumber uint64
		sequence      uint64
		tip           txtypes.Tip
		expectSuccess bool
	}{
		{
			title: "Standard MsgSend",
			fee: txtypes.Fee{
				Amount:   makeCoins(t, "aphoton", math.NewInt(2000)),
				GasLimit: 20000,
				Payer:    sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			memo: "",
			msgs: []sdk.Msg{
				banktypes.NewMsgSend(
					createTestAddress(t),
					createTestAddress(t),
					makeCoins(t, "photon", math.NewInt(1)),
				),
			},
			accountNumber: 8,
			sequence:      5,
			tip: txtypes.Tip{
				Amount: makeCoins(t, "aphoton", math.NewInt(20000)),
				Tipper: sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			expectSuccess: true,
		},
		{
			title: "Standard MsgVote",
			fee: txtypes.Fee{
				Amount:   makeCoins(t, "aphoton", math.NewInt(2000)),
				GasLimit: 20000,
				Payer:    sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					createTestAddress(t),
					5,
					govtypes.OptionNo,
				),
			},
			accountNumber: 25,
			sequence:      78,
			tip: txtypes.Tip{
				Amount: makeCoins(t, "aphoton", math.NewInt(500000)),
				Tipper: sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			expectSuccess: true,
		},
		{
			title: "Two MsgVotes",
			fee: txtypes.Fee{
				Amount:   makeCoins(t, "aphoton", math.NewInt(2000)),
				GasLimit: 20000,
				Payer:    sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					createTestAddress(t),
					5,
					govtypes.OptionNo,
				),
				govtypes.NewMsgVote(
					createTestAddress(t),
					25,
					govtypes.OptionAbstain,
				),
			},
			accountNumber: 25,
			sequence:      78,
			tip: txtypes.Tip{
				Amount: makeCoins(t, "aphoton", math.NewInt(500000)),
				Tipper: sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			expectSuccess: false, // Multiple messages (check for multiple signers in AnteHandler)
		},
		{
			title: "MsgSend + MsgVote",
			fee: txtypes.Fee{
				Amount:   makeCoins(t, "aphoton", math.NewInt(2000)),
				GasLimit: 20000,
				Payer:    sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			memo: "",
			msgs: []sdk.Msg{
				govtypes.NewMsgVote(
					createTestAddress(t),
					5,
					govtypes.OptionNo,
				),
				banktypes.NewMsgSend(
					createTestAddress(t),
					createTestAddress(t),
					makeCoins(t, "photon", math.NewInt(50)),
				),
			},
			accountNumber: 25,
			sequence:      78,
			tip: txtypes.Tip{
				Amount: makeCoins(t, "aphoton", math.NewInt(500000)),
				Tipper: sdk.MustBech32ifyAddressBytes("ethm", createTestAddress(t)),
			},
			expectSuccess: false, // Multiple messages
		},
	}

	for _, tc := range testCases {
		for _, signMode := range signModes {
			privKey, pubKey := createTestKeyPair(t)

			// Init tx builder
			txBuilder := clientCtx.TxConfig.NewTxBuilder()

			// Set fees
			txBuilder.SetGasLimit(tc.fee.GasLimit)
			txBuilder.SetFeePayer(sdk.MustAccAddressFromBech32(tc.fee.Payer))
			txBuilder.SetFeeAmount(tc.fee.Amount)

			// Set messages
			err := txBuilder.SetMsgs(tc.msgs...)
			require.NoError(t, err)

			// Set memo and tip
			txBuilder.SetMemo(tc.memo)
			txBuilder.SetTip(&tc.tip)

			// Prepare signature field
			txSigData := signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: nil,
			}
			txSig := signing.SignatureV2{
				PubKey:   pubKey,
				Data:     &txSigData,
				Sequence: tc.sequence,
			}

			err = txBuilder.SetSignatures([]signing.SignatureV2{
				txSig,
			}...)
			require.NoError(t, err)

			// Declare signerData
			signerData := authsigning.SignerData{
				ChainID:       "ethermint_9000-1",
				AccountNumber: tc.accountNumber,
				Sequence:      tc.sequence,
				PubKey:        pubKey,
				Address:       sdk.MustBech32ifyAddressBytes("ethm", pubKey.Bytes()),
			}

			bz, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(
				signMode,
				signerData,
				txBuilder.GetTx(),
			)
			require.NoError(t, err)

			verifyEIP712SignatureVerification(t, tc.expectSuccess, *privKey, *pubKey, bz)
		}
	}
}

// Verify that the payload passes signature verification if signed as its EIP-712 representation.
func verifyEIP712SignatureVerification(t *testing.T, expectedSuccess bool, privKey ethsecp256k1.PrivKey, pubKey ethsecp256k1.PubKey, signBytes []byte) {
	t.Helper()

	// Convert to EIP712 hash and sign
	eip712Hash, err := eip712.GetEIP712HashForMsg(signBytes)
	if expectedSuccess {
		require.NoError(t, err)
	} else {
		// Expect failure generating EIP-712 hash
		require.Error(t, err)
		return
	}
	sigHash := crypto.Keccak256Hash(eip712Hash)
	sig, err := privKey.Sign(sigHash.Bytes())
	require.NoError(t, err)

	// Verify against original payload bytes. This should pass, even though it is not
	// the original message that was signed.
	res := pubKey.VerifySignature(signBytes, sig)
	require.True(t, res)

	// Verify against the signed EIP-712 bytes. This should pass, since it is the message signed.
	res = pubKey.VerifySignature(eip712Hash, sig)
	require.True(t, res)

	// Verify against random bytes to ensure it does not pass unexpectedly (sanity check).
	randBytes := make([]byte, len(signBytes))
	copy(randBytes, signBytes)
	randBytes[0] = (signBytes[0] + 10) % 128
	res = pubKey.VerifySignature(randBytes, sig)
	require.False(t, res)
}

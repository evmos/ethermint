package eip712_test

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/app"
	"github.com/evmos/ethermint/encoding"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/tests"
	"github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Test standard payload preprocessing for Ledger EIP-712-signed transactions
func TestPreprocessLedger(t *testing.T) {
	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	ctx := client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	_, privKey := tests.NewAddrKey()
	sdk.GetConfig().SetBech32PrefixForAccount("ethm", "")
	chainId := "ethermint_9000-1"

	txBuilder := ctx.TxConfig.NewTxBuilder()

	txBuilder.SetFeeAmount(sdk.NewCoins(
		sdk.NewCoin(
			evmtypes.DefaultParams().EvmDenom,
			math.NewInt(2000)),
	))

	feePayerAddress := "ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl"
	feePayer, err := sdk.AccAddressFromBech32(feePayerAddress)
	if err != nil {
		t.Errorf("Invalid bech32 address: %v", err)
	}

	txBuilder.SetFeePayer(feePayer)
	txBuilder.SetGasLimit(200000)
	txBuilder.SetMemo("")

	msgSend := banktypes.MsgSend{
		FromAddress: feePayerAddress,
		ToAddress:   "ethm12luku6uxehhak02py4rcz65zu0swh7wjun6msa",
		Amount: sdk.NewCoins(
			sdk.NewCoin(
				evmtypes.DefaultParams().EvmDenom,
				math.NewInt(10000000),
			),
		),
	}

	txBuilder.SetMsgs(&msgSend)

	// Create signature unrelated to payload for testing
	signatureHex := strings.Repeat("01", 65)
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		t.Errorf("Could not decode hex bytes: %v", err)
	}

	sigsV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(), // Use unrelated public key for testing
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signatureBytes,
		},
		Sequence: 0,
	}

	txBuilder.SetSignatures(sigsV2)

	// Run pre-processing
	err = eip712.PreprocessLedgerTx(
		chainId,
		keyring.TypeLedger,
		txBuilder,
	)

	if err != nil {
		t.Errorf("Could not preprocess Ledger Tx: %v", err)
	}

	// Verify Web3 Extension
	hasExtOptsTx, ok := txBuilder.(ante.HasExtensionOptionsTx)
	if !ok {
		t.Errorf("Tx does not have extension after preprocessing as Ledger")
	}

	hasOneExt := len(hasExtOptsTx.GetExtensionOptions()) == 1
	if !hasOneExt {
		t.Errorf("Invalid number of extensions, expected 1")
	}

	expectedExt := types.ExtensionOptionsWeb3Tx{
		TypedDataChainID: 9000,
		FeePayer:         feePayerAddress,
		FeePayerSig:      signatureBytes,
	}

	expectedExtAny, err := codectypes.NewAnyWithValue(&expectedExt)
	if err != nil {
		t.Errorf("Could not decode extension with err: %v", err)
	}

	extensionAny := hasExtOptsTx.GetExtensionOptions()[0]

	if !reflect.DeepEqual(expectedExtAny, extensionAny) {
		t.Errorf(
			"Extension does not match expected:\n Expected: %v\n, Actual: %v\n", expectedExtAny.GetCachedValue(), extensionAny.GetCachedValue(),
		)
	}

	// Verify signature type
	formattedSigs, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		t.Errorf("Could not get signatures from Tx Builder: %v", err)
	}

	if len(formattedSigs) != 1 {
		t.Errorf("Invalid number of signatures from Tx Builder, expected 1 but got %v", len(formattedSigs))
	}

	formattedSig := formattedSigs[0].Data.(*signing.SingleSignatureData)
	if formattedSig.SignMode != signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON {
		t.Errorf("Invalid sign mode, expected LEGACY_AMINO_JSON")
	}

	// Verify signature is blank
	if len(formattedSig.Signature) != 0 {
		t.Errorf("Expected blank signature, but got %v", formattedSig)
	}

	// Verify tx fields are unchanged
	tx := txBuilder.GetTx()

	if tx.FeePayer().String() != feePayer.String() {
		t.Errorf("Fee payer changed in Tx Builder")
	}

	if tx.GetGas() != 200000 {
		t.Errorf("Gas field changed in Tx Builder")
	}

	if tx.GetFee().AmountOf("aphoton") != math.NewInt(2000) {
		t.Errorf("Fee amount changed in Tx Builder")
	}

	if tx.GetMemo() != "" {
		t.Errorf("Memo changed in Tx Builder")
	}

	if len(tx.GetMsgs()) != 1 || tx.GetMsgs()[0].String() != msgSend.String() {
		t.Errorf("Messages changed in Tx Builder")
	}
}

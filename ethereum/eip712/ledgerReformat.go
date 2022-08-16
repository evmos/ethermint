package eip712

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cosmoskr "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/evmos/ethermint/types"
)

// ReformatLedgerTx reformats Ledger-signed Cosmos transactions to match the fork expected by Evmos
// by including the signature in a Web3Tx extension.
func ReformatLedgerTx(chainID string, keyType cosmoskr.KeyType, txBuilder client.TxBuilder) error {
	if keyType != cosmoskr.TypeLedger {
		// Only process Ledger transactions
		return nil
	}

	// Init extension builder to set Web3 extension
	extensionBuilder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return fmt.Errorf("Error setting extension options: cannot cast to ExtensionOptionsTxBuilder")
	}

	// Get signatures from TxBuilder
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		return fmt.Errorf("Could not get signatures with error: %w", err)
	}

	// Verify single-signer
	if len(sigs) != 1 {
		return fmt.Errorf("Invalid number of signatures, expected 1 and got %v", len(sigs))
	}

	signature := sigs[0]
	sigBytes := signature.Data.(*signing.SingleSignatureData).Signature

	// Parse Chain ID as big.Int
	chainIDInt, err := types.ParseChainID(chainID)
	if err != nil {
		return fmt.Errorf("Error parsing chain id: %v\n", err)
	}

	// Add ExtensionOptionsWeb3Tx extension with signature
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&types.ExtensionOptionsWeb3Tx{
		FeePayer:         signature.PubKey.Address().String(),
		TypedDataChainID: chainIDInt.Uint64(),
		FeePayerSig:      sigBytes,
	})
	extensionBuilder.SetExtensionOptions(option)

	// Set signature data with to indicate Amino Sign Type
	// (Regardless of input signMode, Evmos requires Amino signature type for Ledger)
	blankSig := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   signature.PubKey,
		Data:     &blankSig,
		Sequence: signature.Sequence,
	}

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return fmt.Errorf("Unable to set signatures on payload: %v\n", err)
	}

	return nil
}

// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
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

// PreprocessLedgerTx reformats Ledger-signed Cosmos transactions to match the fork expected by Ethermint
// by including the signature in a Web3Tx extension and sending a blank signature in the body.
func PreprocessLedgerTx(chainID string, keyType cosmoskr.KeyType, txBuilder client.TxBuilder) error {
	// Only process Ledger transactions
	if keyType != cosmoskr.TypeLedger {
		return nil
	}

	// Init extension builder to set Web3 extension
	extensionBuilder, ok := txBuilder.(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		return fmt.Errorf("cannot cast TxBuilder to ExtensionOptionsTxBuilder")
	}

	// Get signatures from TxBuilder
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		return fmt.Errorf("could not get signatures: %w", err)
	}

	// Verify single-signer
	if len(sigs) != 1 {
		return fmt.Errorf("invalid number of signatures, expected 1 and got %v", len(sigs))
	}

	signature := sigs[0]
	sigData, ok := signature.Data.(*signing.SingleSignatureData)
	if !ok {
		return fmt.Errorf("unexpected signature type, expected SingleSignatureData")
	}
	sigBytes := sigData.Signature

	// Parse Chain ID as big.Int
	chainIDInt, err := types.ParseChainID(chainID)
	if err != nil {
		return fmt.Errorf("could not parse chain id: %w", err)
	}

	// Add ExtensionOptionsWeb3Tx extension with signature
	var option *codectypes.Any
	option, err = codectypes.NewAnyWithValue(&types.ExtensionOptionsWeb3Tx{
		FeePayer:         txBuilder.GetTx().FeePayer().String(),
		TypedDataChainID: chainIDInt.Uint64(),
		FeePayerSig:      sigBytes,
	})
	if err != nil {
		return fmt.Errorf("could not set extension as any: %w", err)
	}

	extensionBuilder.SetExtensionOptions(option)

	// Set blank signature with Amino Sign Type
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
		return fmt.Errorf("unable to set signatures on payload: %w", err)
	}

	return nil
}

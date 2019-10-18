package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Ethermint error codes
const (
	// DefaultCodespace reserves a Codespace for Ethermint.
	DefaultCodespace sdk.CodespaceType = "ethermint"

	CodeInvalidValue   sdk.CodeType = 1
	CodeInvalidChainID sdk.CodeType = 2
	CodeInvalidSender  sdk.CodeType = 3
	CodeVMExecution    sdk.CodeType = 4
	CodeInvalidNonce   sdk.CodeType = 5
)

// CodeToDefaultMsg takes the CodeType variable and returns the error string
func CodeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidValue:
		return "invalid value"
	case CodeInvalidChainID:
		return "invalid chain ID"
	case CodeInvalidSender:
		return "could not derive sender from transaction"
	case CodeVMExecution:
		return "error while executing evm transaction"
	case CodeInvalidNonce:
		return "invalid nonce"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

// ErrInvalidValue returns a standardized SDK error resulting from an invalid value.
func ErrInvalidValue(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidValue, msg)
}

// ErrInvalidChainID returns a standardized SDK error resulting from an invalid chain ID.
func ErrInvalidChainID(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidChainID, msg)
}

// ErrInvalidSender returns a standardized SDK error resulting from an invalid transaction sender.
func ErrInvalidSender(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidSender, msg)
}

// ErrVMExecution returns a standardized SDK error resulting from an error in EVM execution.
func ErrVMExecution(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeVMExecution, msg)
}

// ErrVMExecution returns a standardized SDK error resulting from an error in EVM execution.
func ErrInvalidNonce(msg string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeInvalidNonce, msg)
}

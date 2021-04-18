package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NOTE: We can't use 1 since that error code is reserved for internal errors.

var (
	// ErrInvalidState returns an error resulting from an invalid Storage State.
	ErrInvalidState = sdkerrors.Register(ModuleName, 2, "invalid storage state")

	// ErrChainConfigNotFound returns an error if the chain config cannot be found on the store.
	ErrChainConfigNotFound = sdkerrors.Register(ModuleName, 3, "chain configuration not found")

	// ErrInvalidChainConfig returns an error resulting from an invalid ChainConfig.
	ErrInvalidChainConfig = sdkerrors.Register(ModuleName, 4, "invalid chain configuration")

	// ErrZeroAddress returns an error resulting from an zero (empty) ethereum Address.
	ErrZeroAddress = sdkerrors.Register(ModuleName, 5, "invalid zero address")

	// ErrEmptyHash returns an error resulting from an empty ethereum Hash.
	ErrEmptyHash = sdkerrors.Register(ModuleName, 6, "empty hash")

	// ErrBloomNotFound returns an error if the block bloom cannot be found on the store.
	ErrBloomNotFound = sdkerrors.Register(ModuleName, 7, "block bloom not found")

	// ErrInvalidValue returns an error resulting from an invalid value.
	ErrInvalidValue = sdkerrors.Register(ModuleName, 8, "invalid value")

	// ErrInvalidChainID returns an error resulting from an invalid chain ID.
	ErrInvalidChainID = sdkerrors.Register(ModuleName, 9, "invalid chain ID")

	// ErrVMExecution returns an error resulting from an error in EVM execution.
	ErrVMExecution = sdkerrors.Register(ModuleName, 10, "error while executing evm transaction")

	// ErrTxReceiptNotFound returns an error if the transaction receipt could not be found
	ErrTxReceiptNotFound = sdkerrors.Register(ModuleName, 11, "transaction receipt not found")

	// ErrCreateDisabled returns an error if the EnableCreate parameter is false.
	ErrCreateDisabled = sdkerrors.Register(ModuleName, 12, "EVM Create operation is disabled")

	// ErrCallDisabled returns an error if the EnableCall parameter is false.
	ErrCallDisabled = sdkerrors.Register(ModuleName, 13, "EVM Call operation is disabled")
)

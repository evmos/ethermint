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
)

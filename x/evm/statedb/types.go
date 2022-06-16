package statedb

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ExtState manage some extended states which needs to be committed together with StateDB.
// mainly for the stateful precompiled contracts.
type ExtState interface {
	// write the dirty states to cosmos-sdk storage
	Commit(sdk.Context) error
}

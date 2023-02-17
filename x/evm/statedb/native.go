package statedb

import (
	"github.com/cosmos/cosmos-sdk/store/types"
	"github.com/ethereum/go-ethereum/common"
)

var _ JournalEntry = nativeChange{}

type nativeChange struct {
	snapshot types.MultiStore
}

func (native nativeChange) Dirtied() *common.Address {
	return nil
}

func (native nativeChange) Revert(s *StateDB) {
	s.restoreNativeState(native.snapshot)
}

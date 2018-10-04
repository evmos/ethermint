package state

import (
	ethstate "github.com/ethereum/go-ethereum/core/state"
)

// RawDump returns a raw state drump.
//
// TODO: Implement if we need it, especially for the RPC API.
func (csdb *CommitStateDB) RawDump() ethstate.Dump {
	return ethstate.Dump{}
}

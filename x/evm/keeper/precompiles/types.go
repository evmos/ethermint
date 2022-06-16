package precompiles

import "github.com/evmos/ethermint/x/evm/statedb"

// ExtStateDB defines extra methods of statedb to support stateful precompiled contracts
type ExtStateDB interface {
	AppendJournalEntry(statedb.JournalEntry)
}

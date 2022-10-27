package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/x/evm/statedb"
)

// stateful precompiles can embed the BasePrecompile to avoid explicitly implementing Run
type BasePrecompile struct{}

func (bpc *BasePrecompile) Run(input []byte) ([]byte, error) {
	return []byte{}, nil
}

// holds dirty state of stateful precompile
type PrecompileJournalEntry struct {
	cacheCtx     sdk.Context
	commit       func() // only needed for the last valid one
	extStateDB   statedb.ExtStateDB
	journalIndex int
}

// NewPrecompileJournalEntry gets the last valid, non-reverted cosmos sdk context from the statedb and
// uses CacheContext to generate a temporary context
func NewPrecompileJournalEntry(stateDB vm.StateDB) (*PrecompileJournalEntry, error) {
	extStateDB, ok := stateDB.(statedb.ExtStateDB)
	if !ok {
		return &PrecompileJournalEntry{}, errors.New("statedb not of external type")
	}

	eje, ok := extStateDB.GetLastValidExtJournalEntry()
	var (
		cacheContext sdk.Context
		commitFunc   func()
	)
	if !ok {
		// use the original context passed in at the beginning of transaction
		cacheContext, commitFunc = extStateDB.GetContext().CacheContext()
	} else {
		ps, ok := eje.(*PrecompileJournalEntry)
		if !ok {
			return &PrecompileJournalEntry{}, errors.New("statedb using different external journal entry")
		}
		cacheContext, commitFunc = ps.GetCacheCtx().CacheContext()
	}

	return &PrecompileJournalEntry{
		cacheCtx:   cacheContext,
		commit:     commitFunc,
		extStateDB: extStateDB,
	}, nil
}

// GetCacheCtx returns the cached context, to be used in the stateful precompile
func (ps *PrecompileJournalEntry) GetCacheCtx() sdk.Context { return ps.cacheCtx }

// implements ExtJournalEntry
func (ps *PrecompileJournalEntry) AppendToJournal() {
	ps.journalIndex = ps.extStateDB.AppendExtJournalEntry(ps)
}

// implements ExtJournalEntry
func (ps *PrecompileJournalEntry) GetJournalIndex() int { return ps.journalIndex }

// implements ExtJournalEntry
func (ps *PrecompileJournalEntry) Commit() { ps.commit() }

// implements ExtJournalEntry
// garbage collector will delete the unneeded context
func (ps *PrecompileJournalEntry) Revert(s *statedb.StateDB) { s.RevertExtJournalEntry(ps) }

// implements ExtJournalEntry
// nil returned because all dirtied states are handled by cache ctx
func (ps *PrecompileJournalEntry) Dirtied() *common.Address { return nil }

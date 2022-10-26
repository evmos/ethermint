package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/x/evm/statedb"
)

type BasePrecompile struct{}

func (bpc *BasePrecompile) Run(input []byte) ([]byte, error) {
	return []byte{}, nil
}

// ExtJournalEntry is used for Stateful Precompiles
type ExtJournalEntry interface {
	statedb.JournalEntry

	// Commit will commit the dirtied state to Cosmos stores
	Commit()
}

// holds dirty state of stateful precompile
type PrecompileContext struct {
	cacheCtx   sdk.Context
	commit     func()
	extStateDB statedb.ExtStateDB
}

// NewPrecompileContext uses CacheContext to generate a temporary context and returns
// a PrecompileContext struct, which can be used as a journal entry
func NewPrecompileContext(stateDB vm.StateDB) (*PrecompileContext, error) {
	extStateDB, ok := stateDB.(statedb.ExtStateDB)
	if !ok {
		return &PrecompileContext{}, errors.New("statedb not of ethermint type")
	}
	cacheCtx, commit := extStateDB.GetContext().CacheContext()
	return &PrecompileContext{
		cacheCtx:   cacheCtx,
		commit:     commit,
		extStateDB: extStateDB,
	}, nil
}

// GetCacheCtx returns the cached context, to be used in the stateful precompile
func (ps *PrecompileContext) GetCacheCtx() sdk.Context { return ps.cacheCtx }

// AppendToJournal adds the dirtied state to the statedb journal
func (ps *PrecompileContext) AppendToJournal() { ps.extStateDB.AppendJournalEntry(ps) }

// implements ExtJournalEntry
func (ps *PrecompileContext) Commit() { ps.commit() }

// implements ExtJournalEntry
// garbage collector will delete the unneeded context
func (ps *PrecompileContext) Revert(*statedb.StateDB) {}

// implements ExtJournalEntry
// nil returned because all dirtied states are handled by cache ctx
func (ps *PrecompileContext) Dirtied() *common.Address { return nil }

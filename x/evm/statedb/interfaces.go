package statedb

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// ExtStateDB defines an extension to the interface provided by the go-ethereum
// codebase to support additional state transition functionalities. In particular
// it supports appending a new entry to the state journal through
// AppendJournalEntry so that the state can be reverted after running
// stateful precompiled contracts.
type ExtStateDB interface {
	vm.StateDB

	// GetContext returns the cosmos sdk context passed in to the statedb when initiating a tx
	GetContext() sdk.Context

	// AppendJournalEntry allows external module to append external journal entry, to support
	// snapshot revert for external states. Returns the index in the journal at which the external
	// journal entry is stored
	AppendExtJournalEntry(ExtJournalEntry) int

	// RevertJournalEntry will remove the external journal entry from the underlying
	// statedb so the last valid external journal entry can be updated
	// should be called by Revert() for ExtJournalEntry types
	RevertExtJournalEntry(ExtJournalEntry)

	// GetLastValidExtJournalEntry retrieves the last valid external journal entry for external
	// states if it exists
	GetLastValidExtJournalEntry() (ExtJournalEntry, bool)
}

// ExtJournalEntry is used to store external state in the ethermint statedb journal
type ExtJournalEntry interface {
	JournalEntry

	// Commit will commit the dirtied state to Cosmos stores. Only the last valid external journal
	// entry is committed to Cosmos, i.e. this function is only called once per transaction
	Commit()

	// AppendToJournal adds the dirtied state to the statedb journal returns the index in the
	// in the statedb journal in which this journal entry is stored. Should be called after
	// external state is written to
	AppendToJournal()

	// GetJournalIndex returns the index in the statedb journal in which this journal entry is
	// stored
	GetJournalIndex() int
}

// Keeper provide underlying storage of StateDB
type Keeper interface {
	// Read methods
	GetAccount(ctx sdk.Context, addr common.Address) *Account
	GetState(ctx sdk.Context, addr common.Address, key common.Hash) common.Hash
	GetCode(ctx sdk.Context, codeHash common.Hash) []byte
	// the callback returns false to break early
	ForEachStorage(ctx sdk.Context, addr common.Address, cb func(key, value common.Hash) bool)

	// Write methods, only called by `StateDB.Commit()`
	SetAccount(ctx sdk.Context, addr common.Address, account Account) error
	SetState(ctx sdk.Context, addr common.Address, key common.Hash, value []byte)
	SetCode(ctx sdk.Context, codeHash []byte, code []byte)
	DeleteAccount(ctx sdk.Context, addr common.Address) error
}

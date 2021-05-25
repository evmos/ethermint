package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var ripemd = ethcmn.HexToAddress("0000000000000000000000000000000000000003")

// journalEntry is a modification entry in the state change journal that can be
// reverted on demand.
type journalEntry interface {
	// revert undoes the changes introduced by this journal entry.
	revert(*CommitStateDB)

	// dirtied returns the Ethereum address modified by this journal entry.
	dirtied() *ethcmn.Address
}

// journal contains the list of state modifications applied since the last state
// commit. These are tracked to be able to be reverted in case of an execution
// exception or revertal request.
type journal struct {
	entries               []journalEntry         // Current changes tracked by the journal
	dirties               []dirty                // Dirty accounts and the number of changes
	addressToJournalIndex map[ethcmn.Address]int // map from address to the index of the dirties slice
}

// dirty represents a single key value pair of the journal dirties, where the
// key correspons to the account address and the value to the number of
// changes for that account.
type dirty struct {
	address ethcmn.Address
	changes int
}

// newJournal create a new initialized journal.
func newJournal() *journal {
	return &journal{
		dirties:               []dirty{},
		addressToJournalIndex: make(map[ethcmn.Address]int),
	}
}

// append inserts a new modification entry to the end of the change journal.
func (j *journal) append(entry journalEntry) {
	j.entries = append(j.entries, entry)
	if addr := entry.dirtied(); addr != nil {
		j.addDirty(*addr)
	}
}

// revert undoes a batch of journalled modifications along with any reverted
// dirty handling too.
func (j *journal) revert(statedb *CommitStateDB, snapshot int) {
	for i := len(j.entries) - 1; i >= snapshot; i-- {
		// Undo the changes made by the operation
		j.entries[i].revert(statedb)

		// Drop any dirty tracking induced by the change
		if addr := j.entries[i].dirtied(); addr != nil {
			j.substractDirty(*addr)
			if j.getDirty(*addr) == 0 {
				j.deleteDirty(*addr)
			}
		}
	}
	j.entries = j.entries[:snapshot]
}

// dirty explicitly sets an address to dirty, even if the change entries would
// otherwise suggest it as clean. This method is an ugly hack to handle the RIPEMD
// precompile consensus exception.
func (j *journal) dirty(addr ethcmn.Address) {
	j.addDirty(addr)
}

// length returns the current number of entries in the journal.
func (j *journal) length() int {
	return len(j.entries)
}

// getDirty returns the dirty count for a given address. If the address is not
// found it returns 0.
func (j *journal) getDirty(addr ethcmn.Address) int {
	idx, found := j.addressToJournalIndex[addr]
	if !found {
		return 0
	}

	return j.dirties[idx].changes
}

// addDirty adds 1 to the dirty count of an address. If the dirty entry is not
// found it creates it.
func (j *journal) addDirty(addr ethcmn.Address) {
	idx, found := j.addressToJournalIndex[addr]
	if !found {
		j.dirties = append(j.dirties, dirty{address: addr, changes: 0})
		idx = len(j.dirties) - 1
		j.addressToJournalIndex[addr] = idx
	}

	j.dirties[idx].changes++
}

// substractDirty subtracts 1 to the dirty count of an address. It performs a
// no-op if the address is not found.
func (j *journal) substractDirty(addr ethcmn.Address) {
	idx, found := j.addressToJournalIndex[addr]
	if !found {
		return
	}

	if j.dirties[idx].changes == 0 {
		return
	}
	j.dirties[idx].changes--
}

// deleteDirty deletes a dirty entry from the jounal's dirties slice. If the
// entry is not found it performs a no-op.
func (j *journal) deleteDirty(addr ethcmn.Address) {
	idx, found := j.addressToJournalIndex[addr]
	if !found {
		return
	}

	j.dirties = append(j.dirties[:idx], j.dirties[idx+1:]...)
	delete(j.addressToJournalIndex, addr)
}

type (
	// Changes to the account trie.
	createObjectChange struct {
		account *ethcmn.Address
	}

	resetObjectChange struct {
		prev *stateObject
	}

	suicideChange struct {
		account     *ethcmn.Address
		prev        bool // whether account had already suicided
		prevBalance sdk.Int
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *ethcmn.Address
		prev    sdk.Int
	}

	nonceChange struct {
		account *ethcmn.Address
		prev    uint64
	}

	storageChange struct {
		account        *ethcmn.Address
		key, prevValue ethcmn.Hash
	}

	codeChange struct {
		account            *ethcmn.Address
		prevCode, prevHash []byte
	}

	// Changes to other state values.
	refundChange struct {
		prev uint64
	}

	addLogChange struct {
		txhash ethcmn.Hash
	}

	addPreimageChange struct {
		hash ethcmn.Hash
	}

	touchChange struct {
		account *ethcmn.Address
		// prev      bool
		// prevDirty bool
	}
	accessListAddAccountChange struct {
		address *ethcmn.Address
	}
	accessListAddSlotChange struct {
		address *ethcmn.Address
		slot    *ethcmn.Hash
	}
)

func (ch createObjectChange) revert(s *CommitStateDB) {
	delete(s.stateObjectsDirty, *ch.account)

	idx, exists := s.addressToObjectIndex[*ch.account]
	if !exists {
		// perform no-op
		return
	}

	// remove from the slice
	delete(s.addressToObjectIndex, *ch.account)

	// if the slice contains one element, delete it
	if len(s.stateObjects) == 1 {
		s.stateObjects = []stateEntry{}
		return
	}

	// move the elements one position left on the array
	for i := idx + 1; i < len(s.stateObjects); i++ {
		s.stateObjects[i-1] = s.stateObjects[i]
		// the new index is i - 1
		s.addressToObjectIndex[s.stateObjects[i].address] = i - 1
	}

	//  finally, delete the last element of the slice to account for the removed object
	s.stateObjects = s.stateObjects[:len(s.stateObjects)-1]
}

func (ch createObjectChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch resetObjectChange) revert(s *CommitStateDB) {
	s.setStateObject(ch.prev)
}

func (ch resetObjectChange) dirtied() *ethcmn.Address {
	return nil
}

func (ch suicideChange) revert(s *CommitStateDB) {
	so := s.getStateObject(*ch.account)
	if so != nil {
		so.suicided = ch.prev
		so.setBalance(ch.prevBalance)
	}
}

func (ch suicideChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch touchChange) revert(s *CommitStateDB) {
}

func (ch touchChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch balanceChange) revert(s *CommitStateDB) {
	s.getStateObject(*ch.account).setBalance(ch.prev)
}

func (ch balanceChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch nonceChange) revert(s *CommitStateDB) {
	s.getStateObject(*ch.account).setNonce(ch.prev)
}

func (ch nonceChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch codeChange) revert(s *CommitStateDB) {
	s.getStateObject(*ch.account).setCode(ethcmn.BytesToHash(ch.prevHash), ch.prevCode)
}

func (ch codeChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch storageChange) revert(s *CommitStateDB) {
	s.getStateObject(*ch.account).setState(ch.key, ch.prevValue)
}

func (ch storageChange) dirtied() *ethcmn.Address {
	return ch.account
}

func (ch refundChange) revert(s *CommitStateDB) {
	s.refund = ch.prev
}

func (ch refundChange) dirtied() *ethcmn.Address {
	return nil
}

func (ch addLogChange) revert(s *CommitStateDB) {
	logs, err := s.GetLogs(ch.txhash)
	if err != nil {
		// panic on unmarshal error
		panic(err)
	}

	// delete logs if entry is empty or has only one item
	if len(logs) <= 1 {
		s.DeleteLogs(ch.txhash)
	} else if err := s.SetLogs(ch.txhash, logs[:len(logs)-1]); err != nil {
		// panic on marshal error
		panic(err)
	}

	s.logSize--
}

func (ch addLogChange) dirtied() *ethcmn.Address {
	return nil
}

func (ch addPreimageChange) revert(s *CommitStateDB) {
	idx, exists := s.hashToPreimageIndex[ch.hash]
	if !exists {
		// perform no-op
		return
	}

	// remove from the slice
	delete(s.hashToPreimageIndex, ch.hash)

	// if the slice contains one element, delete it
	if len(s.preimages) == 1 {
		s.preimages = []preimageEntry{}
		return
	}

	// move the elements one position left on the array
	for i := idx + 1; i < len(s.preimages); i++ {
		s.preimages[i-1] = s.preimages[i]
		// the new index is i - 1
		s.hashToPreimageIndex[s.preimages[i].hash] = i - 1
	}

	//  finally, delete the last element

	s.preimages = s.preimages[:len(s.preimages)-1]
}

func (ch addPreimageChange) dirtied() *ethcmn.Address {
	return nil
}

func (ch accessListAddAccountChange) revert(s *CommitStateDB) {
	/*
		One important invariant here, is that whenever a (addr, slot) is added, if the
		addr is not already present, the add causes two journal entries:
		- one for the address,
		- one for the (address,slot)
		Therefore, when unrolling the change, we can always blindly delete the
		(addr) at this point, since no storage adds can remain when come upon
		a single (addr) change.
	*/
	s.accessList.DeleteAddress(*ch.address)
}

func (ch accessListAddAccountChange) dirtied() *ethcmn.Address {
	return nil
}

func (ch accessListAddSlotChange) revert(s *CommitStateDB) {
	s.accessList.DeleteSlot(*ch.address, *ch.slot)
}

func (ch accessListAddSlotChange) dirtied() *ethcmn.Address {
	return nil
}

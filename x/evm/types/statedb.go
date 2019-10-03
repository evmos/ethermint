package types

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

var (
	_ ethvm.StateDB = (*CommitStateDB)(nil)

	zeroBalance = sdk.ZeroInt().BigInt()
)

type revision struct {
	id           int
	journalIndex int
}

// CommitStateDB implements the Geth state.StateDB interface. Instead of using
// a trie and database for querying and persistence, the Keeper uses KVStores
// and an account mapper is used to facilitate state transitions.
//
// TODO: This implementation is subject to change in regards to its statefull
// manner. In otherwords, how this relates to the keeper in this module.
type CommitStateDB struct {
	// TODO: We need to store the context as part of the structure itself opposed
	// to being passed as a parameter (as it should be) in order to implement the
	// StateDB interface. Perhaps there is a better way.
	ctx sdk.Context

	ak         auth.AccountKeeper
	storageKey sdk.StoreKey
	codeKey    sdk.StoreKey

	// maps that hold 'live' objects, which will get modified while processing a
	// state transition
	stateObjects      map[ethcmn.Address]*stateObject
	stateObjectsDirty map[ethcmn.Address]struct{}

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash ethcmn.Hash
	txIndex      int
	logs         map[ethcmn.Hash][]*ethtypes.Log
	logSize      uint

	// TODO: Determine if we actually need this as we do not need preimages in
	// the SDK, but it seems to be used elsewhere in Geth.
	preimages map[ethcmn.Hash][]byte

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memo-ized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionID int

	// mutex for state deep copying
	lock sync.Mutex
}

// NewCommitStateDB returns a reference to a newly initialized CommitStateDB
// which implements Geth's state.StateDB interface.
//
// CONTRACT: Stores used for state must be cache-wrapped as the ordering of the
// key/value space matters in determining the merkle root.
func NewCommitStateDB(ctx sdk.Context, ak auth.AccountKeeper, storageKey, codeKey sdk.StoreKey) *CommitStateDB {
	return &CommitStateDB{
		ctx:               ctx,
		ak:                ak,
		storageKey:        storageKey,
		codeKey:           codeKey,
		stateObjects:      make(map[ethcmn.Address]*stateObject),
		stateObjectsDirty: make(map[ethcmn.Address]struct{}),
		logs:              make(map[ethcmn.Hash][]*ethtypes.Log),
		preimages:         make(map[ethcmn.Hash][]byte),
		journal:           newJournal(),
	}
}

// WithContext returns a Database with an updated sdk context
func (csdb *CommitStateDB) WithContext(ctx sdk.Context) *CommitStateDB {
	csdb.ctx = ctx
	return csdb
}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetBalance sets the balance of an account.
func (csdb *CommitStateDB) SetBalance(addr ethcmn.Address, amount *big.Int) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetBalance(amount)
	}
}

// AddBalance adds amount to the account associated with addr.
func (csdb *CommitStateDB) AddBalance(addr ethcmn.Address, amount *big.Int) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (csdb *CommitStateDB) SubBalance(addr ethcmn.Address, amount *big.Int) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SubBalance(amount)
	}
}

// SetNonce sets the nonce (sequence number) of an account.
func (csdb *CommitStateDB) SetNonce(addr ethcmn.Address, nonce uint64) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetNonce(nonce)
	}
}

// SetState sets the storage state with a key, value pair for an account.
func (csdb *CommitStateDB) SetState(addr ethcmn.Address, key, value ethcmn.Hash) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetState(nil, key, value)
	}
}

// SetCode sets the code for a given account.
func (csdb *CommitStateDB) SetCode(addr ethcmn.Address, code []byte) {
	so := csdb.GetOrNewStateObject(addr)
	if so != nil {
		so.SetCode(ethcrypto.Keccak256Hash(code), code)
	}
}

// AddLog adds a new log to the state and sets the log metadata from the state.
func (csdb *CommitStateDB) AddLog(log *ethtypes.Log) {
	csdb.journal.append(addLogChange{txhash: csdb.thash})

	log.TxHash = csdb.thash
	log.BlockHash = csdb.bhash
	log.TxIndex = uint(csdb.txIndex)
	log.Index = csdb.logSize
	csdb.logs[csdb.thash] = append(csdb.logs[csdb.thash], log)
	csdb.logSize++
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (csdb *CommitStateDB) AddPreimage(hash ethcmn.Hash, preimage []byte) {
	if _, ok := csdb.preimages[hash]; !ok {
		csdb.journal.append(addPreimageChange{hash: hash})

		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		csdb.preimages[hash] = pi
	}
}

// AddRefund adds gas to the refund counter.
func (csdb *CommitStateDB) AddRefund(gas uint64) {
	csdb.journal.append(refundChange{prev: csdb.refund})
	csdb.refund += gas
}

// SubRefund removes gas from the refund counter. It will panic if the refund
// counter goes below zero.
func (csdb *CommitStateDB) SubRefund(gas uint64) {
	csdb.journal.append(refundChange{prev: csdb.refund})
	if gas > csdb.refund {
		panic("refund counter below zero")
	}

	csdb.refund -= gas
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// GetBalance retrieves the balance from the given address or 0 if object not
// found.
func (csdb *CommitStateDB) GetBalance(addr ethcmn.Address) *big.Int {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Balance()
	}

	return zeroBalance
}

// GetNonce returns the nonce (sequence number) for a given account.
func (csdb *CommitStateDB) GetNonce(addr ethcmn.Address) uint64 {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Nonce()
	}

	return 0
}

// TxIndex returns the current transaction index set by Prepare.
func (csdb *CommitStateDB) TxIndex() int {
	return csdb.txIndex
}

// BlockHash returns the current block hash set by Prepare.
func (csdb *CommitStateDB) BlockHash() ethcmn.Hash {
	return csdb.bhash
}

// GetCode returns the code for a given account.
func (csdb *CommitStateDB) GetCode(addr ethcmn.Address) []byte {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.Code(nil)
	}

	return nil
}

// GetCodeSize returns the code size for a given account.
func (csdb *CommitStateDB) GetCodeSize(addr ethcmn.Address) int {
	so := csdb.getStateObject(addr)
	if so == nil {
		return 0
	}

	if so.code != nil {
		return len(so.code)
	}

	// TODO: we may need to cache these lookups directly
	return len(so.Code(nil))
}

// GetCodeHash returns the code hash for a given account.
func (csdb *CommitStateDB) GetCodeHash(addr ethcmn.Address) ethcmn.Hash {
	so := csdb.getStateObject(addr)
	if so == nil {
		return ethcmn.Hash{}
	}

	return ethcmn.BytesToHash(so.CodeHash())
}

// GetState retrieves a value from the given account's storage store.
func (csdb *CommitStateDB) GetState(addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.GetState(nil, hash)
	}

	return ethcmn.Hash{}
}

// GetCommittedState retrieves a value from the given account's committed
// storage.
func (csdb *CommitStateDB) GetCommittedState(addr ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.GetCommittedState(nil, hash)
	}

	return ethcmn.Hash{}
}

// GetLogs returns the current logs for a given hash in the state.
func (csdb *CommitStateDB) GetLogs(hash ethcmn.Hash) []*ethtypes.Log {
	return csdb.logs[hash]
}

// Logs returns all the current logs in the state.
func (csdb *CommitStateDB) Logs() []*ethtypes.Log {
	var logs []*ethtypes.Log
	for _, lgs := range csdb.logs {
		logs = append(logs, lgs...)
	}

	return logs
}

// GetRefund returns the current value of the refund counter.
func (csdb *CommitStateDB) GetRefund() uint64 {
	return csdb.refund
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (csdb *CommitStateDB) Preimages() map[ethcmn.Hash][]byte {
	return csdb.preimages
}

// HasSuicided returns if the given account for the specified address has been
// killed.
func (csdb *CommitStateDB) HasSuicided(addr ethcmn.Address) bool {
	so := csdb.getStateObject(addr)
	if so != nil {
		return so.suicided
	}

	return false
}

// StorageTrie returns nil as the state in Ethermint does not use a direct
// storage trie.
func (csdb *CommitStateDB) StorageTrie(addr ethcmn.Address) ethstate.Trie {
	return nil
}

// ----------------------------------------------------------------------------
// Persistence
// ----------------------------------------------------------------------------

// Commit writes the state to the appropriate KVStores. For each state object
// in the cache, it will either be removed, or have it's code set and/or it's
// state (storage) updated. In addition, the state object (account) itself will
// be written. Finally, the root hash (version) will be returned.
func (csdb *CommitStateDB) Commit(deleteEmptyObjects bool) (root ethcmn.Hash, err error) {
	defer csdb.clearJournalAndRefund()

	// remove dirty state object entries based on the journal
	for addr := range csdb.journal.dirties {
		csdb.stateObjectsDirty[addr] = struct{}{}
	}

	// set the state objects
	for addr, so := range csdb.stateObjects {
		_, isDirty := csdb.stateObjectsDirty[addr]

		switch {
		case so.suicided || (isDirty && deleteEmptyObjects && so.empty()):
			// If the state object has been removed, don't bother syncing it and just
			// remove it from the store.
			csdb.deleteStateObject(so)

		case isDirty:
			// write any contract code associated with the state object
			if so.code != nil && so.dirtyCode {
				so.commitCode()
				so.dirtyCode = false
			}

			// update the object in the KVStore
			csdb.updateStateObject(so)
		}

		delete(csdb.stateObjectsDirty, addr)
	}

	// NOTE: Ethereum returns the trie merkle root here, but as commitment
	// actually happens in the BaseApp at EndBlocker, we do not know the root at
	// this time.
	return
}

// Finalise finalizes the state objects (accounts) state by setting their state,
// removing the csdb destructed objects and clearing the journal as well as the
// refunds.
func (csdb *CommitStateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range csdb.journal.dirties {
		so, exist := csdb.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx:
			// 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			//
			// That tx goes out of gas, and although the notion of 'touched' does not
			// exist there, the touch-event will still be recorded in the journal.
			// Since ripeMD is a special snowflake, it will persist in the journal even
			// though the journal is reverted. In this special circumstance, it may
			// exist in journal.dirties but not in stateObjects. Thus, we can safely
			// ignore it here.
			continue
		}

		if so.suicided || (deleteEmptyObjects && so.empty()) {
			csdb.deleteStateObject(so)
		} else {
			// Set all the dirty state storage items for the state object in the
			// KVStore and finally set the account in the account mapper.
			so.commitState()
			csdb.updateStateObject(so)
		}

		csdb.stateObjectsDirty[addr] = struct{}{}
	}

	// invalidate journal because reverting across transactions is not allowed
	csdb.clearJournalAndRefund()
}

// IntermediateRoot returns the current root hash of the state. It is called in
// between transactions to get the root hash that goes into transaction
// receipts.
//
// NOTE: The SDK has not concept or method of getting any intermediate merkle
// root as commitment of the merkle-ized tree doesn't happen until the
// BaseApps' EndBlocker.
func (csdb *CommitStateDB) IntermediateRoot(deleteEmptyObjects bool) ethcmn.Hash {
	csdb.Finalise(deleteEmptyObjects)

	return ethcmn.Hash{}
}

// updateStateObject writes the given state object to the store.
func (csdb *CommitStateDB) updateStateObject(so *stateObject) {
	csdb.ak.SetAccount(csdb.ctx, so.account)
}

// deleteStateObject removes the given state object from the state store.
func (csdb *CommitStateDB) deleteStateObject(so *stateObject) {
	so.deleted = true
	csdb.ak.RemoveAccount(csdb.ctx, so.account)
}

// ----------------------------------------------------------------------------
// Snapshotting
// ----------------------------------------------------------------------------

// Snapshot returns an identifier for the current revision of the state.
func (csdb *CommitStateDB) Snapshot() int {
	id := csdb.nextRevisionID
	csdb.nextRevisionID++

	csdb.validRevisions = append(
		csdb.validRevisions,
		revision{
			id:           id,
			journalIndex: csdb.journal.length(),
		},
	)

	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (csdb *CommitStateDB) RevertToSnapshot(revID int) {
	// find the snapshot in the stack of valid snapshots
	idx := sort.Search(len(csdb.validRevisions), func(i int) bool {
		return csdb.validRevisions[i].id >= revID
	})

	if idx == len(csdb.validRevisions) || csdb.validRevisions[idx].id != revID {
		panic(fmt.Errorf("revision ID %v cannot be reverted", revID))
	}

	snapshot := csdb.validRevisions[idx].journalIndex

	// replay the journal to undo changes and remove invalidated snapshots
	csdb.journal.revert(csdb, snapshot)
	csdb.validRevisions = csdb.validRevisions[:idx]
}

// ----------------------------------------------------------------------------
// Auxiliary
// ----------------------------------------------------------------------------

// Database retrieves the low level database supporting the lower level trie
// ops. It is not used in Ethermint, so it returns nil.
func (csdb *CommitStateDB) Database() ethstate.Database {
	return nil
}

// Empty returns whether the state object is either non-existent or empty
// according to the EIP161 specification (balance = nonce = code = 0).
func (csdb *CommitStateDB) Empty(addr ethcmn.Address) bool {
	so := csdb.getStateObject(addr)
	return so == nil || so.empty()
}

// Exist reports whether the given account address exists in the state. Notably,
// this also returns true for suicided accounts.
func (csdb *CommitStateDB) Exist(addr ethcmn.Address) bool {
	return csdb.getStateObject(addr) != nil
}

// Error returns the first non-nil error the StateDB encountered.
func (csdb *CommitStateDB) Error() error {
	return csdb.dbErr
}

// Suicide marks the given account as suicided and clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (csdb *CommitStateDB) Suicide(addr ethcmn.Address) bool {
	so := csdb.getStateObject(addr)
	if so == nil {
		return false
	}

	csdb.journal.append(suicideChange{
		account:     &addr,
		prev:        so.suicided,
		prevBalance: sdk.NewIntFromBigInt(so.Balance()),
	})

	so.markSuicided()
	so.SetBalance(new(big.Int))

	return true
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying account mapper and store keys to avoid reloading data for the
// next operations.
func (csdb *CommitStateDB) Reset(root ethcmn.Hash) error {
	csdb.stateObjects = make(map[ethcmn.Address]*stateObject)
	csdb.stateObjectsDirty = make(map[ethcmn.Address]struct{})
	csdb.thash = ethcmn.Hash{}
	csdb.bhash = ethcmn.Hash{}
	csdb.txIndex = 0
	csdb.logs = make(map[ethcmn.Hash][]*ethtypes.Log)
	csdb.logSize = 0
	csdb.preimages = make(map[ethcmn.Hash][]byte)

	csdb.clearJournalAndRefund()
	return nil
}

func (csdb *CommitStateDB) clearJournalAndRefund() {
	csdb.journal = newJournal()
	csdb.validRevisions = csdb.validRevisions[:0]
	csdb.refund = 0
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (csdb *CommitStateDB) Prepare(thash, bhash ethcmn.Hash, txi int) {
	csdb.thash = thash
	csdb.bhash = bhash
	csdb.txIndex = txi
}

// CreateAccount explicitly creates a state object. If a state object with the
// address already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might
// arise that a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (csdb *CommitStateDB) CreateAccount(addr ethcmn.Address) {
	newobj, prevobj := csdb.createObject(addr)
	if prevobj != nil {
		newobj.setBalance(sdk.NewIntFromBigInt(prevobj.Balance()))
	}
}

// Copy creates a deep, independent copy of the state.
//
// NOTE: Snapshots of the copied state cannot be applied to the copy.
func (csdb *CommitStateDB) Copy() ethvm.StateDB {
	csdb.lock.Lock()
	defer csdb.lock.Unlock()

	// copy all the basic fields, initialize the memory ones
	state := &CommitStateDB{
		ctx:               csdb.ctx,
		ak:                csdb.ak,
		storageKey:        csdb.storageKey,
		codeKey:           csdb.codeKey,
		stateObjects:      make(map[ethcmn.Address]*stateObject, len(csdb.journal.dirties)),
		stateObjectsDirty: make(map[ethcmn.Address]struct{}, len(csdb.journal.dirties)),
		refund:            csdb.refund,
		logs:              make(map[ethcmn.Hash][]*ethtypes.Log, len(csdb.logs)),
		logSize:           csdb.logSize,
		preimages:         make(map[ethcmn.Hash][]byte),
		journal:           newJournal(),
	}

	// copy the dirty states, logs, and preimages
	for addr := range csdb.journal.dirties {
		// There is a case where an object is in the journal but not in the
		// stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we
		// need to check for nil.
		//
		// Ref: https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527
		if object, exist := csdb.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	// Above, we don't copy the actual journal. This means that if the copy is
	// copied, the loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies.
	for addr := range csdb.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = csdb.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}

	// copy logs
	for hash, logs := range csdb.logs {
		cpy := make([]*ethtypes.Log, len(logs))
		for i, l := range logs {
			cpy[i] = new(ethtypes.Log)
			*cpy[i] = *l
		}
		state.logs[hash] = cpy
	}

	// copy pre-images
	for hash, preimage := range csdb.preimages {
		state.preimages[hash] = preimage
	}

	return state
}

// ForEachStorage iterates over each storage items, all invokes the provided
// callback on each key, value pair .
func (csdb *CommitStateDB) ForEachStorage(addr ethcmn.Address, cb func(key, value ethcmn.Hash) bool) error {
	so := csdb.getStateObject(addr)
	if so == nil {
		return nil
	}

	store := csdb.ctx.KVStore(csdb.storageKey)
	iter := sdk.KVStorePrefixIterator(store, so.Address().Bytes())

	for ; iter.Valid(); iter.Next() {
		key := ethcmn.BytesToHash(iter.Key())
		value := iter.Value()

		if value, dirty := so.dirtyStorage[key]; dirty {
			cb(key, value)
			continue
		}

		cb(key, ethcmn.BytesToHash(value))
	}

	iter.Close()
	return nil
}

// GetOrNewStateObject retrieves a state object or create a new state object if
// nil.
func (csdb *CommitStateDB) GetOrNewStateObject(addr ethcmn.Address) StateObject {
	so := csdb.getStateObject(addr)
	if so == nil || so.deleted {
		so, _ = csdb.createObject(addr)
	}

	return so
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (csdb *CommitStateDB) createObject(addr ethcmn.Address) (newObj, prevObj *stateObject) {
	prevObj = csdb.getStateObject(addr)

	acc := csdb.ak.NewAccountWithAddress(csdb.ctx, sdk.AccAddress(addr.Bytes()))
	newObj = newObject(csdb, acc)
	newObj.setNonce(0) // sets the object to dirty

	if prevObj == nil {
		csdb.journal.append(createObjectChange{account: &addr})
	} else {
		csdb.journal.append(resetObjectChange{prev: prevObj})
	}

	csdb.setStateObject(newObj)
	return newObj, prevObj
}

// setError remembers the first non-nil error it is called with.
func (csdb *CommitStateDB) setError(err error) {
	if csdb.dbErr == nil {
		csdb.dbErr = err
	}
}

// getStateObject attempts to retrieve a state object given by the address.
// Returns nil and sets an error if not found.
func (csdb *CommitStateDB) getStateObject(addr ethcmn.Address) (stateObject *stateObject) {
	// prefer 'live' (cached) objects
	if so := csdb.stateObjects[addr]; so != nil {
		if so.deleted {
			return nil
		}

		return so
	}

	// otherwise, attempt to fetch the account from the account mapper
	acc := csdb.ak.GetAccount(csdb.ctx, addr.Bytes())
	if acc == nil {
		csdb.setError(fmt.Errorf("no account found for address: %X", addr.Bytes()))
		return nil
	}

	// insert the state object into the live set
	so := newObject(csdb, acc)
	csdb.setStateObject(so)

	return so
}

func (csdb *CommitStateDB) setStateObject(so *stateObject) {
	csdb.stateObjects[so.Address()] = so
}

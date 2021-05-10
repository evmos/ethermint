package types

import (
	"fmt"
	"math/big"
	"sort"
	"sync"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	ethermint "github.com/cosmos/ethermint/types"

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
// and an AccountKeeper to facilitate state transitions.
//
// TODO: This implementation is subject to change in regards to its statefull
// manner. In otherwords, how this relates to the keeper in this module.
type CommitStateDB struct {
	// TODO: We need to store the context as part of the structure itself opposed
	// to being passed as a parameter (as it should be) in order to implement the
	// StateDB interface. Perhaps there is a better way.
	ctx sdk.Context

	storeKey      sdk.StoreKey
	paramSpace    paramtypes.Subspace
	accountKeeper AccountKeeper
	bankKeeper    BankKeeper

	// array that hold 'live' objects, which will get modified while processing a
	// state transition
	stateObjects         []stateEntry
	addressToObjectIndex map[ethcmn.Address]int // map from address to the index of the state objects slice
	stateObjectsDirty    map[ethcmn.Address]struct{}

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash ethcmn.Hash
	txIndex      int
	logSize      uint

	// TODO: Determine if we actually need this as we do not need preimages in
	// the SDK, but it seems to be used elsewhere in Geth.
	preimages           []preimageEntry
	hashToPreimageIndex map[ethcmn.Hash]int // map from hash to the index of the preimages slice

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

	// Per-transaction access list
	accessList *accessList

	// mutex for state deep copying
	lock sync.Mutex
}

// NewCommitStateDB returns a reference to a newly initialized CommitStateDB
// which implements Geth's state.StateDB interface.
//
// CONTRACT: Stores used for state must be cache-wrapped as the ordering of the
// key/value space matters in determining the merkle root.
func NewCommitStateDB(
	ctx sdk.Context, storeKey sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak AccountKeeper, bankKeeper BankKeeper,
) *CommitStateDB {
	return &CommitStateDB{
		ctx:                  ctx,
		storeKey:             storeKey,
		paramSpace:           paramSpace,
		accountKeeper:        ak,
		bankKeeper:           bankKeeper,
		stateObjects:         []stateEntry{},
		addressToObjectIndex: make(map[ethcmn.Address]int),
		stateObjectsDirty:    make(map[ethcmn.Address]struct{}),
		preimages:            []preimageEntry{},
		hashToPreimageIndex:  make(map[ethcmn.Hash]int),
		journal:              newJournal(),
		accessList:           newAccessList(),
	}
}

// WithContext returns a Database with an updated SDK context
func (csdb *CommitStateDB) WithContext(ctx sdk.Context) *CommitStateDB {
	csdb.ctx = ctx
	return csdb
}

// ----------------------------------------------------------------------------
// Setters
// ----------------------------------------------------------------------------

// SetHeightHash sets the block header hash associated with a given height.
func (csdb *CommitStateDB) SetHeightHash(height uint64, hash ethcmn.Hash) {
	store := prefix.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixBlockHeightHash)
	key := KeyBlockHeightHash(height)
	store.Set(key, hash.Bytes())
}

// SetParams sets the evm parameters to the param space.
func (csdb *CommitStateDB) SetParams(params Params) {
	csdb.paramSpace.SetParamSet(csdb.ctx, &params)
}

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

// ----------------------------------------------------------------------------
// Transaction logs
// Required for upgrade logic or ease of querying.
// NOTE: we use BinaryLengthPrefixed since the tx logs are also included on Result data,
// which can't use BinaryBare.
// ----------------------------------------------------------------------------

// SetLogs sets the logs for a transaction in the KVStore.
func (csdb *CommitStateDB) SetLogs(hash ethcmn.Hash, logs []*ethtypes.Log) error {
	store := prefix.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixLogs)

	txLogs := NewTransactionLogsFromEth(hash, logs)
	bz, err := ModuleCdc.MarshalBinaryBare(&txLogs)
	if err != nil {
		return err
	}

	store.Set(hash.Bytes(), bz)
	csdb.logSize = uint(len(logs))
	return nil
}

// DeleteLogs removes the logs from the KVStore. It is used during journal.Revert.
func (csdb *CommitStateDB) DeleteLogs(hash ethcmn.Hash) {
	store := prefix.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixLogs)
	store.Delete(hash.Bytes())
}

// AddLog adds a new log to the state and sets the log metadata from the state.
func (csdb *CommitStateDB) AddLog(log *ethtypes.Log) {
	csdb.journal.append(addLogChange{txhash: csdb.thash})

	log.TxHash = csdb.thash
	log.BlockHash = csdb.bhash
	log.TxIndex = uint(csdb.txIndex)
	log.Index = csdb.logSize

	logs, err := csdb.GetLogs(csdb.thash)
	if err != nil {
		// panic on unmarshal error
		panic(err)
	}

	if err = csdb.SetLogs(csdb.thash, append(logs, log)); err != nil {
		// panic on marshal error
		panic(err)
	}
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (csdb *CommitStateDB) AddPreimage(hash ethcmn.Hash, preimage []byte) {
	if _, ok := csdb.hashToPreimageIndex[hash]; !ok {
		csdb.journal.append(addPreimageChange{hash: hash})

		pi := make([]byte, len(preimage))
		copy(pi, preimage)

		csdb.preimages = append(csdb.preimages, preimageEntry{hash: hash, preimage: pi})
		csdb.hashToPreimageIndex[hash] = len(csdb.preimages) - 1
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

// AddAddressToAccessList adds the given address to the access list
func (csdb *CommitStateDB) AddAddressToAccessList(addr ethcmn.Address) {
	if csdb.accessList.AddAddress(addr) {
		csdb.journal.append(accessListAddAccountChange{&addr})
	}
}

// AddSlotToAccessList adds the given (address, slot)-tuple to the access list
func (csdb *CommitStateDB) AddSlotToAccessList(addr ethcmn.Address, slot ethcmn.Hash) {
	addrMod, slotMod := csdb.accessList.AddSlot(addr, slot)
	if addrMod {
		// In practice, this should not happen, since there is no way to enter the
		// scope of 'address' without having the 'address' become already added
		// to the access list (via call-variant, create, etc).
		// Better safe than sorry, though
		csdb.journal.append(accessListAddAccountChange{&addr})
	}
	if slotMod {
		csdb.journal.append(accessListAddSlotChange{
			address: &addr,
			slot:    &slot,
		})
	}
}

// AddressInAccessList returns true if the given address is in the access list.
func (csdb *CommitStateDB) AddressInAccessList(addr ethcmn.Address) bool {
	return csdb.accessList.ContainsAddress(addr)
}

// SlotInAccessList returns true if the given (address, slot)-tuple is in the access list.
func (csdb *CommitStateDB) SlotInAccessList(addr ethcmn.Address, slot ethcmn.Hash) (bool, bool) {
	return csdb.accessList.Contains(addr, slot)
}

// PrepareAccessList handles the preparatory steps for executing a state transition with
// regards to both EIP-2929 and EIP-2930:
//
// - Add sender to access list (2929)
// - Add destination to access list (2929)
// - Add precompiles to access list (2929)
// - Add the contents of the optional tx access list (2930)
//
// This method should only be called if Yolov3/Berlin/2929+2930 is applicable at the current number.
func (csdb *CommitStateDB) PrepareAccessList(sender ethcmn.Address, dst *ethcmn.Address, precompiles []ethcmn.Address, list ethtypes.AccessList) {
	csdb.AddAddressToAccessList(sender)
	if dst != nil {
		csdb.AddAddressToAccessList(*dst)
		// If it's a create-tx, the destination will be added inside evm.create
	}
	for _, addr := range precompiles {
		csdb.AddAddressToAccessList(addr)
	}
	for _, el := range list {
		csdb.AddAddressToAccessList(el.Address)
		for _, key := range el.StorageKeys {
			csdb.AddSlotToAccessList(el.Address, key)
		}
	}
}

// ----------------------------------------------------------------------------
// Getters
// ----------------------------------------------------------------------------

// GetHeightHash returns the block header hash associated with a given block height and chain epoch number.
func (csdb *CommitStateDB) GetHeightHash(height uint64) ethcmn.Hash {
	store := prefix.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixBlockHeightHash)
	key := KeyBlockHeightHash(height)
	bz := store.Get(key)
	if len(bz) == 0 {
		return ethcmn.Hash{}
	}

	return ethcmn.BytesToHash(bz)
}

// GetParams returns the total set of evm parameters.
func (csdb *CommitStateDB) GetParams() (params Params) {
	csdb.paramSpace.GetParamSet(csdb.ctx, &params)
	return params
}

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

// GetLogs returns the current logs for a given transaction hash from the KVStore.
func (csdb *CommitStateDB) GetLogs(hash ethcmn.Hash) ([]*ethtypes.Log, error) {
	store := prefix.NewStore(csdb.ctx.KVStore(csdb.storeKey), KeyPrefixLogs)
	bz := store.Get(hash.Bytes())
	if len(bz) == 0 {
		// return nil error if logs are not found
		return []*ethtypes.Log{}, nil
	}

	var txLogs TransactionLogs
	if err := ModuleCdc.UnmarshalBinaryBare(bz, &txLogs); err != nil {
		return []*ethtypes.Log{}, err
	}

	allLogs := []*ethtypes.Log{}
	for _, txLog := range txLogs.Logs {
		allLogs = append(allLogs, txLog.ToEthereum())
	}

	return allLogs, nil
}

// AllLogs returns all the current logs in the state.
func (csdb *CommitStateDB) AllLogs() []*ethtypes.Log {
	store := csdb.ctx.KVStore(csdb.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, KeyPrefixLogs)
	defer iterator.Close()

	allLogs := []*ethtypes.Log{}
	for ; iterator.Valid(); iterator.Next() {
		var txLogs TransactionLogs
		ModuleCdc.MustUnmarshalBinaryBare(iterator.Value(), &txLogs)

		for _, txLog := range txLogs.Logs {
			allLogs = append(allLogs, txLog.ToEthereum())
		}
	}

	return allLogs
}

// GetRefund returns the current value of the refund counter.
func (csdb *CommitStateDB) GetRefund() uint64 {
	return csdb.refund
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (csdb *CommitStateDB) Preimages() map[ethcmn.Hash][]byte {
	preimages := map[ethcmn.Hash][]byte{}

	for _, pe := range csdb.preimages {
		preimages[pe.hash] = pe.preimage
	}
	return preimages
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
func (csdb *CommitStateDB) Commit(deleteEmptyObjects bool) (ethcmn.Hash, error) {
	defer csdb.clearJournalAndRefund()

	// remove dirty state object entries based on the journal
	for _, dirty := range csdb.journal.dirties {
		csdb.stateObjectsDirty[dirty.address] = struct{}{}
	}

	// set the state objects
	for _, stateEntry := range csdb.stateObjects {
		_, isDirty := csdb.stateObjectsDirty[stateEntry.address]

		switch {
		case stateEntry.stateObject.suicided || (isDirty && deleteEmptyObjects && stateEntry.stateObject.empty()):
			// If the state object has been removed, don't bother syncing it and just
			// remove it from the store.
			csdb.deleteStateObject(stateEntry.stateObject)

		case isDirty:
			// write any contract code associated with the state object
			if stateEntry.stateObject.code != nil && stateEntry.stateObject.dirtyCode {
				stateEntry.stateObject.commitCode()
				stateEntry.stateObject.dirtyCode = false
			}

			// update the object in the KVStore
			if err := csdb.updateStateObject(stateEntry.stateObject); err != nil {
				return ethcmn.Hash{}, err
			}
		}

		delete(csdb.stateObjectsDirty, stateEntry.address)
	}

	// NOTE: Ethereum returns the trie merkle root here, but as commitment
	// actually happens in the BaseApp at EndBlocker, we do not know the root at
	// this time.
	return ethcmn.Hash{}, nil
}

// Finalise finalizes the state objects (accounts) state by setting their state,
// removing the csdb destructed objects and clearing the journal as well as the
// refunds.
func (csdb *CommitStateDB) Finalise(deleteEmptyObjects bool) error {
	for _, dirty := range csdb.journal.dirties {
		idx, exist := csdb.addressToObjectIndex[dirty.address]
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

		stateEntry := csdb.stateObjects[idx]
		if stateEntry.stateObject.suicided || (deleteEmptyObjects && stateEntry.stateObject.empty()) {
			csdb.deleteStateObject(stateEntry.stateObject)
		} else {
			// Set all the dirty state storage items for the state object in the
			// KVStore and finally set the account in the account mapper.
			stateEntry.stateObject.commitState()
			if err := csdb.updateStateObject(stateEntry.stateObject); err != nil {
				return err
			}
		}

		csdb.stateObjectsDirty[dirty.address] = struct{}{}
	}

	// invalidate journal because reverting across transactions is not allowed
	csdb.clearJournalAndRefund()
	return nil
}

// IntermediateRoot returns the current root hash of the state. It is called in
// between transactions to get the root hash that goes into transaction
// receipts.
//
// NOTE: The SDK has not concept or method of getting any intermediate merkle
// root as commitment of the merkle-ized tree doesn't happen until the
// BaseApps' EndBlocker.
func (csdb *CommitStateDB) IntermediateRoot(deleteEmptyObjects bool) (ethcmn.Hash, error) {
	if err := csdb.Finalise(deleteEmptyObjects); err != nil {
		return ethcmn.Hash{}, err
	}

	return ethcmn.Hash{}, nil
}

// updateStateObject writes the given state object to the store.
func (csdb *CommitStateDB) updateStateObject(so *stateObject) error {
	evmDenom := csdb.GetParams().EvmDenom
	// NOTE: we don't use sdk.NewCoin here to avoid panic on test importer's genesis
	newBalance := sdk.Coin{Denom: evmDenom, Amount: sdk.NewIntFromBigInt(so.Balance())}
	if !newBalance.IsValid() {
		return fmt.Errorf("invalid balance %s", newBalance)
	}

	err := csdb.bankKeeper.SetBalance(csdb.ctx, so.account.GetAddress(), newBalance)
	if err != nil {
		return err
	}
	csdb.accountKeeper.SetAccount(csdb.ctx, so.account)
	return nil
}

// deleteStateObject removes the given state object from the state store.
func (csdb *CommitStateDB) deleteStateObject(so *stateObject) {
	so.deleted = true
	csdb.accountKeeper.RemoveAccount(csdb.ctx, so.account)
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
func (csdb *CommitStateDB) Reset(_ ethcmn.Hash) error {
	csdb.stateObjects = []stateEntry{}
	csdb.addressToObjectIndex = make(map[ethcmn.Address]int)
	csdb.stateObjectsDirty = make(map[ethcmn.Address]struct{})
	csdb.thash = ethcmn.Hash{}
	csdb.bhash = ethcmn.Hash{}
	csdb.txIndex = 0
	csdb.logSize = 0
	csdb.preimages = []preimageEntry{}
	csdb.hashToPreimageIndex = make(map[ethcmn.Hash]int)
	csdb.accessList = newAccessList()

	csdb.clearJournalAndRefund()
	return nil
}

// UpdateAccounts updates the nonce and coin balances of accounts
func (csdb *CommitStateDB) UpdateAccounts() {
	for _, stateEntry := range csdb.stateObjects {
		address := sdk.AccAddress(stateEntry.address.Bytes())
		currAccount := csdb.accountKeeper.GetAccount(csdb.ctx, address)
		ethAcc, ok := currAccount.(*ethermint.EthAccount)
		if !ok {
			continue
		}

		evmDenom := csdb.GetParams().EvmDenom
		balance := csdb.bankKeeper.GetBalance(csdb.ctx, address, evmDenom)

		if stateEntry.stateObject.Balance() != balance.Amount.BigInt() && balance.IsValid() {
			stateEntry.stateObject.balance = balance.Amount
		}

		if stateEntry.stateObject.Nonce() != ethAcc.GetSequence() {
			stateEntry.stateObject.account = ethAcc
		}
	}
}

// ClearStateObjects clears cache of state objects to handle account changes outside of the EVM
func (csdb *CommitStateDB) ClearStateObjects() {
	csdb.stateObjects = []stateEntry{}
	csdb.addressToObjectIndex = make(map[ethcmn.Address]int)
	csdb.stateObjectsDirty = make(map[ethcmn.Address]struct{})
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
func (csdb *CommitStateDB) Copy() *CommitStateDB {

	// copy all the basic fields, initialize the memory ones
	state := &CommitStateDB{}
	CopyCommitStateDB(csdb, state)

	return state
}

func CopyCommitStateDB(from, to *CommitStateDB) {
	from.lock.Lock()
	defer from.lock.Unlock()

	to.ctx = from.ctx
	to.storeKey = from.storeKey
	to.paramSpace = from.paramSpace
	to.accountKeeper = from.accountKeeper
	to.bankKeeper = from.bankKeeper
	to.stateObjects = []stateEntry{}
	to.addressToObjectIndex = make(map[ethcmn.Address]int)
	to.stateObjectsDirty = make(map[ethcmn.Address]struct{})
	to.refund = from.refund
	to.logSize = from.logSize
	to.preimages = make([]preimageEntry, len(from.preimages))
	to.hashToPreimageIndex = make(map[ethcmn.Hash]int, len(from.hashToPreimageIndex))
	to.journal = newJournal()
	to.thash = from.thash
	to.bhash = from.bhash
	to.txIndex = from.txIndex
	validRevisions := make([]revision, len(from.validRevisions))
	copy(validRevisions, from.validRevisions)
	to.validRevisions = validRevisions
	to.nextRevisionID = from.nextRevisionID
	to.accessList = from.accessList.Copy()

	// copy the dirty states, logs, and preimages
	for _, dirty := range from.journal.dirties {
		// There is a case where an object is in the journal but not in the
		// stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we
		// need to check for nil.
		//
		// Ref: https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527
		if idx, exist := from.addressToObjectIndex[dirty.address]; exist {
			to.stateObjects = append(to.stateObjects, stateEntry{
				address:     dirty.address,
				stateObject: from.stateObjects[idx].stateObject.deepCopy(to),
			})
			to.addressToObjectIndex[dirty.address] = len(to.stateObjects) - 1
			to.stateObjectsDirty[dirty.address] = struct{}{}
		}
	}

	// Above, we don't copy the actual journal. This means that if the copy is
	// copied, the loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies.
	for addr := range from.stateObjectsDirty {
		if idx, exist := to.addressToObjectIndex[addr]; !exist {
			to.setStateObject(from.stateObjects[idx].stateObject.deepCopy(to))
			to.stateObjectsDirty[addr] = struct{}{}
		}
	}

	// copy pre-images
	for i, preimageEntry := range from.preimages {
		to.preimages[i] = preimageEntry
		to.hashToPreimageIndex[preimageEntry.hash] = i
	}
}

// ForEachStorage iterates over each storage items, all invoke the provided
// callback on each key, value pair.
func (csdb *CommitStateDB) ForEachStorage(addr ethcmn.Address, cb func(key, value ethcmn.Hash) (stop bool)) error {
	so := csdb.getStateObject(addr)
	if so == nil {
		return nil
	}

	store := csdb.ctx.KVStore(csdb.storeKey)
	prefix := AddressStoragePrefix(so.Address())
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := ethcmn.BytesToHash(iterator.Key())
		value := ethcmn.BytesToHash(iterator.Value())

		if idx, dirty := so.keyToDirtyStorageIndex[key]; dirty {
			// check if iteration stops
			if cb(key, ethcmn.HexToHash(so.dirtyStorage[idx].Value)) {
				break
			}

			continue
		}

		// check if iteration stops
		if cb(key, value) {
			return nil
		}
	}

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

	acc := csdb.accountKeeper.NewAccountWithAddress(csdb.ctx, sdk.AccAddress(addr.Bytes()))

	newObj = newStateObject(csdb, acc, sdk.ZeroInt())
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
	if idx, found := csdb.addressToObjectIndex[addr]; found {
		// prefer 'live' (cached) objects
		if so := csdb.stateObjects[idx].stateObject; so != nil {
			if so.deleted {
				return nil
			}

			return so
		}
	}

	// otherwise, attempt to fetch the account from the account mapper
	acc := csdb.accountKeeper.GetAccount(csdb.ctx, sdk.AccAddress(addr.Bytes()))
	if acc == nil {
		csdb.setError(fmt.Errorf("no account found for address: %s", addr.String()))
		return nil
	}

	evmDenom := csdb.GetParams().EvmDenom
	balance := csdb.bankKeeper.GetBalance(csdb.ctx, acc.GetAddress(), evmDenom)

	// insert the state object into the live set
	so := newStateObject(csdb, acc, balance.Amount)
	csdb.setStateObject(so)

	return so
}

func (csdb *CommitStateDB) setStateObject(so *stateObject) {
	if idx, found := csdb.addressToObjectIndex[so.Address()]; found {
		// update the existing object
		csdb.stateObjects[idx].stateObject = so
		return
	}

	// append the new state object to the stateObjects slice
	se := stateEntry{
		address:     so.Address(),
		stateObject: so,
	}

	csdb.stateObjects = append(csdb.stateObjects, se)
	csdb.addressToObjectIndex[se.address] = len(csdb.stateObjects) - 1
}

// RawDump returns a raw state dump.
//
// TODO: Implement if we need it, especially for the RPC API.
func (csdb *CommitStateDB) RawDump() ethstate.Dump {
	return ethstate.Dump{}
}

type preimageEntry struct {
	// hash key of the preimage entry
	hash     ethcmn.Hash
	preimage []byte
}

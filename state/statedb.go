package state

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

var _ ethstate.StateDB = (*CommitStateDB)(nil)

type CommitStateDB struct {
	// TODO: Figure out a way to not need to store a context as part of the
	// structure
	ctx sdk.Context

	am         auth.AccountMapper
	storageKey sdk.StoreKey

	// maps that hold 'live' objects, which will get modified while processing a
	// state transition
	stateObjects      map[ethcmn.Address]*stateObject
	stateObjectsDirty map[ethcmn.Address]struct{}

	thash, bhash ethcmn.Hash
	txIndex      int
	logs         map[ethcmn.Hash][]*ethtypes.Log
	logSize      uint

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []ethstate.Revision
	nextRevisionID int
}

func NewCommitStateDB(ctx sdk.Context) (*CommitStateDB, error) {
	// tr, err := db.OpenTrie(root)
	// if err != nil {
	// 	return nil, err
	// }

	return &CommitStateDB{
		// stateObjects:      make(map[ethcmn.Address]*stateObject),
		// stateObjectsDirty: make(map[ethcmn.Address]struct{}),
		// logs:              make(map[ethcmn.Hash][]*types.Log),
		// preimages:         make(map[ethcmn.Hash][]byte),
		journal: newJournal(),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (csdb *CommitStateDB) setError(err error) {
	if csdb.dbErr == nil {
		csdb.dbErr = err
	}
}

// Error returns the first non-nil error the StateDB encountered.
func (csdb *CommitStateDB) Error() error {
	return csdb.dbErr
}

// Retrieve the balance from the given address or 0 if object not found
func (csdb *CommitStateDB) GetBalance(addr ethcmn.Address) *big.Int {
	stateObject := csdb.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}

	return common.Big0
}

// Retrieve a state object given by the address. Returns nil if not found.
func (csdb *CommitStateDB) getStateObject(addr ethcmn.Address) (stateObject *stateObject) {
	// prefer 'live' (cached) objects
	if obj := csdb.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}

		return obj
	}

	acc := csdb.am.GetAccount(csdb.ctx, addr.Bytes())
	if acc == nil {
		csdb.setError(fmt.Errorf("no account found for address: %X", addr.Bytes()))
	}

	// insert the state object into the live set
	obj := newObject(csdb, acc)
	csdb.setStateObject(obj)
	return obj
}

func (csdb *CommitStateDB) setStateObject(object *stateObject) {
	csdb.stateObjects[object.Address()] = object
}

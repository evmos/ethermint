package state

import (
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/core"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtrie "github.com/ethereum/go-ethereum/trie"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	// AccountsKey is the key used for storing Ethereum accounts in the Cosmos
	// SDK multi-store.
	AccountsKey = types.NewKVStoreKey("account")

	// StorageKey is the key used for storing Ethereum contract storage in the
	// Cosmos SDK multi-store.
	StorageKey = types.NewKVStoreKey("storage")

	// CodeKey is the key used for storing Ethereum contract code in the Cosmos
	// SDK multi-store.
	CodeKey = types.NewKVStoreKey("code")
)

// Database implements the Ethereum state.Database interface.
type Database struct {
	// stateStore will be used for the history of accounts (balance, nonce,
	// storage root hash, code hash) and for the history of contract data
	// (effects of SSTORE instruction).
	stateStore    store.CommitMultiStore
	accountsCache store.CacheKVStore
	storageCache  store.CacheKVStore

	// codeDB contains mappings of codeHash => code
	//
	// NOTE: This database will store the information in memory until is it
	// committed, using the function Commit. This function is called outside of
	// the ApplyTransaction function, therefore in Ethermint we need to make
	// sure this commit is invoked somewhere after each block or whatever the
	// appropriate time for it.
	codeDB    dbm.DB
	ethTrieDB *ethtrie.Database

	// TODO: Do we need this/document?
	Tracing bool
}

// NewDatabase returns a reference to an initialized Database type which
// implements Ethereum's state.Database interface. An error is returned if the
// latest state failed to load. The underlying storage structure is defined by
// the Cosmos SDK IAVL tree.
func NewDatabase(stateDB, codeDB dbm.DB) (*Database, error) {
	// Initialize an implementation of Ethereum state.Database and create a
	// Cosmos SDK multi-store.
	db := &Database{
		stateStore: store.NewCommitMultiStore(stateDB),
	}

	// Create the underlying multi-store stores that will persist account and
	// account storage data.
	db.stateStore.MountStoreWithDB(AccountsKey, types.StoreTypeIAVL, nil)
	db.stateStore.MountStoreWithDB(StorageKey, types.StoreTypeIAVL, nil)

	// Load the latest account state from the Cosmos SDK multi-store.
	if err := db.stateStore.LoadLatestVersion(); err != nil {
		return nil, err
	}

	// Set the persistent Cosmos SDK Database and initialize an Ethereum
	// trie.Database using an EthereumDB as the underlying implementation of
	// the ethdb.Database interface. It will be used to facilitate persistence
	// of contract byte code when committing state.
	db.codeDB = codeDB
	db.ethTrieDB = ethtrie.NewDatabase(&core.EthereumDB{CodeDB: codeDB})

	return db, nil
}

// OpenTrie implements Ethereum's state.Database interface. It returns a Trie
// type which implements the Ethereum state.Trie interface. It us used for
// storage of accounts. An error is returned if state cannot load for a
// given version. The account cache is reset if the state is successfully
// loaded and the version is not the latest.
//
// CONTRACT: The root parameter is not interpreted as a state root hash, but as
// an encoding of an Cosmos SDK IAVL tree version.
func (db *Database) OpenTrie(root ethcommon.Hash) (ethstate.Trie, error) {
	version := db.stateStore.LastCommitID().Version

	if !isRootEmpty(root) {
		version = versionFromRootHash(root)

		if db.stateStore.LastCommitID().Version != version {
			if err := db.stateStore.LoadVersion(version); err != nil {
				return nil, err
			}

			db.accountsCache = nil
		}
	}

	if db.accountsCache == nil {
		db.accountsCache = store.NewCacheKVStore(db.stateStore.GetCommitKVStore(AccountsKey))
		db.storageCache = store.NewCacheKVStore(db.stateStore.GetCommitKVStore(StorageKey))
	}

	return &Trie{
		store:         db.accountsCache,
		accountsCache: db.accountsCache,
		storageCache:  db.storageCache,
		ethTrieDB:     db.ethTrieDB,
		empty:         isRootEmpty(root),
		root:          rootHashFromVersion(db.stateStore.LastCommitID().Version),
	}, nil
}

// OpenStorageTrie implements Ethereum's state.Database interface. It returns
// a Trie type which implements the Ethereum state.Trie interface. It is used
// for storage of contract storage (state). Also, this trie is never committed
// separately as all the data is in a single multi-store and is committed when
// the account IAVL tree is committed.
//
// NOTE: It is assumed that the account state has already been loaded via
// OpenTrie.
//
// CONTRACT: The root parameter is not interpreted as a state root hash, but as
// an encoding of an IAVL tree version.
func (db *Database) OpenStorageTrie(addrHash, root ethcommon.Hash) (ethstate.Trie, error) {
	// a contract storage trie does not need an accountCache, storageCache or
	// an Ethereum trie because it will not be used upon commitment.
	return &Trie{
		store:  db.storageCache,
		prefix: addrHash.Bytes(),
		empty:  isRootEmpty(root),
		root:   rootHashFromVersion(db.stateStore.LastCommitID().Version),
	}, nil
}

// CopyTrie implements Ethereum's state.Database interface. For now, it
// performs a no-op as the underlying Cosmos SDK IAVL tree does not support
// such an operation.
//
// TODO: Does the IAVL tree need to support this operation? If so, why and
// how?
func (db *Database) CopyTrie(ethstate.Trie) ethstate.Trie {
	return nil
}

// ContractCode implements Ethereum's state.Database interface. It will return
// the contract byte code for a given code hash. It will not return an error.
func (db *Database) ContractCode(addrHash, codeHash ethcommon.Hash) ([]byte, error) {
	return db.codeDB.Get(codeHash[:]), nil
}

// ContractCodeSize implements Ethereum's state.Database interface. It will
// return the contract byte code size for a given code hash. It will not return
// an error.
func (db *Database) ContractCodeSize(addrHash, codeHash ethcommon.Hash) (int, error) {
	return len(db.codeDB.Get(codeHash[:])), nil
}

// Commit commits the underlying Cosmos SDK multi-store returning the commit
// ID.
func (db *Database) Commit() types.CommitID {
	return db.stateStore.Commit()
}

// TrieDB implements Ethereum's state.Database interface. It returns Ethereum's
// trie.Database low level trie database used for contract state storage. In
// the context of Ethermint, it'll be used to solely store mappings of
// codeHash => code.
func (db *Database) TrieDB() *ethtrie.Database {
	return db.ethTrieDB
}

// isRootEmpty returns true if a given root hash is empty or false otherwise.
func isRootEmpty(root ethcommon.Hash) bool {
	return root == ethcommon.Hash{}
}

package state

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	ethtrie "github.com/ethereum/go-ethereum/trie"

	lru "github.com/hashicorp/golang-lru"
)

const (
	versionLen = 8
)

// Trie implements the Ethereum state.Trie interface.
type Trie struct {
	// accountsCache contains all the accounts in memory to persit when
	// committing the trie. A CacheKVStore is used to provide deterministic
	// ordering.
	accountsCache store.CacheKVStore
	// storageCache contains all the contract storage in memory to persit when
	// committing the trie. A CacheKVStore is used to provide deterministic
	// ordering.
	storageCache store.CacheKVStore

	storeCache *lru.Cache

	// Store is an IAVL KV store that is part of a larger store except it used
	// for a specific prefix. It will either be an accountsCache or a
	// storageCache.
	store store.KVStore

	// prefix is a static prefix used for persistence operations where the
	// storage data is a contract state. This is to prevent key collisions
	// since the IAVL tree is used for all contract state.
	prefix []byte

	// empty reflects if there exists any data in the tree
	empty bool

	// root is the encoding of an IAVL tree root (version)
	root ethcmn.Hash

	ethTrieDB *ethtrie.Database
}

// prefixKey returns a composite key composed of a static prefix and a given
// key. This is used in situations where the storage data is contract state and
// the underlying structure to store said state is a single IAVL tree. To
// prevent collision, a static prefix is used.
func (t *Trie) prefixKey(key []byte) []byte {
	compositeKey := make([]byte, len(t.prefix)+len(key))

	copy(compositeKey, t.prefix)
	copy(compositeKey[len(t.prefix):], key)

	return compositeKey
}

// TryGet implements the Ethereum state.Trie interface. It returns the value
// for key stored in the trie. The value bytes must not be modified by the
// caller.
func (t *Trie) TryGet(key []byte) ([]byte, error) {
	if t.IsStorageTrie() {
		key = t.prefixKey(key)
	}
	keyStr := string(key)
	if cached, ok := t.storeCache.Get(keyStr); ok {
		return cached.([]byte), nil
	}
	value := t.store.Get(key)
	t.storeCache.Add(keyStr, value)
	return value, nil
}

// TryUpdate implements the Ethereum state.Trie interface. It associates a
// given key with a value in the trie. Subsequent calls to Get will return a
// value. It also marks the tree as not empty.
//
// CONTRACT: The order of insertions must be deterministic due to the nature of
// the IAVL tree. Since a CacheKVStore is used as the storage type, the keys
// will be sorted giving us a deterministic ordering.
func (t *Trie) TryUpdate(key, value []byte) error {
	t.empty = false

	if t.IsStorageTrie() {
		key = t.prefixKey(key)
	}

	t.store.Set(key, value)
	t.storeCache.Add(string(key), value)
	return nil
}

// TryDelete implements the Ethereum state.Trie interface. It removes any
// existing value for a given key from the trie.
//
// CONTRACT: The order of deletions must be deterministic due to the nature of
// the IAVL tree. Since a CacheKVStore is used as the storage type, the keys
// will be sorted giving us a deterministic ordering.
func (t *Trie) TryDelete(key []byte) error {
	if t.IsStorageTrie() {
		key = t.prefixKey(key)
	}

	t.store.Delete(key)
	t.storeCache.Remove(string(key))
	return nil
}

// Commit implements the Ethereum state.Trie interface. It persists transient
// state. State is held by a merkelized multi-store IAVL tree. Commitment will
// only occur through an account trie, in other words, when the prefix of the
// trie is nil. In such a case, if either the accountCache or the storageCache
// are not nil, they are persisted. In addition, all the mappings of
// codeHash => code are also persisted. All these operations are performed in a
// deterministic order. Transient state is built up in a CacheKVStore. Finally,
// a root hash is returned or an error if any operation fails.
//
// CONTRACT: The root is an encoded IAVL tree version and each new commitment
// increments the version by one.
func (t *Trie) Commit(_ ethtrie.LeafCallback) (ethcmn.Hash, error) {
	if t.empty {
		return ethcmn.Hash{}, nil
	}

	newRoot := rootHashFromVersion(versionFromRootHash(t.root) + 1)

	if !t.IsStorageTrie() {
		if t.accountsCache != nil {
			t.accountsCache.Write()
			t.accountsCache = nil
		}

		if t.storageCache != nil {
			t.storageCache.Write()
			t.storageCache = nil
		}

		// persist the mappings of codeHash => code
		for _, n := range t.ethTrieDB.Nodes() {
			if err := t.ethTrieDB.Commit(n, false); err != nil {
				return ethcmn.Hash{}, err
			}
		}
	}

	t.root = newRoot
	return newRoot, nil
}

// Hash implements the Ethereum state.Trie interface. It returns the state root
// of the Trie which is an encoding of the underlying IAVL tree.
//
// CONTRACT: The root is an encoded IAVL tree version.
func (t *Trie) Hash() ethcmn.Hash {
	return t.root
}

// NodeIterator implements the Ethereum state.Trie interface. Such a node
// iterator is used primarily for the implementation of RPC API functions. It
// performs a no-op.
//
// TODO: Determine if we need to implement such functionality for an IAVL tree.
// This will ultimately be related to if we want to support web3.
func (t *Trie) NodeIterator(startKey []byte) ethtrie.NodeIterator {
	return nil
}

// GetKey implements the Ethereum state.Trie interface. Since the IAVL does not
// need to store preimages of keys, a simple identity can be returned.
func (t *Trie) GetKey(key []byte) []byte {
	return key
}

// Prove implements the Ethereum state.Trie interface. It writes a Merkle proof
// to a ethdb.Putter, proofDB, for a given key starting at fromLevel.
//
// TODO: Determine how to integrate this with Cosmos SDK to provide such
// proofs.
func (t *Trie) Prove(key []byte, fromLevel uint, proofDB ethdb.Putter) error {
	return nil
}

// IsStorageTrie returns a boolean reflecting if the Trie is created for
// contract storage.
func (t *Trie) IsStorageTrie() bool {
	return t.prefix != nil
}

// versionFromRootHash returns a Cosmos SDK IAVL version from an Ethereum state
// root hash.
//
// CONTRACT: The encoded version is the eight MSB bytes of the root hash.
func versionFromRootHash(root ethcmn.Hash) int64 {
	return int64(binary.BigEndian.Uint64(root[:versionLen]))
}

// rootHashFromVersion returns a state root hash from a Cosmos SDK IAVL
// version.
func rootHashFromVersion(version int64) (root ethcmn.Hash) {
	binary.BigEndian.PutUint64(root[:versionLen], uint64(version))
	return
}

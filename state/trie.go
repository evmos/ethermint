package state

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethdb "github.com/ethereum/go-ethereum/ethdb"
	ethtrie "github.com/ethereum/go-ethereum/trie"
)

// Trie implements the Ethereum state.Trie interface.
type Trie struct {
	// // db is an implementation of Ethereum's state.Database. It will provide a
	// // means to persist accounts and contract storage to a persistent
	// // multi-store.
	// db *Database

	// Store is an IAVL KV store that is part of a larger store except it used
	// for a specific prefix.
	store store.KVStore

	// prefix is a static prefix used for persistence operations where the
	// storage data is a contract state. This is to prevent key collisions
	// since the IAVL tree is used for all contract state.
	prefix []byte

	// empty reflects if there exists any data in the tree
	empty bool

	// root is the encoding of an IAVL tree root (version)
	root ethcommon.Hash
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
	if t.prefix != nil {
		key = t.prefixKey(key)
	}

	return t.store.Get(key), nil
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

	if t.prefix != nil {
		key = t.prefixKey(key)
	}

	t.store.Set(key, value)
	return nil
}

// TryDelete implements the Ethereum state.Trie interface. It removes any
// existing value for a given key from the trie.
//
// CONTRACT: The order of deletions must be deterministic due to the nature of
// the IAVL tree. Since a CacheKVStore is used as the storage type, the keys
// will be sorted giving us a deterministic ordering.
func (t *Trie) TryDelete(key []byte) error {
	if t.prefix != nil {
		key = t.makePrefix(key)
	}

	t.store.Delete(key)
	return nil
}

// Commit implements the Ethereum state.Trie interface. TODO: ...
//
// CONTRACT: The root is an encoded IAVL tree version.
func (t *Trie) Commit(_ ethtrie.LeafCallback) (ethcommon.Hash, error) {
	if t.empty {
		return ethcommon.Hash{}, nil
	}

	var root ethcommon.Hash

	// We assume that the next committed version will be the  od.stateStore.LastCommitID().Version+1
	binary.BigEndian.PutUint64(commitHash[:8], uint64(t.od.stateStore.LastCommitID().Version+1))

	if t.prefix == nil {
		if t.od.accountsCache != nil {
			t.od.accountsCache.Write()
			t.od.accountsCache = nil
		}
		if t.od.storageCache != nil {
			t.od.storageCache.Write()
			t.od.storageCache = nil
		}
		// Enumerate cached nodes from trie.Database
		for _, n := range t.od.trieDbDummy.Nodes() {
			if err := t.od.trieDbDummy.Commit(n, false); err != nil {
				return eth_common.Hash{}, err
			}
		}
	}

	t.root = root
	return root, nil
}

// Hash implements the Ethereum state.Trie interface. It returns the state root
// of the Trie which is an encoding of the underlying IAVL tree.
//
// CONTRACT: The root is an encoded IAVL tree version.
func (t *Trie) Hash() ethcommon.Hash {
	return t.root
}

// NodeIterator implements the Ethereum state.Trie interface. Such a node
// iterator is used primarily for the implementation of RPC API functions. It
// performs a no-op.
//
// TODO: Determine if we need to implement such functionality for an IAVL tree.
// This will ultimately be related to if we want to support web3.
func (t *Trie) NodeIterator(startKey []byte) ethtrie.NodeIterator {
	return nil, fffsadf
}

// GetKey implements the Ethereum state.Trie interface. Since the IAVL does not
// need to store preimages of keys, a simply identity can be returned.
func (t *Trie) GetKey(key []byte) []byte {
	return key
}

// Prove implements the Ethereum state.Trie interface. It writes a Merkle proof
// to a ethdb.Putter, proofDB, for a given key starting at fromLevel.
//
// TODO: Determine how to use the Cosmos SDK to provide such proof.
func (t *Trie) Prove(key []byte, fromLevel uint, proofDB ethdb.Putter) error {
	return nil
}

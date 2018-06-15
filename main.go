package main

import (
	"fmt"

	eth_common "github.com/ethereum/go-ethereum/common"
	eth_state "github.com/ethereum/go-ethereum/core/state"
	eth_ethdb "github.com/ethereum/go-ethereum/ethdb"
	eth_trie "github.com/ethereum/go-ethereum/trie"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	// Key for the sub-store with Ethereum accounts
	AccountsKey= types.NewKVStoreKey("account")
	// Key for the sub-store with storage data of Ethereum contracts
	StorageKey = types.NewKVStoreKey("storage")
)

// Implementation of eth_state.Database
type OurDatabase struct {
	stateStore        store.CommitMultiStore // For the history of accounts <balance, nonce, storage root hash, code hash>
										     // Also, for the history of contract data (effects of SSTORE instruction)
	lookupDb          dbm.DB // Maping [accounts_trie_root_hash] => <version_id>.
	                         // This mapping exists so that we can implement OpenTrie function of the state.Database interface
                             // Also mapping [storage_trie_root_hash] => <contract_address, version_id>
	                         // This mapping exists so that we can implement OpenStorageTrie function of the state.Database interface
}

func OurNewDatabase(stateDb, lookupDb dbm.DB) *OurDatabase {
	od := &OurDatabase{}
	od.stateStore = store.NewCommitMultiStore(stateDb)
	od.stateStore.MountStoreWithDB(AccountsKey, types.StoreTypeIAVL, nil)
	od.stateStore.MountStoreWithDB(StorageKey, types.StoreTypeIAVL, nil)
	od.lookupDb = lookupDb
	return od
}

func (od *OurDatabase) OpenTrie(root eth_common.Hash) (eth_state.Trie, error) {
	// Need to map root hash to the CommitID to be able to load the trie
	return &OurTrie{}, nil
}

func (od *OurDatabase) OpenStorageTrie(addrHash, root eth_common.Hash) (eth_state.Trie, error) {
	return nil, nil
}

func (od *OurDatabase) CopyTrie(eth_state.Trie) eth_state.Trie {
	return nil
}

func (od *OurDatabase) ContractCode(addrHash, codeHash eth_common.Hash) ([]byte, error) {
	return nil, nil
}

func (od *OurDatabase) ContractCodeSize(addrHash, codeHash eth_common.Hash) (int, error) {
	return 0, nil
}

func (od *OurDatabase) TrieDB() *eth_trie.Database {
	return nil
}

// Implementation of eth_state.Trie
type OurTrie struct {
	// This is essentially part of the KVStore for a specific prefix
	st store.KVStore
	prefix []byte
}

func (ot *OurTrie) TryGet(key []byte) ([]byte, error) {
	return nil, nil
}

func (ot *OurTrie) TryUpdate(key, value []byte) error {
	return nil
}

func (ot *OurTrie) TryDelete(key []byte) error {
	return nil
}

func (ot *OurTrie) Commit(onleaf eth_trie.LeafCallback) (eth_common.Hash, error) {
	return eth_common.Hash{}, nil
}

func (ot *OurTrie) Hash() eth_common.Hash {
	return eth_common.Hash{}
}

func (ot *OurTrie) NodeIterator(startKey []byte) eth_trie.NodeIterator {
	return nil
}

func (ot *OurTrie) GetKey([]byte) []byte {
	return nil
}

func (ot *OurTrie) Prove(key []byte, fromLevel uint, proofDb eth_ethdb.Putter) error {
	return nil
}

func main() {
	fmt.Printf("Instantiating state.Database\n")
	stateDb := dbm.NewDB("state" /* name */, dbm.MemDBBackend, "" /* dir */)
	lookupDb := dbm.NewDB("lookup" /* name */, dbm.MemDBBackend, "" /* dir */)
	var d eth_state.Database
	d = OurNewDatabase(stateDb, lookupDb)
	d.OpenTrie(eth_common.Hash{})
}
package main

import (
	"fmt"

	eth_common "github.com/ethereum/go-ethereum/common"
	eth_state "github.com/ethereum/go-ethereum/core/state"
	eth_ethdb "github.com/ethereum/go-ethereum/ethdb"
	eth_trie "github.com/ethereum/go-ethereum/trie"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/cosmos/cosmos-sdk/store"
)

// Implementation of eth_state.Database
type OurDatabase struct {
	st store.CommitStore
}

func OurNewDatabase(db dbm.DB, id store.CommitID) (od *OurDatabase, err error) {
	od = &OurDatabase{}
	if od.st, err = store.LoadIAVLStore(db, id); err != nil {
		return nil, err
	}
	return
}

func (od *OurDatabase) OpenTrie(root eth_common.Hash) (eth_state.Trie, error) {
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
	db := dbm.NewDB("test" /* name */, dbm.MemDBBackend, "" /* dir */)
	var d eth_state.Database
	var err error
	if d, err = OurNewDatabase(db, store.CommitID{}); err != nil {
		panic(err)
	}
	d.OpenTrie(eth_common.Hash{})
}
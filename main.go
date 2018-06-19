package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	eth_common "github.com/ethereum/go-ethereum/common"
	eth_core "github.com/ethereum/go-ethereum/core"
	eth_state "github.com/ethereum/go-ethereum/core/state"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	eth_rlp "github.com/ethereum/go-ethereum/rlp"
	eth_ethdb "github.com/ethereum/go-ethereum/ethdb"
	eth_params "github.com/ethereum/go-ethereum/params"
	eth_trie "github.com/ethereum/go-ethereum/trie"

	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/go-amino"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
)

var (
	// Key for the sub-store with Ethereum accounts
	AccountsKey = types.NewKVStoreKey("account")
	// Key for the sub-store with storage data of Ethereum contracts
	StorageKey = types.NewKVStoreKey("storage")
	// Key for the sub-store with the code for contracts
	CodeKey = types.NewKVStoreKey("code")
)

// This is what stored in the lookupDb
type LookupValue struct {
	VersionId int64
}

// Implementation of eth_state.Database
type OurDatabase struct {
	stateStore        store.CommitMultiStore // For the history of accounts <balance, nonce, storage root hash, code hash>
										     // Also, for the history of contract data (effects of SSTORE instruction)
	lookupDb          dbm.DB // Maping [trie_root_hash] => <version_id>.
	                         // This mapping exists so that we can implement OpenTrie and OpenStorageTrie functions 
	                         // of the state.Database interface
	codeDb            dbm.DB // Mapping [codeHash] -> <code>
	addrPreimageDb    dbm.DB // Mapping [contract_address_hash] -> <contract_address>
	cdc               *amino.Codec // Amino codec to encode the values forthe lookupDb
}

func OurNewDatabase(stateDb, lookupDb, addrPreimageDb, codeDb dbm.DB) (*OurDatabase, error) {
	od := &OurDatabase{}
	od.stateStore = store.NewCommitMultiStore(stateDb)
	od.stateStore.MountStoreWithDB(AccountsKey, types.StoreTypeIAVL, nil)
	od.stateStore.MountStoreWithDB(StorageKey, types.StoreTypeIAVL, nil)
	if err := od.stateStore.LoadLatestVersion(); err != nil {
		return nil, err
	}
	od.lookupDb = lookupDb
	od.addrPreimageDb = addrPreimageDb
	od.codeDb = codeDb
	od.cdc = amino.NewCodec()
	return od, nil
}

func (od *OurDatabase) OpenTrie(root eth_common.Hash) (eth_state.Trie, error) {
	// Look up version id to use
	if root != (eth_common.Hash{}) {
		val := od.lookupDb.Get(root[:])
		if val == nil {
			return nil, fmt.Errorf("Could not find version with root hash %x", root[:])
		}
		var versionId int64
		_, err := od.cdc.UnmarshalBinaryReader(bytes.NewBuffer(val), &versionId, 0)
		if err != nil {
			return nil, err
		}
		od.stateStore.LoadVersion(versionId)
	}
	st := od.stateStore.GetCommitKVStore(AccountsKey)
	return &OurTrie{od: od, st: st, prefix: nil}, nil
}

func (od *OurDatabase) OpenStorageTrie(addrHash, root eth_common.Hash) (eth_state.Trie, error) {
	if root != (eth_common.Hash{}) {
		val := od.lookupDb.Get(root[:])
		if val == nil {
			return nil, fmt.Errorf("Could not find version with root hash %x", root[:])
		}
		var versionId int64
		_, err := od.cdc.UnmarshalBinaryReader(bytes.NewBuffer(val), &versionId, 0)
		if err != nil {
			return nil, err
		}
		od.stateStore.LoadVersion(versionId)     // This might not be required,
		                                        // we just need to check that accounts and storage are consistent
	}
	st := od.stateStore.GetCommitKVStore(StorageKey)
	return &OurTrie{od:od, st: st, prefix: addrHash[:]}, nil
}

func (od *OurDatabase) CopyTrie(eth_state.Trie) eth_state.Trie {
	return nil
}

func (od *OurDatabase) ContractCode(addrHash, codeHash eth_common.Hash) ([]byte, error) {
	code := od.codeDb.Get(codeHash[:])
	return code, nil
}

func (od *OurDatabase) ContractCodeSize(addrHash, codeHash eth_common.Hash) (int, error) {
	code := od.codeDb.Get(codeHash[:])
	return len(code), nil
}

func (od *OurDatabase) TrieDB() *eth_trie.Database {
	return nil
}

// Implementation of eth_state.Trie
type OurTrie struct {
	od *OurDatabase
	// This is essentially part of the KVStore for a specific prefix
	st store.CommitKVStore
	prefix []byte
}

func (ot *OurTrie) makePrefix(key []byte) []byte {
	kk := make([]byte, len(ot.prefix)+len(key))
	copy(kk, ot.prefix)
	copy(kk[len(ot.prefix):], key)
	return kk
}

func (ot *OurTrie) TryGet(key []byte) ([]byte, error) {
	if ot.prefix == nil {
		return ot.st.Get(key), nil
	}
	return ot.st.Get(ot.makePrefix(key)), nil
}

func (ot *OurTrie) TryUpdate(key, value []byte) error {
	if ot.prefix == nil {
		ot.st.Set(key, value)
		return nil
	}
	ot.st.Set(ot.makePrefix(key), value)
	return nil
}

func (ot *OurTrie) TryDelete(key []byte) error {
	if ot.prefix == nil {
		ot.st.Delete(key)
		return nil
	}
	ot.st.Delete(ot.makePrefix(key))
	return nil
}

func (ot *OurTrie) Commit(onleaf eth_trie.LeafCallback) (eth_common.Hash, error) {
	commitId := ot.st.Commit()
	var hash eth_common.Hash
	copy(hash[:], commitId.Hash)
	b, err := ot.od.cdc.MarshalBinary(commitId.Version)
	if err != nil {
		return hash, err
	}
	ot.od.lookupDb.Set(hash[:], b)
	return hash, nil
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
	addrPreimageDb := dbm.NewDB("addrPreimage" /* name */, dbm.MemDBBackend, "" /* dir */)
	codeDb := dbm.NewDB("code" /* name */, dbm.MemDBBackend, "" /* dir */)
	d, err := OurNewDatabase(stateDb, lookupDb, addrPreimageDb, codeDb)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Instantiating state.StateDB\n")
	// With empty root hash, i.e. empty state
	statedb, err := eth_state.New(eth_common.Hash{}, d)
	if err != nil {
		panic(err)
	}
	g := eth_core.DefaultGenesisBlock()
	for addr, account := range g.Alloc {
		statedb.AddBalance(addr, account.Balance)
		statedb.SetCode(addr, account.Code)
		statedb.SetNonce(addr, account.Nonce)
		for key, value := range account.Storage {
			statedb.SetState(addr, key, value)
		}
	}

	// One of the genesis account having 200 ETH
	b := statedb.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0"))
	fmt.Printf("Balance: %s\n", b)
	genesis_root, err := statedb.Commit(false /* deleteEmptyObjects */)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Genesis state root hash: %x\n", genesis_root[:])
	// File with blockchain data exported from geth by using "geth expordb" command
	input, err := os.Open("/Users/alexeyakhunov/mygit/blockchain")
	if err != nil {
		panic(err)
	}
	defer input.Close()
	// Ethereum mainnet config
	chainConfig := eth_params.MainnetChainConfig
	stream := eth_rlp.NewStream(input, 0)
	var block eth_types.Block
	n := 0
	var root500 eth_common.Hash // Root hash after block 500
	var root501 eth_common.Hash // Root hash after block 501
	for {
		if err = stream.Decode(&block); err == io.EOF {
			err = nil // Clear it
			break
		} else if err != nil {
			panic(fmt.Errorf("at block %d: %v", n, err))
		}
		// don't import first block
		if block.NumberU64() == 0 {
			continue
		}
		header := block.Header()
		// Apply mining rewards to the statedb
		accumulateRewards(chainConfig, statedb, header, block.Uncles())
		// Commit block
		root, err := statedb.Commit(chainConfig.IsEIP158(block.Number()) /* deleteEmptyObjects */)
		if err != nil {
			panic(err)
		}
		switch n {
		case 500:
			root500 = root
		case 501:
			root501 = root
		}
		n++
		if n >= 1000 {
			break
		}
	}
	fmt.Printf("Processed %d blocks\n", n)
	genesis_state, err := eth_state.New(genesis_root, d)
	fmt.Printf("Balance of one of the genesis investors: %s\n", genesis_state.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")))
	miner501 := eth_common.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D") // Miner of the block 501
	// Try to create a new statedb from root of the block 500
	state500, err := eth_state.New(root500, d)
	if err != nil {
		panic(err)
	}
	miner501_balance_at_500 := state500.GetBalance(miner501)
	state501, err := eth_state.New(root501, d)
	if err != nil {
		panic(err)
	}
	miner501_balance_at_501 := state501.GetBalance(miner501)
	fmt.Printf("Miner of block 501's balance after block 500: %d\n", miner501_balance_at_500)
	fmt.Printf("Miner of block 501's balance after block 501: %d\n", miner501_balance_at_501)
}
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"

	eth_common "github.com/ethereum/go-ethereum/common"
	eth_core "github.com/ethereum/go-ethereum/core"
	eth_state "github.com/ethereum/go-ethereum/core/state"
	eth_types "github.com/ethereum/go-ethereum/core/types"
	eth_vm "github.com/ethereum/go-ethereum/core/vm"
	eth_rlp "github.com/ethereum/go-ethereum/rlp"
	eth_ethdb "github.com/ethereum/go-ethereum/ethdb"
	eth_params "github.com/ethereum/go-ethereum/params"
	eth_rpc "github.com/ethereum/go-ethereum/rpc"
	eth_trie "github.com/ethereum/go-ethereum/trie"
	eth_consensus "github.com/ethereum/go-ethereum/consensus"
	eth_misc "github.com/ethereum/go-ethereum/consensus/misc"

	dbm "github.com/tendermint/tmlibs/db"
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

type CommitHashPreimage struct {
	VersionId int64
	Prefix []byte
}

var miner501 = eth_common.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D")

// Implementation of eth_state.Database
type OurDatabase struct {
	stateStore        store.CommitMultiStore // For the history of accounts <balance, nonce, storage root hash, code hash>
										     // Also, for the history of contract data (effects of SSTORE instruction)
	accountsCache     store.CacheKVStore
	storageCache      store.CacheKVStore
	codeDb            dbm.DB // Mapping [codeHash] -> <code>
	tracing           bool
	trieDbDummy       *eth_trie.Database
}

func OurNewDatabase(stateDb, codeDb dbm.DB) (*OurDatabase, error) {
	od := &OurDatabase{}
	od.stateStore = store.NewCommitMultiStore(stateDb)
	od.stateStore.MountStoreWithDB(AccountsKey, types.StoreTypeIAVL, nil)
	od.stateStore.MountStoreWithDB(StorageKey, types.StoreTypeIAVL, nil)
	if err := od.stateStore.LoadLatestVersion(); err != nil {
		return nil, err
	}
	od.codeDb = codeDb
	od.trieDbDummy = eth_trie.NewDatabase(&OurEthDb{codeDb: codeDb})
	return od, nil
}

// root is not interpreted as a hash, but as an encoding of version
func (od *OurDatabase) OpenTrie(root eth_common.Hash) (eth_state.Trie, error) {
	// Look up version id to use
	hasData := root != (eth_common.Hash{})
	versionId := od.stateStore.LastCommitID().Version
	if hasData {
		// First 8 bytes encode version
		versionId = int64(binary.BigEndian.Uint64(root[:8]))
		if od.stateStore.LastCommitID().Version != versionId {
			if err := od.stateStore.LoadVersion(versionId); err != nil {
				return nil, err
			}
			od.accountsCache = nil
		}
	}
	if od.accountsCache == nil {
		od.accountsCache = store.NewCacheKVStore(od.stateStore.GetCommitKVStore(AccountsKey))
		od.storageCache = store.NewCacheKVStore(od.stateStore.GetCommitKVStore(StorageKey))
	}
	return &OurTrie{od: od, st: od.accountsCache, prefix: nil, hasData: hasData}, nil
}

func (od *OurDatabase) OpenStorageTrie(addrHash, root eth_common.Hash) (eth_state.Trie, error) {
	hasData := root != (eth_common.Hash{})
	return &OurTrie{od:od, st: od.storageCache, prefix: addrHash[:], hasData: hasData}, nil
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
	return od.trieDbDummy
}

// Implementation of state.Trie from go-ethereum
type OurTrie struct {
	od *OurDatabase
	// This is essentially part of the KVStore for a specific prefix
	st store.KVStore
	prefix []byte
	hasData bool
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
	ot.hasData = true
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
	if !ot.hasData {
		return eth_common.Hash{}, nil
	}
	var commitHash eth_common.Hash
	// We assume here that the next committed version will be od.stateStore.LastCommitID().Version+1
	binary.BigEndian.PutUint64(commitHash[:8], uint64(ot.od.stateStore.LastCommitID().Version+1))
	if ot.prefix == nil {
		if ot.od.accountsCache != nil {
			ot.od.accountsCache.Write()
			ot.od.accountsCache = nil
		}
		if ot.od.storageCache != nil {
			ot.od.storageCache.Write()
			ot.od.storageCache = nil
		}
		// Enumerate cached nodes from trie.Database
		for _, n := range ot.od.trieDbDummy.Nodes() {
			if err := ot.od.trieDbDummy.Commit(n, false); err != nil {
				return eth_common.Hash{}, err
			}
		}
	}
	return commitHash, nil
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

// Dummy implementation of core.ChainContext and consensus Engine from go-ethereum
type OurChainContext struct {
	coinbase eth_common.Address // This is where the transaction fees will go
}

func (occ *OurChainContext) Engine() eth_consensus.Engine {
	return occ
}

func (occ *OurChainContext) GetHeader(eth_common.Hash, uint64) *eth_types.Header {
	return nil
}

func (occ *OurChainContext) Author(header *eth_types.Header) (eth_common.Address, error) {
	return occ.coinbase, nil
}

func (occ *OurChainContext) APIs(chain eth_consensus.ChainReader) []eth_rpc.API {
	return nil
}

func (occ *OurChainContext) CalcDifficulty(chain eth_consensus.ChainReader, time uint64, parent *eth_types.Header) *big.Int {
	return nil
}

func (occ *OurChainContext) Finalize(chain eth_consensus.ChainReader, header *eth_types.Header, state *eth_state.StateDB, txs []*eth_types.Transaction,
		uncles []*eth_types.Header, receipts []*eth_types.Receipt) (*eth_types.Block, error) {
	return nil, nil
}

func (occ *OurChainContext) Prepare(chain eth_consensus.ChainReader, header *eth_types.Header) error {
	return nil
}

func (occ *OurChainContext) Seal(chain eth_consensus.ChainReader, block *eth_types.Block, stop <-chan struct{}) (*eth_types.Block, error) {
	return nil, nil
}

func (occ *OurChainContext) VerifyHeader(chain eth_consensus.ChainReader, header *eth_types.Header, seal bool) error {
	return nil
}

func (occ *OurChainContext) VerifyHeaders(chain eth_consensus.ChainReader, headers []*eth_types.Header, seals []bool) (chan<- struct{}, <-chan error) {
	return nil, nil
}

func (occ *OurChainContext) VerifySeal(chain eth_consensus.ChainReader, header *eth_types.Header) error {
	return nil
}

func (occ *OurChainContext) VerifyUncles(chain eth_consensus.ChainReader, block *eth_types.Block) error {
	return nil
}

// Implementation of ethdb.Database and ethdb.Batch from go-ethereum
type OurEthDb struct {
	codeDb dbm.DB
}

func (oedb *OurEthDb) Put(key []byte, value []byte) error {
	oedb.codeDb.Set(key, value)
	return nil
}

func (oedb *OurEthDb) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (oedb *OurEthDb) Has(key []byte) (bool, error) {
	return false, nil
}

func (oedb *OurEthDb) Delete(key []byte) error {
	return nil
}

func (oedb *OurEthDb) Close() {
}

func (oedb *OurEthDb) NewBatch() eth_ethdb.Batch {
	return oedb
}

func (oedb *OurEthDb) ValueSize() int {
	return 0
}

func (oedb *OurEthDb) Write() error {
	return nil
}

func (oedb *OurEthDb) Reset() {
}

func main() {
	stateDb := dbm.NewDB("state" /* name */, dbm.MemDBBackend, "" /* dir */)
	codeDb := dbm.NewDB("code" /* name */, dbm.MemDBBackend, "" /* dir */)
	d, err := OurNewDatabase(stateDb, codeDb)
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
	commitID := d.stateStore.Commit()
	fmt.Printf("CommitID after genesis: %v\n", commitID)
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
	prev_root := genesis_root
	d.tracing = true
	chainContext := &OurChainContext{}
	vmConfig := eth_vm.Config{}
	for {
		if err = stream.Decode(&block); err == io.EOF {
			err = nil // Clear it
			break
		} else if err != nil {
			panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
		}
		// don't import first block
		if block.NumberU64() == 0 {
			continue
		}
		header := block.Header()
		chainContext.coinbase = header.Coinbase
		statedb, err := eth_state.New(prev_root, d)
		if err != nil {
			panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
		}
		var (
			receipts eth_types.Receipts
			usedGas  = new(uint64)
			allLogs  []*eth_types.Log
			gp       = new(eth_core.GasPool).AddGas(block.GasLimit())
		)
		if chainConfig.DAOForkSupport && chainConfig.DAOForkBlock != nil && chainConfig.DAOForkBlock.Cmp(block.Number()) == 0 {
			eth_misc.ApplyDAOHardFork(statedb)
		}
		for i, tx := range block.Transactions() {
			statedb.Prepare(tx.Hash(), block.Hash(), i)
			var h eth_common.Hash = tx.Hash()
			if bytes.Equal(h[:], eth_common.FromHex("0xc438cfcc3b74a28741bda361032f1c6362c34aa0e1cedff693f31ec7d6a12717")) {
				vmConfig.Tracer = eth_vm.NewStructLogger(&eth_vm.LogConfig{})
				vmConfig.Debug = true
			}
			receipt, _, err := eth_core.ApplyTransaction(chainConfig, chainContext, nil, gp, statedb, header, tx, usedGas, vmConfig)
			if vmConfig.Tracer != nil {
				w, err := os.Create("structlogs.txt")
				if err != nil {
					panic(err)
				}
				encoder := json.NewEncoder(w)
				logs := FormatLogs(vmConfig.Tracer.(*eth_vm.StructLogger).StructLogs())
				if err := encoder.Encode(logs); err != nil {
					panic(err)
				}
				if err := w.Close(); err != nil {
					panic(err)
				}
				vmConfig.Debug = false
				vmConfig.Tracer = nil
			}
			if err != nil {
				panic(fmt.Errorf("at block %d, tx %x: %v", block.NumberU64(), tx.Hash(), err))
			}
			receipts = append(receipts, receipt)
			allLogs = append(allLogs, receipt.Logs...)
		}
		// Apply mining rewards to the statedb
		accumulateRewards(chainConfig, statedb, header, block.Uncles())
		// Commit block
		prev_root, err = statedb.Commit(chainConfig.IsEIP158(block.Number()) /* deleteEmptyObjects */)
		if err != nil {
			panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
		}
		//fmt.Printf("State root after block %d: %x\n", block.NumberU64(), prev_root)
		d.stateStore.Commit()
		//fmt.Printf("CommitID after block %d: %v\n", block.NumberU64(), commitID)
		switch block.NumberU64() {
		case 500:
			root500 = prev_root
		case 501:
			root501 = prev_root
		}
		n++
		if n % 10000 == 0 {
			fmt.Printf("Processed %d blocks\n", n)
		}
		if n >= 100000 {
			break
		}
	}
	fmt.Printf("Processed %d blocks\n", n)
	d.tracing = true
	genesis_state, err := eth_state.New(genesis_root, d)
	fmt.Printf("Balance of one of the genesis investors: %s\n", genesis_state.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")))
	//miner501 := eth_common.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D") // Miner of the block 501
	// Try to create a new statedb from root of the block 500
	fmt.Printf("root500: %x\n", root500[:])
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
	fmt.Printf("Investor's balance after block 500: %d\n", state500.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")))
	fmt.Printf("Miner of block 501's balance after block 500: %d\n", miner501_balance_at_500)
	fmt.Printf("Miner of block 501's balance after block 501: %d\n", miner501_balance_at_501)
}
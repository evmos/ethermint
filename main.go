// The implementation below is to be considered highly unstable and a continual
// WIP. It is a means to replicate and test replaying Ethereum transactions
// using the Cosmos SDK and the EVM. The ultimate result will be what is known
// as Ethermint.
package main

// 	eth_common "github.com/ethereum/go-ethereum/common"
// 	eth_misc "github.com/ethereum/go-ethereum/consensus/misc"
// 	eth_core "github.com/ethereum/go-ethereum/core"
// 	eth_state "github.com/ethereum/go-ethereum/core/state"
// 	eth_types "github.com/ethereum/go-ethereum/core/types"
// 	eth_vm "github.com/ethereum/go-ethereum/core/vm"
// 	eth_params "github.com/ethereum/go-ethereum/params"
// 	eth_rlp "github.com/ethereum/go-ethereum/rlp"
// 	dbm "github.com/tendermint/tendermint/libs/db"

// var (
// 	// TODO: Document...
// 	miner501 = eth_common.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D")
// )

func main() {
	// stateDb := dbm.NewDB("state" /* name */, dbm.MemDBBackend, "" /* dir */)
	// codeDB := dbm.NewDB("code" /* name */, dbm.MemDBBackend, "" /* dir */)
	// d, err := OurNewDatabase(stateDb, codeDB)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Instantiating state.StateDB\n")
	// // With empty root hash, i.e. empty state
	// statedb, err := eth_state.New(eth_common.Hash{}, d)
	// if err != nil {
	// 	panic(err)
	// }
	// g := eth_core.DefaultGenesisBlock()
	// for addr, account := range g.Alloc {
	// 	statedb.AddBalance(addr, account.Balance)
	// 	statedb.SetCode(addr, account.Code)
	// 	statedb.SetNonce(addr, account.Nonce)
	// 	for key, value := range account.Storage {
	// 		statedb.SetState(addr, key, value)
	// 	}
	// }

	// // One of the genesis account having 200 ETH
	// b := statedb.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0"))
	// fmt.Printf("Balance: %s\n", b)
	// genesis_root, err := statedb.Commit(false /* deleteEmptyObjects */)
	// if err != nil {
	// 	panic(err)
	// }
	// commitID := d.stateStore.Commit()
	// fmt.Printf("CommitID after genesis: %v\n", commitID)
	// fmt.Printf("Genesis state root hash: %x\n", genesis_root[:])
	// // File with blockchain data exported from geth by using "geth expordb" command
	// input, err := os.Open("/Users/alexeyakhunov/mygit/blockchain")
	// if err != nil {
	// 	panic(err)
	// }
	// defer input.Close()
	// // Ethereum mainnet config
	// chainConfig := eth_params.MainnetChainConfig
	// stream := eth_rlp.NewStream(input, 0)
	// var block eth_types.Block
	// n := 0
	// var root500 eth_common.Hash // Root hash after block 500
	// var root501 eth_common.Hash // Root hash after block 501
	// prev_root := genesis_root
	// d.tracing = true
	// chainContext := &OurChainContext{}
	// vmConfig := eth_vm.Config{}
	// for {
	// 	if err = stream.Decode(&block); err == io.EOF {
	// 		err = nil // Clear it
	// 		break
	// 	} else if err != nil {
	// 		panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
	// 	}
	// 	// don't import first block
	// 	if block.NumberU64() == 0 {
	// 		continue
	// 	}
	// 	header := block.Header()
	// 	chainContext.coinbase = header.Coinbase
	// 	statedb, err := eth_state.New(prev_root, d)
	// 	if err != nil {
	// 		panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
	// 	}
	// 	var (
	// 		receipts eth_types.Receipts
	// 		usedGas  = new(uint64)
	// 		allLogs  []*eth_types.Log
	// 		gp       = new(eth_core.GasPool).AddGas(block.GasLimit())
	// 	)
	// 	if chainConfig.DAOForkSupport && chainConfig.DAOForkBlock != nil && chainConfig.DAOForkBlock.Cmp(block.Number()) == 0 {
	// 		eth_misc.ApplyDAOHardFork(statedb)
	// 	}
	// 	for i, tx := range block.Transactions() {
	// 		statedb.Prepare(tx.Hash(), block.Hash(), i)
	// 		var h eth_common.Hash = tx.Hash()
	// 		if bytes.Equal(h[:], eth_common.FromHex("0xc438cfcc3b74a28741bda361032f1c6362c34aa0e1cedff693f31ec7d6a12717")) {
	// 			vmConfig.Tracer = eth_vm.NewStructLogger(&eth_vm.LogConfig{})
	// 			vmConfig.Debug = true
	// 		}
	// 		receipt, _, err := eth_core.ApplyTransaction(chainConfig, chainContext, nil, gp, statedb, header, tx, usedGas, vmConfig)
	// 		if vmConfig.Tracer != nil {
	// 			w, err := os.Create("structlogs.txt")
	// 			if err != nil {
	// 				panic(err)
	// 			}
	// 			encoder := json.NewEncoder(w)
	// 			logs := FormatLogs(vmConfig.Tracer.(*eth_vm.StructLogger).StructLogs())
	// 			if err := encoder.Encode(logs); err != nil {
	// 				panic(err)
	// 			}
	// 			if err := w.Close(); err != nil {
	// 				panic(err)
	// 			}
	// 			vmConfig.Debug = false
	// 			vmConfig.Tracer = nil
	// 		}
	// 		if err != nil {
	// 			panic(fmt.Errorf("at block %d, tx %x: %v", block.NumberU64(), tx.Hash(), err))
	// 		}
	// 		receipts = append(receipts, receipt)
	// 		allLogs = append(allLogs, receipt.Logs...)
	// 	}
	// 	// Apply mining rewards to the statedb
	// 	accumulateRewards(chainConfig, statedb, header, block.Uncles())
	// 	// Commit block
	// 	prev_root, err = statedb.Commit(chainConfig.IsEIP158(block.Number()) /* deleteEmptyObjects */)
	// 	if err != nil {
	// 		panic(fmt.Errorf("at block %d: %v", block.NumberU64(), err))
	// 	}
	// 	//fmt.Printf("State root after block %d: %x\n", block.NumberU64(), prev_root)
	// 	d.stateStore.Commit()
	// 	//fmt.Printf("CommitID after block %d: %v\n", block.NumberU64(), commitID)
	// 	switch block.NumberU64() {
	// 	case 500:
	// 		root500 = prev_root
	// 	case 501:
	// 		root501 = prev_root
	// 	}
	// 	n++
	// 	if n%10000 == 0 {
	// 		fmt.Printf("Processed %d blocks\n", n)
	// 	}
	// 	if n >= 100000 {
	// 		break
	// 	}
	// }
	// fmt.Printf("Processed %d blocks\n", n)
	// d.tracing = true
	// genesis_state, err := eth_state.New(genesis_root, d)
	// fmt.Printf("Balance of one of the genesis investors: %s\n", genesis_state.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")))
	// //miner501 := eth_common.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D") // Miner of the block 501
	// // Try to create a new statedb from root of the block 500
	// fmt.Printf("root500: %x\n", root500[:])
	// state500, err := eth_state.New(root500, d)
	// if err != nil {
	// 	panic(err)
	// }
	// miner501_balance_at_500 := state500.GetBalance(miner501)
	// state501, err := eth_state.New(root501, d)
	// if err != nil {
	// 	panic(err)
	// }
	// miner501_balance_at_501 := state501.GetBalance(miner501)
	// fmt.Printf("Investor's balance after block 500: %d\n", state500.GetBalance(eth_common.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")))
	// fmt.Printf("Miner of block 501's balance after block 500: %d\n", miner501_balance_at_500)
	// fmt.Printf("Miner of block 501's balance after block 501: %d\n", miner501_balance_at_501)
}

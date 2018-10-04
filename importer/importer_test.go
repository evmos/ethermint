package importer

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/core"
	"github.com/cosmos/ethermint/state"
	"github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/bank"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethvm "github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"
	ethrlp "github.com/ethereum/go-ethereum/rlp"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var (
	flagDataDir    string
	flagBlockchain string
	flagCPUProfile string

	miner501    = ethcmn.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D")
	genInvestor = ethcmn.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")

	accKey     = sdk.NewKVStoreKey("acc")
	storageKey = sdk.NewKVStoreKey("storage")
	codeKey    = sdk.NewKVStoreKey("code")

	logger = tmlog.NewNopLogger()
)

func init() {
	flag.StringVar(&flagCPUProfile, "cpu-profile", "", "write CPU profile")
	flag.StringVar(&flagDataDir, "datadir", "", "test data directory for state storage")
	flag.StringVar(&flagBlockchain, "blockchain", "data/blockchain", "ethereum block export file (blocks to import)")
	flag.Parse()
}

func newTestCodec() *wire.Codec {
	codec := wire.NewCodec()

	types.RegisterWire(codec)
	wire.RegisterCrypto(codec)

	return codec
}

func createAndTestGenesis(t *testing.T, cms sdk.CommitMultiStore, am auth.AccountMapper) {
	genBlock := ethcore.DefaultGenesisBlock()
	ms := cms.CacheMultiStore()
	ctx := sdk.NewContext(ms, abci.Header{}, false, logger)

	stateDB, err := state.NewCommitStateDB(ctx, am, storageKey, codeKey)
	require.NoError(t, err, "failed to create a StateDB instance")

	// sort the addresses and insertion of key/value pairs matters
	genAddrs := make([]string, len(genBlock.Alloc))
	i := 0
	for addr := range genBlock.Alloc {
		genAddrs[i] = addr.String()
		i++
	}

	sort.Strings(genAddrs)

	for _, addrStr := range genAddrs {
		addr := ethcmn.HexToAddress(addrStr)
		acc := genBlock.Alloc[addr]

		stateDB.AddBalance(addr, acc.Balance)
		stateDB.SetCode(addr, acc.Code)
		stateDB.SetNonce(addr, acc.Nonce)

		for key, value := range acc.Storage {
			stateDB.SetState(addr, key, value)
		}
	}

	// get balance of one of the genesis account having 200 ETH
	b := stateDB.GetBalance(genInvestor)
	require.Equal(t, "200000000000000000000", b.String())

	// commit the stateDB with 'false' to delete empty objects
	//
	// NOTE: Commit does not yet return the intra merkle root (version)
	_, err = stateDB.Commit(false)
	require.NoError(t, err)

	// persist multi-store cache state
	ms.Write()

	// persist multi-store root state
	commitID := cms.Commit()
	require.Equal(t, "F162678AD57BBE352BE0CFCFCD90E394C4781D31", fmt.Sprintf("%X", commitID.Hash))

	// verify account mapper state
	genAcc := am.GetAccount(ctx, sdk.AccAddress(genInvestor.Bytes()))
	require.NotNil(t, genAcc)
	require.Equal(t, sdk.NewIntFromBigInt(b), genAcc.GetCoins().AmountOf(bank.DenomEthereum))
}

func TestImportBlocks(t *testing.T) {
	if flagDataDir == "" {
		flagDataDir = os.TempDir()
	}

	if flagCPUProfile != "" {
		f, err := os.Create(flagCPUProfile)
		require.NoError(t, err, "failed to create CPU profile")

		err = pprof.StartCPUProfile(f)
		require.NoError(t, err, "failed to start CPU profile")
	}

	db := dbm.NewDB("state", dbm.LevelDBBackend, flagDataDir)
	cb := func() {
		fmt.Println("cleaning up")
		os.RemoveAll(flagDataDir)
		pprof.StopCPUProfile()
	}

	trapSignal(cb)

	// create logger, codec and root multi-store
	cdc := newTestCodec()
	cms := store.NewCommitMultiStore(db)

	// create account mapper
	am := auth.NewAccountMapper(
		cdc,
		accKey,
		types.ProtoBaseAccount,
	)

	// mount stores
	keys := []*sdk.KVStoreKey{accKey, storageKey, codeKey}
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, nil)
	}

	// load latest version (root)
	err := cms.LoadLatestVersion()
	require.NoError(t, err)

	// set and test genesis block
	createAndTestGenesis(t, cms, am)

	// open blockchain export file
	blockchainInput, err := os.Open(flagBlockchain)
	require.Nil(t, err)

	defer blockchainInput.Close()

	// ethereum mainnet config
	chainContext := core.NewChainContext()
	vmConfig := ethvm.Config{}
	chainConfig := ethparams.MainnetChainConfig

	// create RLP stream for exported blocks
	stream := ethrlp.NewStream(blockchainInput, 0)
	startTime := time.Now()

	var block ethtypes.Block
	for {
		err = stream.Decode(&block)
		if err == io.EOF {
			break
		}

		require.NoError(t, err, "failed to decode block")

		var (
			usedGas = new(uint64)
			gp      = new(ethcore.GasPool).AddGas(block.GasLimit())
		)

		header := block.Header()
		chainContext.Coinbase = header.Coinbase

		chainContext.SetHeader(block.NumberU64(), header)

		// Create a cached-wrapped multi-store based on the commit multi-store and
		// create a new context based off of that.
		ms := cms.CacheMultiStore()
		ctx := sdk.NewContext(ms, abci.Header{}, false, logger)

		// stateDB, err := state.NewCommitStateDB(ctx, am, storageKey, codeKey)
		// require.NoError(t, err, "failed to create a StateDB instance")

		// if chainConfig.DAOForkSupport && chainConfig.DAOForkBlock != nil && chainConfig.DAOForkBlock.Cmp(block.Number()) == 0 {
		// 	ethmisc.ApplyDAOHardFork(stateDB)
		// }

		for i, tx := range block.Transactions() {
			msCache := ms.CacheMultiStore()
			ctx = ctx.WithMultiStore(msCache)

			stateDB, err := state.NewCommitStateDB(ctx, am, storageKey, codeKey)
			require.NoError(t, err, "failed to create a StateDB instance")

			stateDB.Prepare(tx.Hash(), block.Hash(), i)

			txHash := tx.Hash()
			if bytes.Equal(txHash[:], ethcmn.FromHex("0xc438cfcc3b74a28741bda361032f1c6362c34aa0e1cedff693f31ec7d6a12717")) {
				vmConfig.Tracer = ethvm.NewStructLogger(&ethvm.LogConfig{})
				vmConfig.Debug = true
			}

			_, _, err = ethcore.ApplyTransaction(
				chainConfig, chainContext, nil, gp, stateDB, header, tx, usedGas, vmConfig,
			)
			require.NoError(t, err, "failed to apply tx at block %d; tx: %d", block.NumberU64(), tx.Hash())

			msCache.Write()
		}

		// commit
		ms.Write()
		cms.Commit()

		if block.NumberU64() > 0 && block.NumberU64()%1000 == 0 {
			fmt.Printf("processed block: %d (time so far: %v)\n", block.NumberU64(), time.Since(startTime))
		}
	}
}

func trapSignal(cb func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		recv := <-c
		fmt.Printf("existing; signal: %s\n", recv)

		if cb != nil {
			cb()
		}

		os.Exit(0)
	}()
}

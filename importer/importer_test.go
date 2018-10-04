package importer

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/state"
	"github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/bank"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmlog "github.com/tendermint/tendermint/libs/log"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
)

var (
	datadir    string
	blockchain string

	miner501    = ethcmn.HexToAddress("0x35e8e5dC5FBd97c5b421A80B596C030a2Be2A04D")
	genInvestor = ethcmn.HexToAddress("0x756F45E3FA69347A9A973A725E3C98bC4db0b5a0")

	accKey     = sdk.NewKVStoreKey("acc")
	storageKey = sdk.NewKVStoreKey("storage")
	codeKey    = sdk.NewKVStoreKey("code")

	logger = tmlog.NewNopLogger()
)

func init() {
	flag.StringVar(&datadir, "datadir", "", "test data directory for state storage")
	flag.StringVar(&blockchain, "blockchain", "data/blockchain", "ethereum block export file (blocks to import)")
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
	if datadir == "" {
		datadir = os.TempDir()
	}

	db := dbm.NewDB("state", dbm.LevelDBBackend, datadir)

	defer func() {
		db.Close()
		os.RemoveAll(datadir)
	}()

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

	// // process blocks
	// for block := range blocks {
	// 	// Create a cached-wrapped multi-store based on the commit multi-store and
	// 	// create a new context based off of that.
	// 	ms := cms.CacheMultiStore()
	// 	ctx := sdk.NewContext(ms, abci.Header{}, false, logger)

	// 	// For each transaction, create a new cache-wrapped multi-store based off of
	// 	// the existing cache-wrapped multi-store in to create a transient store in
	// 	// case processing the tx fails.
	// 	for tx := range block.txs {
	// 		msCache := ms.CacheMultiStore()
	// 		ctx = ctx.WithMultiStore(msCache)

	// 		// apply tx

	// 		// check error
	// 		if err != nil {
	// 			msCache.Write()
	// 		}
	// 	}

	// 	// commit
	// 	ms.Write()
	// 	cms.Commit()
	// }
}

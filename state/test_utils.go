package state

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"

	dbm "github.com/tendermint/tendermint/libs/db"
)

type (
	kvPair struct {
		key, value []byte
	}

	code struct {
		hash ethcmn.Hash
		blob []byte
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newTestDatabase() *Database {
	memDB := dbm.NewMemDB()

	cms := store.NewCommitMultiStore(memDB)
	cms.SetPruning(sdk.PruneNothing)
	cms.MountStoreWithDB(types.StoreKeyAccount, sdk.StoreTypeIAVL, nil)
	cms.MountStoreWithDB(types.StoreKeyStorage, sdk.StoreTypeIAVL, nil)

	testDB, err := NewDatabase(cms, memDB, 100)
	if err != nil {
		panic(fmt.Sprintf("failed to create database: %v", err))
	}

	testDB.stateStore.LoadLatestVersion()

	return testDB
}

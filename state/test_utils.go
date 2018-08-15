package state

import (
	"fmt"
	"math/rand"
	"time"

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

	testDB, err := NewDatabase(memDB, memDB, 100)
	if err != nil {
		panic(fmt.Sprintf("failed to create database: %v", err))
	}

	return testDB
}

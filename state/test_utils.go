package state

import (
	"math/rand"
	"time"

	ethcmn "github.com/ethereum/go-ethereum/common"
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

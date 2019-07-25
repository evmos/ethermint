package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

type QueryResProtocolVersion struct {
	Version string `json:"result"`
}

func (q QueryResProtocolVersion) String() string {
	return q.Version
}

type QueryResBalance struct {
	Balance *hexutil.Big `json:"result"`
}

func (q QueryResBalance) String() string {
	return q.Balance.String()
}

type QueryResBlockNumber struct {
	Number *big.Int `json:"result"`
}

func (q QueryResBlockNumber) String() string {
	return q.Number.String()
}

type QueryResStorage struct {
	Value []byte `json:"value"`
}

func (q QueryResStorage) String() string {
	return string(q.Value)
}

type QueryResCode struct {
	Code []byte
}

func (q QueryResCode) String() string {
	return string(q.Code)
}

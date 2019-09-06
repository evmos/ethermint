package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// QueryResProtocolVersion is response type for protocol version query
type QueryResProtocolVersion struct {
	Version string `json:"result"`
}

func (q QueryResProtocolVersion) String() string {
	return q.Version
}

// QueryResBalance is response type for balance query
type QueryResBalance struct {
	Balance *hexutil.Big `json:"result"`
}

func (q QueryResBalance) String() string {
	return q.Balance.String()
}

// QueryResBlockNumber is response type for block number query
type QueryResBlockNumber struct {
	Number *big.Int `json:"result"`
}

func (q QueryResBlockNumber) String() string {
	return q.Number.String()
}

// QueryResStorage is response type for storage query
type QueryResStorage struct {
	Value []byte `json:"value"`
}

func (q QueryResStorage) String() string {
	return string(q.Value)
}

// QueryResCode is response type for code query
type QueryResCode struct {
	Code []byte
}

func (q QueryResCode) String() string {
	return string(q.Code)
}

// QueryResNonce is response type for Nonce query
type QueryResNonce struct {
	Nonce uint64 `json:"result"`
}

func (q QueryResNonce) String() string {
	return string(q.Nonce)
}

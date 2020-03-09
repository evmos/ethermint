package types

import (
	"fmt"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Supported endpoints
const (
	QueryProtocolVersion = "protocolVersion"
	QueryBalance         = "balance"
	QueryBlockNumber     = "blockNumber"
	QueryStorage         = "storage"
	QueryCode            = "code"
	QueryNonce           = "nonce"
	QueryHashToHeight    = "hashToHeight"
	QueryTxLogs          = "txLogs"
	QueryLogsBloom       = "logsBloom"
	QueryLogs            = "logs"
	QueryAccount         = "account"
)

// QueryResProtocolVersion is response type for protocol version query
type QueryResProtocolVersion struct {
	Version string `json:"version"`
}

func (q QueryResProtocolVersion) String() string {
	return q.Version
}

// QueryResBalance is response type for balance query
type QueryResBalance struct {
	Balance string `json:"balance"`
}

func (q QueryResBalance) String() string {
	return q.Balance
}

// QueryResBlockNumber is response type for block number query
type QueryResBlockNumber struct {
	Number int64 `json:"blockNumber"`
}

func (q QueryResBlockNumber) String() string {
	return string(q.Number)
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
	Nonce uint64 `json:"nonce"`
}

func (q QueryResNonce) String() string {
	return string(q.Nonce)
}

// QueryETHLogs is response type for tx logs query
type QueryETHLogs struct {
	Logs []*ethtypes.Log `json:"logs"`
}

func (q QueryETHLogs) String() string {
	return fmt.Sprintf("%+v", q.Logs)
}

// QueryBloomFilter is response type for tx logs query
type QueryBloomFilter struct {
	Bloom ethtypes.Bloom `json:"bloom"`
}

func (q QueryBloomFilter) String() string {
	return string(q.Bloom.Bytes())
}

// QueryAccount is response type for querying Ethereum state objects
type QueryResAccount struct {
	Balance  string `json:"balance"`
	CodeHash []byte `json:"codeHash"`
	Nonce    uint64 `json:"nonce"`
}

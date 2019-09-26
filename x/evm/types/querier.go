package types

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

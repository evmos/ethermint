package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Recipient is a wrapper of the
type Recipient struct {
	Address string
}

// TxData implements the Ethereum transaction data structure. It is used
// solely as intended in Ethereum abiding by the protocol.
type TxData struct {
	AccountNonce uint64     `json:"nonce"`
	Price        sdk.Int    `json:"gasPrice"`
	GasLimit     uint64     `json:"gas"`
	Recipient    *Recipient `json:"to" rlp:"nil"` // nil means contract creation
	Amount       sdk.Int    `json:"value"`
	Payload      []byte     `json:"input"`

	// signature values
	V []byte `json:"v"`
	R []byte `json:"r"`
	S []byte `json:"s"`

	// hash is only used when marshaling to JSON
	Hash string `json:"hash" rlp:"-"`
}

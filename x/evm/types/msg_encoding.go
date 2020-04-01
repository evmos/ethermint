package types

import (
	"github.com/cosmos/ethermint/utils"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

// EncodableTxData implements the Ethereum transaction data structure. It is used
// solely as intended in Ethereum abiding by the protocol.
type EncodableTxData struct {
	AccountNonce uint64          `json:"nonce"`
	Price        string          `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	Recipient    *ethcmn.Address `json:"to" rlp:"nil"` // nil means contract creation
	Amount       string          `json:"value"`
	Payload      []byte          `json:"input"`

	// signature values
	V string `json:"v"`
	R string `json:"r"`
	S string `json:"s"`

	// hash is only used when marshaling to JSON
	Hash *ethcmn.Hash `json:"hash" rlp:"-"`
}

func marshalAmino(td EncodableTxData) (string, error) {
	bz, err := ModuleCdc.MarshalBinaryBare(td)
	return string(bz), err
}

func unmarshalAmino(td *EncodableTxData, text string) error {
	return ModuleCdc.UnmarshalBinaryBare([]byte(text), td)
}

// MarshalAmino defines custom encoding scheme for TxData
func (td TxData) MarshalAmino() (string, error) {
	e := EncodableTxData{
		AccountNonce: td.AccountNonce,
		Price:        utils.MarshalBigInt(td.Price),
		GasLimit:     td.GasLimit,
		Recipient:    td.Recipient,
		Amount:       utils.MarshalBigInt(td.Amount),
		Payload:      td.Payload,

		V: utils.MarshalBigInt(td.V),
		R: utils.MarshalBigInt(td.R),
		S: utils.MarshalBigInt(td.S),

		Hash: td.Hash,
	}

	return marshalAmino(e)
}

// UnmarshalAmino defines custom decoding scheme for TxData
func (td *TxData) UnmarshalAmino(text string) error {
	e := new(EncodableTxData)
	err := unmarshalAmino(e, text)
	if err != nil {
		return err
	}

	td.AccountNonce = e.AccountNonce
	td.GasLimit = e.GasLimit
	td.Recipient = e.Recipient
	td.Payload = e.Payload
	td.Hash = e.Hash

	price, err := utils.UnmarshalBigInt(e.Price)
	if err != nil {
		return err
	}

	if td.Price != nil {
		td.Price.Set(price)
	} else {
		td.Price = price
	}

	amt, err := utils.UnmarshalBigInt(e.Amount)
	if err != nil {
		return err
	}

	if td.Amount != nil {
		td.Amount.Set(amt)
	} else {
		td.Amount = amt
	}

	v, err := utils.UnmarshalBigInt(e.V)
	if err != nil {
		return err
	}

	if td.V != nil {
		td.V.Set(v)
	} else {
		td.V = v
	}

	r, err := utils.UnmarshalBigInt(e.R)
	if err != nil {
		return err
	}

	if td.R != nil {
		td.R.Set(r)
	} else {
		td.R = r
	}

	s, err := utils.UnmarshalBigInt(e.S)
	if err != nil {
		return err
	}

	if td.S != nil {
		td.S.Set(s)
	} else {
		td.S = s
	}

	return nil
}

// TODO: Implement JSON marshaling/ unmarshaling for this type

// TODO: Implement YAML marshaling/ unmarshaling for this type

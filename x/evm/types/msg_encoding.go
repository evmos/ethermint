package types

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

var cdc = codec.New()

func init() {
	RegisterAmino(cdc)
}

// RegisterAmino registers all crypto related types in the given (amino) codec.
func RegisterAmino(cdc *codec.Codec) {
	cdc.RegisterConcrete(EncodableTxData{}, "ethermint/EncodedMessage", nil)
}

// TxData implements the Ethereum transaction data structure. It is used
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
	bz, err := cdc.MarshalBinaryBare(td)
	return string(bz), err
}

func unmarshalAmino(td *EncodableTxData, text string) (err error) {
	return cdc.UnmarshalBinaryBare([]byte(text), td)
}

func marshalBigInt(i *big.Int) string {
	bz, err := i.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(bz)
}

func unmarshalBigInt(s string) (*big.Int, error) {
	ret := new(big.Int)
	err := ret.UnmarshalText([]byte(s))
	return ret, err
}

// MarshalAmino defines custom encoding scheme for TxData
func (td TxData) MarshalAmino() (string, error) {
	e := EncodableTxData{
		AccountNonce: td.AccountNonce,
		Price:        marshalBigInt(td.Price),
		GasLimit:     td.GasLimit,
		Recipient:    td.Recipient,
		Amount:       marshalBigInt(td.Amount),
		Payload:      td.Payload,

		V: marshalBigInt(td.V),
		R: marshalBigInt(td.R),
		S: marshalBigInt(td.S),

		Hash: td.Hash,
	}

	return marshalAmino(e)
}

// UnmarshalAmino defines custom decoding scheme for TxData
func (td *TxData) UnmarshalAmino(text string) (err error) {
	e := new(EncodableTxData)
	err = unmarshalAmino(e, text)
	if err != nil {
		return
	}

	td.AccountNonce = e.AccountNonce
	td.GasLimit = e.GasLimit
	td.Recipient = e.Recipient
	td.Payload = e.Payload
	td.Hash = e.Hash

	price, err := unmarshalBigInt(e.Price)
	if err != nil {
		return
	}
	if td.Price != nil {
		td.Price.Set(price)
	} else {
		td.Price = price
	}

	amt, err := unmarshalBigInt(e.Amount)
	if err != nil {
		return
	}
	if td.Amount != nil {
		td.Amount.Set(amt)
	} else {
		td.Amount = amt
	}

	v, err := unmarshalBigInt(e.V)
	if err != nil {
		return
	}
	if td.V != nil {
		td.V.Set(v)
	} else {
		td.V = v
	}

	r, err := unmarshalBigInt(e.R)
	if err != nil {
		return
	}
	if td.R != nil {
		td.R.Set(r)
	} else {
		td.R = r
	}

	s, err := unmarshalBigInt(e.S)
	if err != nil {
		return
	}
	if td.S != nil {
		td.S.Set(s)
	} else {
		td.S = s
	}

	return
}

// TODO: Implement JSON marshaling/ unmarshaling for this type

// TODO: Implement YAML marshaling/ unmarshaling for this type

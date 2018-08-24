package types

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethsha "github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/pkg/errors"
)

const (
	// TypeTxEthereum reflects an Ethereum Transaction type.
	TypeTxEthereum = "Ethereum"
)

type (
	// Transaction implements the Ethereum transaction structure as an exact
	// replica. It implements the Cosmos sdk.Tx interface. Due to the private
	// fields, it must be replicated here and cannot be embedded or used
	// directly.
	//
	// Note: The transaction also implements the sdk.Msg interface to perform
	// basic validation that is done in the BaseApp.
	Transaction struct {
		Data TxData

		// caches
		hash atomic.Value
		size atomic.Value
		from atomic.Value
	}

	// TxData defines internal Ethereum transaction information
	TxData struct {
		AccountNonce uint64          `json:"nonce"`
		Price        sdk.Int         `json:"gasPrice"`
		GasLimit     uint64          `json:"gas"`
		Recipient    *ethcmn.Address `json:"to"` // nil means contract creation
		Amount       sdk.Int         `json:"value"`
		Payload      []byte          `json:"input"`
		Signature    *EthSignature   `json:"signature"`

		// hash is only used when marshaling to JSON
		Hash *ethcmn.Hash `json:"hash"`
	}

	// EthSignature reflects an Ethereum signature. We wrap this in a structure
	// to support Amino serialization of transactions.
	EthSignature struct {
		v, r, s *big.Int
	}
)

// NewEthSignature returns a new instantiated Ethereum signature.
func NewEthSignature(v, r, s *big.Int) *EthSignature {
	return &EthSignature{v, r, s}
}

func (es *EthSignature) sanitize() {
	if es.v == nil {
		es.v = new(big.Int)
	}
	if es.r == nil {
		es.r = new(big.Int)
	}
	if es.s == nil {
		es.s = new(big.Int)
	}
}

// MarshalAmino defines a custom encoding scheme for a EthSignature.
func (es EthSignature) MarshalAmino() ([3]string, error) {
	es.sanitize()
	return ethSigMarshalAmino(es)
}

// UnmarshalAmino defines a custom decoding scheme for a EthSignature.
func (es *EthSignature) UnmarshalAmino(raw [3]string) error {
	es.sanitize()
	return ethSigUnmarshalAmino(es, raw)
}

// NewTransaction mimics ethereum's NewTransaction function. It returns a
// reference to a new Ethereum Transaction.
func NewTransaction(
	nonce uint64, to ethcmn.Address, amount sdk.Int,
	gasLimit uint64, gasPrice sdk.Int, payload []byte,
) Transaction {

	if len(payload) > 0 {
		payload = ethcmn.CopyBytes(payload)
	}

	txData := TxData{
		Recipient:    &to,
		AccountNonce: nonce,
		Payload:      payload,
		GasLimit:     gasLimit,
		Amount:       amount,
		Price:        gasPrice,
		Signature:    NewEthSignature(new(big.Int), new(big.Int), new(big.Int)),
	}

	return Transaction{Data: txData}
}

// Sign calculates a secp256k1 ECDSA signature and signs the transaction. It
// takes a private key and chainID to sign an Ethereum transaction according to
// EIP155 standard. It mutates the transaction as it populates the V, R, S
// fields of the Transaction's Signature.
func (tx *Transaction) Sign(chainID sdk.Int, priv *ecdsa.PrivateKey) {
	h := rlpHash([]interface{}{
		tx.Data.AccountNonce,
		tx.Data.Price.BigInt(),
		tx.Data.GasLimit,
		tx.Data.Recipient,
		tx.Data.Amount.BigInt(),
		tx.Data.Payload,
		chainID.BigInt(), uint(0), uint(0),
	})

	sig, err := ethcrypto.Sign(h[:], priv)
	if err != nil {
		panic(err)
	}

	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])

	var v *big.Int
	if chainID.Sign() == 0 {
		v = new(big.Int).SetBytes([]byte{sig[64] + 27})
	} else {
		v = big.NewInt(int64(sig[64] + 35))
		chainIDMul := new(big.Int).Mul(chainID.BigInt(), big.NewInt(2))
		v.Add(v, chainIDMul)
	}

	tx.Data.Signature.v = v
	tx.Data.Signature.r = r
	tx.Data.Signature.s = s
}

// Type implements the sdk.Msg interface. It returns the type of the
// Transaction.
func (tx Transaction) Type() string {
	return TypeTxEthereum
}

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an sdk.Error if validation fails.
func (tx Transaction) ValidateBasic() sdk.Error {
	if tx.Data.Price.Sign() != 1 {
		return ErrInvalidValue(DefaultCodespace, "price must be positive")
	}

	if tx.Data.Amount.Sign() != 1 {
		return ErrInvalidValue(DefaultCodespace, "amount must be positive")
	}

	return nil
}

// GetSignBytes performs a no-op and should not be used. It implements the
// sdk.Msg Interface
func (tx Transaction) GetSignBytes() (sigBytes []byte) { return }

// GetSigners performs a no-op and should not be used. It implements the
// sdk.Msg Interface
//
// CONTRACT: The transaction must already be signed.
func (tx Transaction) GetSigners() (signers []sdk.AccAddress) { return }

// GetMsgs returns a single message containing the Transaction itself. It
// implements the Cosmos sdk.Tx interface.
func (tx Transaction) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx}
}

// ConvertTx attempts to converts a Transaction to a new Ethereum transaction
// with the signature set. The signature if first recovered and then a new
// Transaction is created with that signature. If setting the signature fails,
// a panic will be triggered.
//
// TODO: To be removed in #470
func (tx Transaction) ConvertTx(chainID *big.Int) ethtypes.Transaction {
	gethTx := ethtypes.NewTransaction(
		tx.Data.AccountNonce, *tx.Data.Recipient, tx.Data.Amount.BigInt(),
		tx.Data.GasLimit, tx.Data.Price.BigInt(), tx.Data.Payload,
	)

	sig := recoverEthSig(tx.Data.Signature, chainID)
	signer := ethtypes.NewEIP155Signer(chainID)

	gethTx, err := gethTx.WithSignature(signer, sig)
	if err != nil {
		panic(errors.Wrap(err, "failed to convert transaction with a given signature"))
	}

	return *gethTx
}

// TxDecoder returns an sdk.TxDecoder that given raw transaction bytes,
// attempts to decode them into a valid sdk.Tx.
func TxDecoder(codec *wire.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		var tx sdk.Tx

		// The given codec should have all the appropriate message types
		// registered.
		err := codec.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("failed to decode tx").TraceSDK(err.Error())
		}

		return tx, nil
	}
}

// recoverEthSig recovers a signature according to the Ethereum specification.
func recoverEthSig(es *EthSignature, chainID *big.Int) []byte {
	var v byte

	r, s := es.r.Bytes(), es.s.Bytes()
	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)

	if chainID.Sign() == 0 {
		v = byte(es.v.Uint64() - 27)
	} else {
		chainIDMul := new(big.Int).Mul(chainID, big.NewInt(2))
		V := new(big.Int).Sub(es.v, chainIDMul)

		v = byte(V.Uint64() - 35)
	}

	sig[64] = v
	return sig
}

func rlpHash(x interface{}) (h ethcmn.Hash) {
	hasher := ethsha.NewKeccak256()

	rlp.Encode(hasher, x)
	hasher.Sum(h[:0])

	return h
}

func ethSigMarshalAmino(es EthSignature) (raw [3]string, err error) {
	vb, err := es.v.MarshalText()
	if err != nil {
		return raw, err
	}
	rb, err := es.r.MarshalText()
	if err != nil {
		return raw, err
	}
	sb, err := es.s.MarshalText()
	if err != nil {
		return raw, err
	}

	raw[0], raw[1], raw[2] = string(vb), string(rb), string(sb)
	return raw, err
}

func ethSigUnmarshalAmino(es *EthSignature, raw [3]string) (err error) {
	if err = es.v.UnmarshalText([]byte(raw[0])); err != nil {
		return
	}
	if err = es.r.UnmarshalText([]byte(raw[1])); err != nil {
		return
	}
	if err = es.s.UnmarshalText([]byte(raw[2])); err != nil {
		return
	}

	return
}

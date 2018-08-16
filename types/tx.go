package types

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
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

// ----------------------------------------------------------------------------
// Ethereum transaction
// ----------------------------------------------------------------------------

type (
	// Transaction implements the Ethereum transaction structure as an exact
	// copy. It implements the Cosmos sdk.Tx interface. Due to the private
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

	// TxData implements the Ethereum transaction data structure as an exact
	// copy. It is used solely as intended in Ethereum abiding by the protocol
	// except for the payload field which may embed a Cosmos SDK transaction.
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

	// sigCache is used to cache the derived sender and contains
	// the signer used to derive it.
	sigCache struct {
		signer ethtypes.Signer
		from   ethcmn.Address
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

func (tx Transaction) VerifySig(chainID *big.Int) (ethcmn.Address, error) {
	signer := ethtypes.NewEIP155Signer(chainID)
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	// Do not allow unprotected chainID
	if chainID.Sign() == 0 {
		return ethcmn.Address{}, errors.New("Cannot have 0 as ChainID")
	}
	
	signBytes := rlpHash([]interface{}{
		tx.Data.AccountNonce,
		tx.Data.Price.BigInt(),
		tx.Data.GasLimit,
		tx.Data.Recipient,
		tx.Data.Amount.BigInt(),
		tx.Data.Payload,
		chainID, uint(0), uint(0),
	})

	sig := recoverEthSig(tx.Data.Signature, chainID)

	pub, err := ethcrypto.Ecrecover(signBytes[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}

	var addr ethcmn.Address
	copy(addr[:], ethcrypto.Keccak256(pub[1:])[12:])

	tx.from.Store(sigCache{signer: signer, from: addr})
	return addr, nil
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

// HasEmbeddedTx returns a boolean reflecting if the transaction contains an
// SDK transaction or not based on the recipient address.
func (tx Transaction) HasEmbeddedTx(addr ethcmn.Address) bool {
	return bytes.Equal(tx.Data.Recipient.Bytes(), addr.Bytes())
}

// GetEmbeddedTx returns the embedded SDK transaction from an Ethereum
// transaction. It returns an error if decoding the inner transaction fails.
//
// CONTRACT: The payload field of an Ethereum transaction must contain a valid
// encoded SDK transaction.
func (tx Transaction) GetEmbeddedTx(codec *wire.Codec) (EmbeddedTx, sdk.Error) {
	etx := EmbeddedTx{}

	err := codec.UnmarshalBinary(tx.Data.Payload, &etx)
	if err != nil {
		return EmbeddedTx{}, sdk.ErrTxDecode("failed to encode embedded tx")
	}

	return etx, nil
}

// Copies Ethereum tx's Protected function
func (tx Transaction) protected() bool {
	if tx.Data.Signature.v.BitLen() <= 8 {
		v := tx.Data.Signature.v.Uint64()
		return v != 27 && v != 28
	}
	return true
}

// ----------------------------------------------------------------------------
// embedded SDK transaction
// ----------------------------------------------------------------------------

type (
	// EmbeddedTx implements an SDK transaction. It is to be encoded into the
	// payload field of an Ethereum transaction in order to route and handle SDK
	// transactions.
	EmbeddedTx struct {
		Messages   []sdk.Msg   `json:"messages"`
		Fee        auth.StdFee `json:"fee"`
		Signatures [][]byte    `json:"signatures"`
	}

	// embeddedSignDoc implements a simple SignDoc for a EmbeddedTx signer to
	// sign over.
	embeddedSignDoc struct {
		ChainID       string            `json:"chainID"`
		AccountNumber int64             `json:"accountNumber"`
		Sequence      int64             `json:"sequence"`
		Messages      []json.RawMessage `json:"messages"`
		Fee           json.RawMessage   `json:"fee"`
	}

	// EmbeddedTxSign implements a structure for containing the information
	// necessary for building and signing an EmbeddedTx.
	EmbeddedTxSign struct {
		ChainID       string
		AccountNumber int64
		Sequence      int64
		Messages      []sdk.Msg
		Fee           auth.StdFee
	}
)

// GetMsgs implements the sdk.Tx interface. It returns all the SDK transaction
// messages.
func (etx EmbeddedTx) GetMsgs() []sdk.Msg {
	return etx.Messages
}

// GetRequiredSigners returns all the required signers of an SDK transaction
// accumulated from messages. It returns them in a deterministic fashion given
// a list of messages.
func (etx EmbeddedTx) GetRequiredSigners() []sdk.AccAddress {
	seen := map[string]bool{}

	var signers []sdk.AccAddress
	for _, msg := range etx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, sdk.AccAddress(addr))
				seen[addr.String()] = true
			}
		}
	}

	return signers
}

// Bytes returns the EmbeddedTxSign signature bytes for a signer to sign over.
func (ets EmbeddedTxSign) Bytes() ([]byte, error) {
	sigBytes, err := EmbeddedSignBytes(ets.ChainID, ets.AccountNumber, ets.Sequence, ets.Messages, ets.Fee)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(sigBytes)
	return hash[:], nil
}

// EmbeddedSignBytes creates signature bytes for a signer to sign an embedded
// transaction. The signature bytes require a chainID and an account number.
// The signature bytes are JSON encoded.
func EmbeddedSignBytes(chainID string, accnum, sequence int64, msgs []sdk.Msg, fee auth.StdFee) ([]byte, error) {
	var msgsBytes []json.RawMessage
	for _, msg := range msgs {
		msgsBytes = append(msgsBytes, json.RawMessage(msg.GetSignBytes()))
	}

	signDoc := embeddedSignDoc{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
		Messages:      msgsBytes,
		Fee:           json.RawMessage(fee.Bytes()),
	}

	bz, err := typesCodec.MarshalJSON(signDoc)
	if err != nil {
		errors.Wrap(err, "failed to JSON encode EmbeddedSignDoc")
	}

	return bz, nil
}

// ----------------------------------------------------------------------------
// Utilities
// ----------------------------------------------------------------------------

// TxDecoder returns an sdk.TxDecoder that given raw transaction bytes,
// attempts to decode them into a Transaction or an EmbeddedTx or returning an
// error if decoding fails.
func TxDecoder(codec *wire.Codec, sdkAddress ethcmn.Address) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = Transaction{}

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		// The given codec should have all the appropriate message types
		// registered.
		err := codec.UnmarshalBinary(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("failed to decode tx").TraceSDK(err.Error())
		}

		// If the transaction is routed as an SDK transaction, decode and
		// return the embedded transaction.
		if tx.HasEmbeddedTx(sdkAddress) {
			etx, err := tx.GetEmbeddedTx(codec)
			if err != nil {
				return nil, err
			}

			return etx, nil
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

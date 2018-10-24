package types

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethsha "github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
)

// TODO: Move to the EVM module

// message constants
const (
	TypeTxEthereum  = "Ethereum"
	RouteTxEthereum = "evm"
)

// ----------------------------------------------------------------------------
// Ethereum transaction
// ----------------------------------------------------------------------------

var _ sdk.Tx = (*Transaction)(nil)

type (
	// Transaction implements the Ethereum transaction structure as an exact
	// replica. It implements the Cosmos sdk.Tx interface. Due to the private
	// fields, it must be replicated here and cannot be embedded or used
	// directly.
	//
	// Note: The transaction also implements the sdk.Msg interface to perform
	// basic validation that is done in the BaseApp.
	Transaction struct {
		data TxData

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
		Price        *big.Int        `json:"gasPrice"`
		GasLimit     uint64          `json:"gas"`
		Recipient    *ethcmn.Address `json:"to" rlp:"nil"` // nil means contract creation
		Amount       *big.Int        `json:"value"`
		Payload      []byte          `json:"input"`

		// signature values
		V *big.Int `json:"v"`
		R *big.Int `json:"r"`
		S *big.Int `json:"s"`

		// hash is only used when marshaling to JSON
		Hash *ethcmn.Hash `json:"hash" rlp:"-"`
	}

	// sigCache is used to cache the derived sender and contains the signer used
	// to derive it.
	sigCache struct {
		signer ethtypes.Signer
		from   ethcmn.Address
	}
)

// NewTransaction returns a reference to a new Ethereum transaction.
func NewTransaction(
	nonce uint64, to ethcmn.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) *Transaction {

	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, payload)
}

// NewContractCreation returns a reference to a new Ethereum transaction
// designated for contract creation.
func NewContractCreation(
	nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) *Transaction {

	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, payload)
}

func newTransaction(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *Transaction {

	if len(payload) > 0 {
		payload = ethcmn.CopyBytes(payload)
	}

	txData := TxData{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      payload,
		GasLimit:     gasLimit,
		Amount:       new(big.Int),
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}

	if amount != nil {
		txData.Amount.Set(amount)
	}
	if gasPrice != nil {
		txData.Price.Set(gasPrice)
	}

	return &Transaction{data: txData}
}

// Data returns the Transaction's data.
func (tx Transaction) Data() TxData {
	return tx.data
}

// EncodeRLP implements the rlp.Encoder interface.
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements the rlp.Decoder interface.
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(ethcmn.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// Hash hashes the RLP encoding of a transaction.
func (tx *Transaction) Hash() ethcmn.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(ethcmn.Hash)
	}

	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// SigHash returns the RLP hash of a transaction with a given chainID used for
// signing.
func (tx Transaction) SigHash(chainID *big.Int) ethcmn.Hash {
	return rlpHash([]interface{}{
		tx.data.AccountNonce,
		tx.data.Price,
		tx.data.GasLimit,
		tx.data.Recipient,
		tx.data.Amount,
		tx.data.Payload,
		chainID, uint(0), uint(0),
	})
}

// Sign calculates a secp256k1 ECDSA signature and signs the transaction. It
// takes a private key and chainID to sign an Ethereum transaction according to
// EIP155 standard. It mutates the transaction as it populates the V, R, S
// fields of the Transaction's Signature.
func (tx *Transaction) Sign(chainID *big.Int, priv *ecdsa.PrivateKey) {
	txHash := tx.SigHash(chainID)

	sig, err := ethcrypto.Sign(txHash[:], priv)
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
		chainIDMul := new(big.Int).Mul(chainID, big.NewInt(2))
		v.Add(v, chainIDMul)
	}

	tx.data.V = v
	tx.data.R = r
	tx.data.S = s
}

// VerifySig attempts to verify a Transaction's signature for a given chainID.
// A derived address is returned upon success or an error if recovery fails.
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

	// do not allow recovery for transactions with an unprotected chainID
	if chainID.Sign() == 0 {
		return ethcmn.Address{}, errors.New("invalid chainID")
	}

	txHash := tx.SigHash(chainID)
	sig := recoverEthSig(tx.data.R, tx.data.S, tx.data.V, chainID)

	pub, err := ethcrypto.Ecrecover(txHash[:], sig)
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
func (tx Transaction) Type() string { return TypeTxEthereum }

// Route implements the sdk.Msg interface. It returns the route of the
// Transaction.
func (tx Transaction) Route() string { return RouteTxEthereum }

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an sdk.Error if validation fails.
func (tx Transaction) ValidateBasic() sdk.Error {
	if tx.data.Price.Sign() != 1 {
		return ErrInvalidValue(DefaultCodespace, "price must be positive")
	}

	if tx.data.Amount.Sign() != 1 {
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

// hasEmbeddedTx returns a boolean reflecting if the transaction contains an
// SDK transaction or not based on the recipient address.
func (tx Transaction) hasEmbeddedTx(addr ethcmn.Address) bool {
	return bytes.Equal(tx.data.Recipient.Bytes(), addr.Bytes())
}

// GetEmbeddedTx returns the embedded SDK transaction from an Ethereum
// transaction. It returns an error if decoding the inner transaction fails.
//
// CONTRACT: The payload field of an Ethereum transaction must contain a valid
// encoded SDK transaction.
func (tx Transaction) GetEmbeddedTx(codec *codec.Codec) (sdk.Tx, sdk.Error) {
	var etx sdk.Tx

	err := codec.UnmarshalBinary(tx.data.Payload, &etx)
	if err != nil {
		return etx, sdk.ErrTxDecode("failed to decode embedded transaction")
	}

	return etx, nil
}

// ----------------------------------------------------------------------------
// Utilities
// ----------------------------------------------------------------------------

// TxDecoder returns an sdk.TxDecoder that given raw transaction bytes and an
// SDK address, attempts to decode them into a Transaction or an EmbeddedTx or
// returning an error if decoding fails.
func TxDecoder(codec *codec.Codec, sdkAddress ethcmn.Address) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx = Transaction{}

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("transaction bytes are empty")
		}

		err := rlp.DecodeBytes(txBytes, &tx)
		if err != nil {
			return nil, sdk.ErrTxDecode("failed to decode transaction").TraceSDK(err.Error())
		}

		// If the transaction is routed as an SDK transaction, decode and return
		// the embedded SDK transaction.
		if tx.hasEmbeddedTx(sdkAddress) {
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
func recoverEthSig(R, S, Vb, chainID *big.Int) []byte {
	var v byte

	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)

	if chainID.Sign() == 0 {
		v = byte(Vb.Uint64() - 27)
	} else {
		chainIDMul := new(big.Int).Mul(chainID, big.NewInt(2))
		V := new(big.Int).Sub(Vb, chainIDMul)

		v = byte(V.Uint64() - 35)
	}

	sig[64] = v
	return sig
}

func rlpHash(x interface{}) (hash ethcmn.Hash) {
	hasher := ethsha.NewKeccak256()

	rlp.Encode(hasher, x)
	hasher.Sum(hash[:0])

	return
}

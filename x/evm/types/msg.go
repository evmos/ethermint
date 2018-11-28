package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var _ sdk.Msg = MsgEthereumTx{}

// message type and route constants
const (
	TypeMsgEthereumTx  = "ethereum_tx"
	RouteMsgEthereumTx = "evm"
)

// MsgEthereumTx encapsulates an Ethereum transaction as an SDK message.
type (
	MsgEthereumTx struct {
		Data TxData

		// caches
		hash atomic.Value
		size atomic.Value
		from atomic.Value
	}

	// TxData implements the Ethereum transaction data structure. It is used
	// solely as intended in Ethereum abiding by the protocol.
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

// NewMsgEthereumTx returns a reference to a new Ethereum transaction message.
func NewMsgEthereumTx(
	nonce uint64, to ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {

	return newMsgEthereumTx(nonce, &to, amount, gasLimit, gasPrice, payload)
}

// NewMsgEthereumTxContract returns a reference to a new Ethereum transaction
// message designated for contract creation.
func NewMsgEthereumTxContract(
	nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {

	return newMsgEthereumTx(nonce, nil, amount, gasLimit, gasPrice, payload)
}

func newMsgEthereumTx(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {

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

	return &MsgEthereumTx{Data: txData}
}

// Route returns the route value of an MsgEthereumTx.
func (msg MsgEthereumTx) Route() string { return RouteMsgEthereumTx }

// Type returns the type value of an MsgEthereumTx.
func (msg MsgEthereumTx) Type() string { return TypeMsgEthereumTx }

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an sdk.Error if validation fails.
func (msg MsgEthereumTx) ValidateBasic() sdk.Error {
	if msg.Data.Price.Sign() != 1 {
		return types.ErrInvalidValue("price must be positive")
	}

	if msg.Data.Amount.Sign() != 1 {
		return types.ErrInvalidValue("amount must be positive")
	}

	return nil
}

// GetSigners returns the expected signers for an Ethereum transaction message.
// For such a message, there should exist only a single 'signer'.
//
// NOTE: This method cannot be used as a chain ID is needed to recover the signer
// from the signature. Use 'VerifySig' instead.
func (msg MsgEthereumTx) GetSigners() []sdk.AccAddress {
	panic("must use 'VerifySig' with a chain ID to get the signer")
}

// GetSignBytes returns the Amino bytes of an Ethereum transaction message used
// for signing.
//
// NOTE: This method cannot be used as a chain ID is needed to create valid bytes
// to sign over. Use 'RLPSignBytes' instead.
func (msg MsgEthereumTx) GetSignBytes() []byte {
	panic("must use 'RLPSignBytes' with a chain ID to get the valid bytes to sign")
}

// RLPSignBytes returns the RLP hash of an Ethereum transaction message with a
// given chainID used for signing.
func (msg MsgEthereumTx) RLPSignBytes(chainID *big.Int) ethcmn.Hash {
	return rlpHash([]interface{}{
		msg.Data.AccountNonce,
		msg.Data.Price,
		msg.Data.GasLimit,
		msg.Data.Recipient,
		msg.Data.Amount,
		msg.Data.Payload,
		chainID, uint(0), uint(0),
	})
}

// EncodeRLP implements the rlp.Encoder interface.
func (msg *MsgEthereumTx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &msg.Data)
}

// DecodeRLP implements the rlp.Decoder interface.
func (msg *MsgEthereumTx) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()

	err := s.Decode(&msg.Data)
	if err == nil {
		msg.size.Store(ethcmn.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// Hash hashes the RLP encoding of a transaction.
func (msg *MsgEthereumTx) Hash() ethcmn.Hash {
	if hash := msg.hash.Load(); hash != nil {
		return hash.(ethcmn.Hash)
	}

	v := rlpHash(msg)
	msg.hash.Store(v)

	return v
}

// Sign calculates a secp256k1 ECDSA signature and signs the transaction. It
// takes a private key and chainID to sign an Ethereum transaction according to
// EIP155 standard. It mutates the transaction as it populates the V, R, S
// fields of the Transaction's Signature.
func (msg *MsgEthereumTx) Sign(chainID *big.Int, priv *ecdsa.PrivateKey) {
	txHash := msg.RLPSignBytes(chainID)

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

	msg.Data.V = v
	msg.Data.R = r
	msg.Data.S = s
}

// VerifySig attempts to verify a Transaction's signature for a given chainID.
// A derived address is returned upon success or an error if recovery fails.
func (msg MsgEthereumTx) VerifySig(chainID *big.Int) (ethcmn.Address, error) {
	signer := ethtypes.NewEIP155Signer(chainID)

	if sc := msg.from.Load(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous call is not the same as
		// used current, invalidate the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	// do not allow recovery for transactions with an unprotected chainID
	if chainID.Sign() == 0 {
		return ethcmn.Address{}, errors.New("invalid chainID")
	}

	txHash := msg.RLPSignBytes(chainID)
	sig := recoverEthSig(msg.Data.R, msg.Data.S, msg.Data.V, chainID)

	pub, err := ethcrypto.Ecrecover(txHash[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}

	var addr ethcmn.Address
	copy(addr[:], ethcrypto.Keccak256(pub[1:])[12:])

	msg.from.Store(sigCache{signer: signer, from: addr})
	return addr, nil
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

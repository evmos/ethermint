package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	_ sdk.Msg = EthereumTxMsg{}
	_ sdk.Tx  = EthereumTxMsg{}
)

var big8 = big.NewInt(8)

// message type and route constants
const (
	TypeEthereumTxMsg  = "ethereum_tx"
	RouteEthereumTxMsg = RouterKey
)

// EthereumTxMsg encapsulates an Ethereum transaction as an SDK message.
type (
	EthereumTxMsg struct {
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

// NewEthereumTxMsg returns a reference to a new Ethereum transaction message.
func NewEthereumTxMsg(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *EthereumTxMsg {

	return newEthereumTxMsg(nonce, to, amount, gasLimit, gasPrice, payload)
}

// NewEthereumTxMsgContract returns a reference to a new Ethereum transaction
// message designated for contract creation.
func NewEthereumTxMsgContract(
	nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) *EthereumTxMsg {

	return newEthereumTxMsg(nonce, nil, amount, gasLimit, gasPrice, payload)
}

func newEthereumTxMsg(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *EthereumTxMsg {

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

	return &EthereumTxMsg{Data: txData}
}

// Route returns the route value of an EthereumTxMsg.
func (msg EthereumTxMsg) Route() string { return RouteEthereumTxMsg }

// Type returns the type value of an EthereumTxMsg.
func (msg EthereumTxMsg) Type() string { return TypeEthereumTxMsg }

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an sdk.Error if validation fails.
func (msg EthereumTxMsg) ValidateBasic() sdk.Error {
	if msg.Data.Price.Sign() != 1 {
		return types.ErrInvalidValue(fmt.Sprintf("Price must be positive: %x", msg.Data.Price))
	}

	// Amount can be 0
	if msg.Data.Amount.Sign() == -1 {
		return types.ErrInvalidValue(fmt.Sprintf("amount must be positive: %x", msg.Data.Amount))
	}

	return nil
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg EthereumTxMsg) To() *ethcmn.Address {
	if msg.Data.Recipient == nil {
		return nil
	}

	return msg.Data.Recipient
}

// GetMsgs returns a single EthereumTxMsg as an sdk.Msg.
func (msg EthereumTxMsg) GetMsgs() []sdk.Msg {
	return []sdk.Msg{msg}
}

// GetSigners returns the expected signers for an Ethereum transaction message.
// For such a message, there should exist only a single 'signer'.
//
// NOTE: This method cannot be used as a chain ID is needed to recover the signer
// from the signature. Use 'VerifySig' instead.
func (msg EthereumTxMsg) GetSigners() []sdk.AccAddress {
	panic("must use 'VerifySig' with a chain ID to get the signer")
}

// GetSignBytes returns the Amino bytes of an Ethereum transaction message used
// for signing.
//
// NOTE: This method cannot be used as a chain ID is needed to create valid bytes
// to sign over. Use 'RLPSignBytes' instead.
func (msg EthereumTxMsg) GetSignBytes() []byte {
	panic("must use 'RLPSignBytes' with a chain ID to get the valid bytes to sign")
}

// RLPSignBytes returns the RLP hash of an Ethereum transaction message with a
// given chainID used for signing.
func (msg EthereumTxMsg) RLPSignBytes(chainID *big.Int) ethcmn.Hash {
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
func (msg *EthereumTxMsg) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &msg.Data)
}

// DecodeRLP implements the rlp.Decoder interface.
func (msg *EthereumTxMsg) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()

	err := s.Decode(&msg.Data)
	if err == nil {
		msg.size.Store(ethcmn.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// Hash hashes the RLP encoding of a transaction.
func (msg *EthereumTxMsg) Hash() ethcmn.Hash {
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
func (msg *EthereumTxMsg) Sign(chainID *big.Int, priv *ecdsa.PrivateKey) {
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
func (msg *EthereumTxMsg) VerifySig(chainID *big.Int) (ethcmn.Address, error) {
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
		return ethcmn.Address{}, errors.New("chainID cannot be zero")
	}

	chainIDMul := new(big.Int).Mul(chainID, big.NewInt(2))
	V := new(big.Int).Sub(msg.Data.V, chainIDMul)
	V.Sub(V, big8)

	sigHash := msg.RLPSignBytes(chainID)
	sender, err := recoverEthSig(msg.Data.R, msg.Data.S, V, sigHash)
	if err != nil {
		return ethcmn.Address{}, err
	}

	msg.from.Store(sigCache{signer: signer, from: sender})
	return sender, nil
}

// Cost returns amount + gasprice * gaslimit.
func (msg EthereumTxMsg) Cost() *big.Int {
	total := msg.Fee()
	total.Add(total, msg.Data.Amount)
	return total
}

// Fee returns gasprice * gaslimit.
func (msg EthereumTxMsg) Fee() *big.Int {
	return new(big.Int).Mul(msg.Data.Price, new(big.Int).SetUint64(msg.Data.GasLimit))
}

// ChainID returns which chain id this transaction was signed for (if at all)
func (msg *EthereumTxMsg) ChainID() *big.Int {
	return deriveChainID(msg.Data.V)
}

// deriveChainID derives the chain id from the given v parameter
func deriveChainID(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}

// ----------------------------------------------------------------------------
// Auxiliary

// TxDecoder returns an sdk.TxDecoder that can decode both auth.StdTx and
// EthereumTxMsg transactions.
func TxDecoder(cdc *codec.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, sdk.Error) {
		var tx sdk.Tx

		if len(txBytes) == 0 {
			return nil, sdk.ErrTxDecode("txBytes are empty")
		}

		err := cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
		if err != nil {
			fmt.Println(err.Error())
			return nil, sdk.ErrTxDecode("failed to decode tx").TraceSDK(err.Error())
		}

		return tx, nil
	}
}

// recoverEthSig recovers a signature according to the Ethereum specification and
// returns the sender or an error.
//
// Ref: Ethereum Yellow Paper (BYZANTIUM VERSION 69351d5) Appendix F
// nolint: gocritic
func recoverEthSig(R, S, Vb *big.Int, sigHash ethcmn.Hash) (ethcmn.Address, error) {
	if Vb.BitLen() > 8 {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	V := byte(Vb.Uint64() - 27)
	if !ethcrypto.ValidateSignatureValues(V, R, S, true) {
		return ethcmn.Address{}, errors.New("invalid signature")
	}

	// encode the signature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, 65)

	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[64] = V

	// recover the public key from the signature
	pub, err := ethcrypto.Ecrecover(sigHash[:], sig)
	if err != nil {
		return ethcmn.Address{}, err
	}

	if len(pub) == 0 || pub[0] != 4 {
		return ethcmn.Address{}, errors.New("invalid public key")
	}

	var addr ethcmn.Address
	copy(addr[:], ethcrypto.Keccak256(pub[1:])[12:])

	return addr, nil
}

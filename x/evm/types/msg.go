package types

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/cosmos/ethermint/types"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	_ sdk.Msg = MsgEthermint{}
	_ sdk.Msg = MsgEthereumTx{}
	_ sdk.Tx  = MsgEthereumTx{}
)

var big8 = big.NewInt(8)

// message type and route constants
const (
	// TypeMsgEthereumTx defines the type string of an Ethereum tranasction
	TypeMsgEthereumTx = "ethereum"
	// TypeMsgEthermint defines the type string of Ethermint message
	TypeMsgEthermint = "ethermint"
)

// MsgEthermint implements a cosmos equivalent structure for Ethereum transactions
type MsgEthermint struct {
	AccountNonce uint64          `json:"nonce"`
	Price        sdk.Int         `json:"gasPrice"`
	GasLimit     uint64          `json:"gas"`
	Recipient    *sdk.AccAddress `json:"to" rlp:"nil"` // nil means contract creation
	Amount       sdk.Int         `json:"value"`
	Payload      []byte          `json:"input"`

	// From address (formerly derived from signature)
	From sdk.AccAddress `json:"from"`
}

// NewMsgEthermint returns a reference to a new Ethermint transaction
func NewMsgEthermint(
	nonce uint64, to *sdk.AccAddress, amount sdk.Int,
	gasLimit uint64, gasPrice sdk.Int, payload []byte, from sdk.AccAddress,
) MsgEthermint {
	return MsgEthermint{
		AccountNonce: nonce,
		Price:        gasPrice,
		GasLimit:     gasLimit,
		Recipient:    to,
		Amount:       amount,
		Payload:      payload,
		From:         from,
	}
}

func (msg MsgEthermint) String() string {
	return fmt.Sprintf("nonce=%d gasPrice=%d gasLimit=%d recipient=%s amount=%d data=0x%x from=%s",
		msg.AccountNonce, msg.Price, msg.GasLimit, msg.Recipient, msg.Amount, msg.Payload, msg.From)
}

// Route should return the name of the module
func (msg MsgEthermint) Route() string { return RouterKey }

// Type returns the action of the message
func (msg MsgEthermint) Type() string { return TypeMsgEthermint }

// GetSignBytes encodes the message for signing
func (msg MsgEthermint) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// ValidateBasic runs stateless checks on the message
func (msg MsgEthermint) ValidateBasic() error {
	if msg.Price.IsZero() {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "gas price cannot be 0")
	}

	if msg.Price.Sign() == -1 {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "gas price cannot be negative %s", msg.Price)
	}

	// Amount can be 0
	if msg.Amount.Sign() == -1 {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "amount cannot be negative %s", msg.Amount)
	}

	return nil
}

// GetSigners defines whose signature is required
func (msg MsgEthermint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg MsgEthermint) To() *ethcmn.Address {
	if msg.Recipient == nil {
		return nil
	}

	addr := ethcmn.BytesToAddress(msg.Recipient.Bytes())
	return &addr
}

// MsgEthereumTx encapsulates an Ethereum transaction as an SDK message.
type MsgEthereumTx struct {
	Data TxData

	// caches
	size atomic.Value
	from atomic.Value
}

// sigCache is used to cache the derived sender and contains the signer used
// to derive it.
type sigCache struct {
	signer ethtypes.Signer
	from   ethcmn.Address
}

// NewMsgEthereumTx returns a reference to a new Ethereum transaction message.
func NewMsgEthereumTx(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) MsgEthereumTx {
	return newMsgEthereumTx(nonce, to, amount, gasLimit, gasPrice, payload)
}

// NewMsgEthereumTxContract returns a reference to a new Ethereum transaction
// message designated for contract creation.
func NewMsgEthereumTxContract(
	nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) MsgEthereumTx {
	return newMsgEthereumTx(nonce, nil, amount, gasLimit, gasPrice, payload)
}

func newMsgEthereumTx(
	nonce uint64, to *ethcmn.Address, amount *big.Int, // nolint: interfacer
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) MsgEthereumTx {
	if len(payload) > 0 {
		payload = ethcmn.CopyBytes(payload)
	}

	var recipient *Recipient
	if to != nil {
		recipient = &Recipient{Address: to.String()}
	}

	txData := TxData{
		AccountNonce: nonce,
		Recipient:    recipient,
		Payload:      payload,
		GasLimit:     gasLimit,
		Amount:       sdk.ZeroInt(),
		Price:        sdk.ZeroInt(),
		V:            []byte{},
		R:            []byte{},
		S:            []byte{},
	}

	if amount != nil {
		txData.Amount = sdk.NewIntFromBigInt(amount)
	}
	if gasPrice != nil {
		txData.Price = sdk.NewIntFromBigInt(gasPrice)
	}

	return MsgEthereumTx{Data: txData}
}

func (msg MsgEthereumTx) String() string {
	out, _ := yaml.Marshal(msg.Data)
	return string(out)
}

// Route returns the route value of an MsgEthereumTx.
func (msg MsgEthereumTx) Route() string { return RouterKey }

// Type returns the type value of an MsgEthereumTx.
func (msg MsgEthereumTx) Type() string { return TypeMsgEthereumTx }

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an error if validation fails.
func (msg MsgEthereumTx) ValidateBasic() error {
	if msg.Data.Price.IsZero() {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "gas price cannot be 0")
	}

	if msg.Data.Price.IsNegative() {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "gas price cannot be negative %s", msg.Data.Price)
	}

	// Amount can be 0
	if msg.Data.Amount.IsNegative() {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "amount cannot be negative %s", msg.Data.Amount)
	}

	return nil
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg MsgEthereumTx) To() *ethcmn.Address {
	if msg.Data.Recipient == nil {
		return nil
	}

	recipient := ethcmn.HexToAddress(msg.Data.Recipient.Address)
	return &recipient
}

// GetMsgs returns a single MsgEthereumTx as an sdk.Msg.
func (msg MsgEthereumTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{msg}
}

// GetSigners returns the expected signers for an Ethereum transaction message.
// For such a message, there should exist only a single 'signer'.
//
// NOTE: This method panics if 'VerifySig' hasn't been called first.
func (msg MsgEthereumTx) GetSigners() []sdk.AccAddress {
	sender := msg.From()
	if sender.Empty() {
		panic("must use 'VerifySig' with a chain ID to get the signer")
	}
	return []sdk.AccAddress{sender}
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
		msg.Data.Price.BigInt(),
		msg.Data.GasLimit,
		msg.To(),
		msg.Data.Amount.BigInt(),
		msg.Data.Payload,
		chainID, uint(0), uint(0),
	})
}

// EncodeRLP implements the rlp.Encoder interface.
func (msg *MsgEthereumTx) EncodeRLP(w io.Writer) error {
	var hash ethcmn.Hash
	if len(msg.Data.Hash) > 0 {
		hash = ethcmn.HexToHash(msg.Data.Hash)
	}

	data := struct {
		AccountNonce uint64
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
	}{
		AccountNonce: msg.Data.AccountNonce,
		Price:        msg.Data.Price.BigInt(),
		GasLimit:     msg.Data.GasLimit,
		Recipient:    msg.To(),
		Amount:       msg.Data.Amount.BigInt(),
		Payload:      msg.Data.Payload,
		V:            new(big.Int).SetBytes(msg.Data.V),
		R:            new(big.Int).SetBytes(msg.Data.R),
		S:            new(big.Int).SetBytes(msg.Data.S),
		Hash:         &hash,
	}
	return rlp.Encode(w, data)
}

// DecodeRLP implements the rlp.Decoder interface.
func (msg *MsgEthereumTx) DecodeRLP(s *rlp.Stream) error {
	_, size, err := s.Kind()
	if err != nil {
		// return error if stream is too large
		return err
	}

	var data struct {
		AccountNonce uint64
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

	if err := s.Decode(&data); err != nil {
		return err
	}

	var hash string
	if data.Hash != nil {
		hash = data.Hash.String()
	}

	var recipient *Recipient
	if data.Recipient != nil {
		recipient = &Recipient{Address: data.Recipient.String()}
	}

	msg.Data = TxData{
		AccountNonce: data.AccountNonce,
		Price:        sdk.NewIntFromBigInt(data.Price),
		GasLimit:     data.GasLimit,
		Recipient:    recipient,
		Amount:       sdk.NewIntFromBigInt(data.Amount),
		Payload:      data.Payload,
		V:            data.V.Bytes(),
		R:            data.R.Bytes(),
		S:            data.S.Bytes(),
		Hash:         hash,
	}

	msg.size.Store(ethcmn.StorageSize(rlp.ListSize(size)))
	return nil
}

// Sign calculates a secp256k1 ECDSA signature and signs the transaction. It
// takes a private key and chainID to sign an Ethereum transaction according to
// EIP155 standard. It mutates the transaction as it populates the V, R, S
// fields of the Transaction's Signature.
func (msg *MsgEthereumTx) Sign(chainID *big.Int, priv *ecdsa.PrivateKey) error {
	txHash := msg.RLPSignBytes(chainID)

	sig, err := ethcrypto.Sign(txHash[:], priv)
	if err != nil {
		return err
	}

	if len(sig) != 65 {
		return fmt.Errorf("wrong size for signature: got %d, want 65", len(sig))
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

	msg.Data.V = v.Bytes()
	msg.Data.R = r.Bytes()
	msg.Data.S = s.Bytes()
	return nil
}

// VerifySig attempts to verify a Transaction's signature for a given chainID.
// A derived address is returned upon success or an error if recovery fails.
func (msg *MsgEthereumTx) VerifySig(chainID *big.Int) (ethcmn.Address, error) {
	v, r, s := msg.RawSignatureValues()
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
	V := new(big.Int).Sub(v, chainIDMul)
	V.Sub(V, big8)

	sigHash := msg.RLPSignBytes(chainID)
	sender, err := recoverEthSig(r, s, V, sigHash)
	if err != nil {
		return ethcmn.Address{}, err
	}

	msg.from.Store(sigCache{signer: signer, from: sender})
	return sender, nil
}

// GetGas implements the GasTx interface. It returns the GasLimit of the transaction.
func (msg MsgEthereumTx) GetGas() uint64 {
	return msg.Data.GasLimit
}

// Fee returns gasprice * gaslimit.
func (msg MsgEthereumTx) Fee() *big.Int {
	gasPrice := msg.Data.Price.BigInt()
	gasLimit := new(big.Int).SetUint64(msg.Data.GasLimit)
	return new(big.Int).Mul(gasPrice, gasLimit)
}

// ChainID returns which chain id this transaction was signed for (if at all)
func (msg *MsgEthereumTx) ChainID() *big.Int {
	v := new(big.Int).SetBytes(msg.Data.V)
	return deriveChainID(v)
}

// Cost returns amount + gasprice * gaslimit.
func (msg MsgEthereumTx) Cost() *big.Int {
	total := msg.Fee()
	total.Add(total, msg.Data.Amount.BigInt())
	return total
}

// RawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (msg MsgEthereumTx) RawSignatureValues() (v, r, s *big.Int) {
	return new(big.Int).SetBytes(msg.Data.V),
		new(big.Int).SetBytes(msg.Data.R),
		new(big.Int).SetBytes(msg.Data.S)
}

// From loads the ethereum sender address from the sigcache and returns an
// sdk.AccAddress from its bytes
func (msg *MsgEthereumTx) From() sdk.AccAddress {
	sc := msg.from.Load()
	if sc == nil {
		return nil
	}

	sigCache := sc.(sigCache)

	if len(sigCache.from.Bytes()) == 0 {
		return nil
	}

	return sdk.AccAddress(sigCache.from.Bytes())
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

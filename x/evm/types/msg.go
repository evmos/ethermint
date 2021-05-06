package types

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	_ sdk.Msg = &MsgEthereumTx{}
	_ sdk.Tx  = &MsgEthereumTx{}
	// _ ethtypes.TxData = &MsgEthereumTx{}
)

var big8 = big.NewInt(8)

// message type and route constants
const (
	// TypeMsgEthereumTx defines the type string of an Ethereum transaction
	TypeMsgEthereumTx = "ethereum"
)

// NewMsgEthereumTx returns a reference to a new Ethereum transaction message.
func NewMsgEthereumTx(
	nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {
	return newMsgEthereumTx(nonce, to, amount, gasLimit, gasPrice, payload)
}

// NewMsgEthereumTxContract returns a reference to a new Ethereum transaction
// message designated for contract creation.
func NewMsgEthereumTxContract(
	nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {
	return newMsgEthereumTx(nonce, nil, amount, gasLimit, gasPrice, payload)
}

func newMsgEthereumTx(
	nonce uint64, to *ethcmn.Address, amount *big.Int, // nolint: interfacer
	gasLimit uint64, gasPrice *big.Int, payload []byte,
) *MsgEthereumTx {
	if len(payload) > 0 {
		payload = ethcmn.CopyBytes(payload)
	}

	var toBz []byte
	if to != nil {
		toBz = to.Bytes()
	}

	// TODO: chain id and access list
	txData := &TxData{
		ChainID:  nil,
		Nonce:    nonce,
		To:       toBz,
		Input:    payload,
		GasLimit: gasLimit,
		Amount:   []byte{},
		GasPrice: []byte{},
		Accesses: nil,
		V:        []byte{},
		R:        []byte{},
		S:        []byte{},
	}

	if amount != nil {
		txData.Amount = amount.Bytes()
	}
	if gasPrice != nil {
		txData.GasPrice = gasPrice.Bytes()
	}

	return &MsgEthereumTx{Data: txData}
}

// Route returns the route value of an MsgEthereumTx.
func (msg MsgEthereumTx) Route() string { return RouterKey }

// Type returns the type value of an MsgEthereumTx.
func (msg MsgEthereumTx) Type() string { return TypeMsgEthereumTx }

// ValidateBasic implements the sdk.Msg interface. It performs basic validation
// checks of a Transaction. If returns an error if validation fails.
func (msg MsgEthereumTx) ValidateBasic() error {
	gasPrice := new(big.Int).SetBytes(msg.Data.GasPrice)
	// if gasPrice.Sign() == 0 {
	// 	return sdkerrors.Wrapf(ErrInvalidValue, "gas price cannot be 0")
	// }

	if gasPrice.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidValue, "gas price cannot be negative %s", gasPrice)
	}

	// Amount can be 0
	amount := new(big.Int).SetBytes(msg.Data.Amount)
	if amount.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidValue, "amount cannot be negative %s", amount)
	}

	return nil
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg MsgEthereumTx) To() *ethcmn.Address {
	if len(msg.Data.To) == 0 {
		return nil
	}

	recipient := ethcmn.BytesToAddress(msg.Data.To)
	return &recipient
}

// GetMsgs returns a single MsgEthereumTx as an sdk.Msg.
func (msg *MsgEthereumTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{msg}
}

// GetSigners returns the expected signers for an Ethereum transaction message.
// For such a message, there should exist only a single 'signer'.
//
// NOTE: This method panics if 'VerifySig' hasn't been called first.
func (msg MsgEthereumTx) GetSigners() []sdk.AccAddress {
	if IsZeroAddress(msg.From) {
		panic("must use 'VerifySig' with a chain ID to get the signer")
	}

	signer := sdk.AccAddress(ethcmn.HexToAddress(msg.From).Bytes())
	return []sdk.AccAddress{signer}
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
		new(big.Int).SetBytes(msg.Data.ChainID),
		msg.Data.Nonce,
		new(big.Int).SetBytes(msg.Data.GasPrice),
		msg.Data.GasLimit,
		msg.To(),
		new(big.Int).SetBytes(msg.Data.Amount),
		new(big.Int).SetBytes(msg.Data.Input),
		msg.Data.Accesses.ToEthAccessList(),
	})
}

// RLPSignHomesteadBytes returns the RLP hash of an Ethereum transaction message with a
// a Homestead layout without chainID.
func (msg MsgEthereumTx) RLPSignHomesteadBytes() ethcmn.Hash {
	return rlpHash([]interface{}{
		msg.Data.Nonce,
		msg.Data.GasPrice,
		msg.Data.GasLimit,
		msg.To(),
		msg.Data.Amount,
		msg.Data.Input,
	})
}

// EncodeRLP implements the rlp.Encoder interface.
func (msg *MsgEthereumTx) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &msg.Data)
}

// DecodeRLP implements the rlp.Decoder interface.
func (msg *MsgEthereumTx) DecodeRLP(s *rlp.Stream) error {
	_, size, err := s.Kind()
	if err != nil {
		// return error if stream is too large
		return err
	}

	if err := s.Decode(&msg.Data); err != nil {
		return err
	}

	msg.Size_ = float64(ethcmn.StorageSize(rlp.ListSize(size)))
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

// GetGas implements the GasTx interface. It returns the GasLimit of the transaction.
func (msg MsgEthereumTx) GetGas() uint64 {
	return msg.Data.GasLimit
}

// Fee returns gasprice * gaslimit.
func (msg MsgEthereumTx) Fee() *big.Int {
	gasPrice := new(big.Int).SetBytes(msg.Data.GasPrice)
	gasLimit := new(big.Int).SetUint64(msg.Data.GasLimit)
	return new(big.Int).Mul(gasPrice, gasLimit)
}

// ChainID returns which chain id this transaction was signed for (if at all)
func (msg *MsgEthereumTx) ChainID() *big.Int {
	return deriveChainID(new(big.Int).SetBytes(msg.Data.V))
}

// Cost returns amount + gasprice * gaslimit.
func (msg MsgEthereumTx) Cost() *big.Int {
	total := msg.Fee()
	total.Add(total, new(big.Int).SetBytes(msg.Data.Amount))
	return total
}

// RawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (msg MsgEthereumTx) RawSignatureValues() (v, r, s *big.Int) {
	return new(big.Int).SetBytes(msg.Data.V),
		new(big.Int).SetBytes(msg.Data.R),
		new(big.Int).SetBytes(msg.Data.S)
}

// GetFrom loads the ethereum sender address from the sigcache and returns an
// sdk.AccAddress from its bytes
func (msg *MsgEthereumTx) GetFrom() sdk.AccAddress {
	if msg.From == "" {
		return nil
	}

	return ethcmn.HexToAddress(msg.From).Bytes()
}

func (msg MsgEthereumTx) AsTransaction() *ethtypes.Transaction {
	return ethtypes.NewTx(msg.Data.AsEthereumData())
}

// AsEthereumData returns an AccessListTx transaction data from the proto-formatted
// TxData defined on the Cosmos EVM.
func (data TxData) AsEthereumData() ethtypes.TxData {
	var to *ethcmn.Address
	if data.To != nil {
		toAddr := ethcmn.BytesToAddress(data.To)
		to = &toAddr
	}

	return &ethtypes.AccessListTx{
		ChainID:    new(big.Int).SetBytes(data.ChainID),
		Nonce:      data.Nonce,
		GasPrice:   new(big.Int).SetBytes(data.GasPrice),
		Gas:        data.GasLimit,
		To:         to,
		Value:      new(big.Int).SetBytes(data.Amount),
		Data:       data.Input,
		AccessList: *data.Accesses.ToEthAccessList(),
		V:          new(big.Int).SetBytes(data.V),
		R:          new(big.Int).SetBytes(data.R),
		S:          new(big.Int).SetBytes(data.S),
	}
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

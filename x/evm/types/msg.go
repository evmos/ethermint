package types

import (
	"fmt"
	"io"
	"math/big"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	_ sdk.Msg = &MsgEthereumTx{}
	_ sdk.Tx  = &MsgEthereumTx{}
)

// message type and route constants
const (
	// TypeMsgEthereumTx defines the type string of an Ethereum transaction
	TypeMsgEthereumTx = "ethereum"
)

// NewMsgEthereumTx returns a reference to a new Ethereum transaction message.
func NewMsgEthereumTx(
	chainID *big.Int, nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, input []byte, accesses *ethtypes.AccessList,
) *MsgEthereumTx {
	return newMsgEthereumTx(chainID, nonce, to, amount, gasLimit, gasPrice, input, accesses)
}

// NewMsgEthereumTxContract returns a reference to a new Ethereum transaction
// message designated for contract creation.
func NewMsgEthereumTxContract(
	chainID *big.Int, nonce uint64, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, input []byte, accesses *ethtypes.AccessList,
) *MsgEthereumTx {
	return newMsgEthereumTx(chainID, nonce, nil, amount, gasLimit, gasPrice, input, accesses)
}

func newMsgEthereumTx(
	chainID *big.Int, nonce uint64, to *ethcmn.Address, amount *big.Int,
	gasLimit uint64, gasPrice *big.Int, input []byte, accesses *ethtypes.AccessList,
) *MsgEthereumTx {
	if len(input) > 0 {
		input = ethcmn.CopyBytes(input)
	}

	var toHex string
	if to != nil {
		toHex = to.Hex()
	}

	var chainIDBz []byte
	if chainID != nil {
		chainIDBz = chainID.Bytes()
	}

	txData := &TxData{
		ChainID:  chainIDBz,
		Nonce:    nonce,
		To:       toHex,
		Input:    input,
		GasLimit: gasLimit,
		Amount:   []byte{},
		GasPrice: []byte{},
		Accesses: NewAccessList(accesses),
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
	if gasPrice.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidValue, "gas price cannot be negative %s", gasPrice)
	}

	// Amount can be 0
	amount := new(big.Int).SetBytes(msg.Data.Amount)
	if amount.Sign() == -1 {
		return sdkerrors.Wrapf(ErrInvalidValue, "amount cannot be negative %s", amount)
	}

	if msg.Data.To != "" {
		if err := types.ValidateAddress(msg.Data.To); err != nil {
			return sdkerrors.Wrap(err, "invalid to address")
		}
	}

	if msg.From != "" {
		if err := types.ValidateAddress(msg.From); err != nil {
			return sdkerrors.Wrap(err, "invalid from address")
		}
	}

	return nil
}

// To returns the recipient address of the transaction. It returns nil if the
// transaction is a contract creation.
func (msg MsgEthereumTx) To() *ethcmn.Address {
	if msg.Data.To == "" {
		return nil
	}

	recipient := ethcmn.HexToAddress(msg.Data.To)
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
	if msg.Data.ChainID != nil {
		chainID = new(big.Int).SetBytes(msg.Data.ChainID)
	}

	var accessList *ethtypes.AccessList
	if msg.Data.Accesses != nil {
		accessList = msg.Data.Accesses.ToEthAccessList()
	}

	return rlpHash([]interface{}{
		chainID,
		msg.Data.Nonce,
		new(big.Int).SetBytes(msg.Data.GasPrice),
		msg.Data.GasLimit,
		msg.To(),
		new(big.Int).SetBytes(msg.Data.Amount),
		new(big.Int).SetBytes(msg.Data.Input),
		accessList,
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
// takes a keyring signer and the chainID to sign an Ethereum transaction according to
// EIP155 standard.
// This method mutates the transaction as it populates the V, R, S
// fields of the Transaction's Signature.
// The function will fail if the sender address is not defined for the msg or if
// the sender is not registered on the keyring
func (msg *MsgEthereumTx) Sign(chainID *big.Int, signer keyring.Signer) error {
	from := msg.GetFrom()
	if from == nil {
		return fmt.Errorf("sender address not defined for message")
	}

	txHash := msg.RLPSignBytes(chainID)

	sig, _, err := signer.SignByAddress(from, txHash[:])
	if err != nil {
		return err
	}

	if len(sig) != crypto.SignatureLength {
		return fmt.Errorf(
			"wrong size for signature: got %d, want %d",
			len(sig),
			crypto.SignatureLength,
		)
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
	return new(big.Int).SetBytes(msg.Data.ChainID)
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

// AsTransaction creates an Ethereum Transaction type from the msg fields
func (msg MsgEthereumTx) AsTransaction() *ethtypes.Transaction {
	return ethtypes.NewTx(msg.Data.AsEthereumData())
}

// AsMessage creates an Ethereum core.Message from the msg fields. This method
// fails if the sender address is not defined
func (msg MsgEthereumTx) AsMessage() (core.Message, error) {
	if msg.From == "" {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "'from' address cannot be empty")
	}

	from := ethcmn.HexToAddress(msg.From)

	var to *ethcmn.Address
	if msg.Data.To != "" {
		toAddr := ethcmn.HexToAddress(msg.Data.To)
		to = &toAddr
	}

	var accessList ethtypes.AccessList
	if msg.Data.Accesses != nil {
		accessList = *msg.Data.Accesses.ToEthAccessList()
	}

	return ethtypes.NewMessage(
		from,
		to,
		msg.Data.Nonce,
		new(big.Int).SetBytes(msg.Data.Amount),
		msg.Data.GasLimit,
		new(big.Int).SetBytes(msg.Data.GasPrice),
		msg.Data.Input,
		accessList,
		true,
	), nil
}

// AsEthereumData returns an AccessListTx transaction data from the proto-formatted
// TxData defined on the Cosmos EVM.
func (data TxData) AsEthereumData() ethtypes.TxData {
	var to *ethcmn.Address
	if data.To != "" {
		toAddr := ethcmn.HexToAddress(data.To)
		to = &toAddr
	}

	var accessList ethtypes.AccessList
	if data.Accesses != nil {
		accessList = *data.Accesses.ToEthAccessList()
	}

	return &ethtypes.AccessListTx{
		ChainID:    new(big.Int).SetBytes(data.ChainID),
		Nonce:      data.Nonce,
		GasPrice:   new(big.Int).SetBytes(data.GasPrice),
		Gas:        data.GasLimit,
		To:         to,
		Value:      new(big.Int).SetBytes(data.Amount),
		Data:       data.Input,
		AccessList: accessList,
		V:          new(big.Int).SetBytes(data.V),
		R:          new(big.Int).SetBytes(data.R),
		S:          new(big.Int).SetBytes(data.S),
	}
}

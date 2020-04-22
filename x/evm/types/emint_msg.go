package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
)

var (
	_ sdk.Msg = MsgEthermint{}
)

const (
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
	if msg.Price.Sign() != 1 {
		return sdkerrors.Wrapf(types.ErrInvalidValue, "price must be positive %s", msg.Price)
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

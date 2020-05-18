package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgFund funds a recipient address
type MsgFund struct {
	Amount    sdk.Coins      `json:"amount" yaml:"amount"`
	Sender    sdk.AccAddress `json:"sender" yaml:"sender"`
	Recipient sdk.AccAddress `json:"receipient" yaml:"receipient"`
}

// NewMsgFund is a constructor function for NewMsgFund
func NewMsgFund(amount sdk.Coins, sender, recipient sdk.AccAddress) MsgFund {
	return MsgFund{
		Amount:    amount,
		Sender:    sender,
		Recipient: recipient,
	}
}

// Route should return the name of the module
func (msg MsgFund) Route() string { return RouterKey }

// Type should return the action
func (msg MsgFund) Type() string { return "fund" }

// ValidateBasic runs stateless checks on the message
func (msg MsgFund) ValidateBasic() error {
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}
	if msg.Sender.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "sender %s", msg.Sender.String())
	}
	if msg.Recipient.Empty() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "recipient %s", msg.Recipient.String())
	}
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgFund) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgFund) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRemoveAdmin{}

func NewMsgRemoveAdmin(fromAddress sdk.AccAddress, address string) *MsgRemoveAdmin {
	return &MsgRemoveAdmin{
		Sender:       fromAddress.String(),
		AdminAddress: address,
	}
}

func (msg *MsgRemoveAdmin) Route() string {
	return RouterKey
}

func (msg *MsgRemoveAdmin) Type() string {
	return "RemoveAdmin"
}

func (msg MsgRemoveAdmin) GetSigners() []sdk.AccAddress {
	if len(msg.Sender) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Sender)}
}

func (msg *MsgRemoveAdmin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveAdmin) ValidateBasic() error {
	if len(msg.AdminAddress) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Admin address can't be empty")
	}

	if len(msg.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Sender address can't be empty")
	}

	_, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Wrong admin address")
	}

	return nil
}

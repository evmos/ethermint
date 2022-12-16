package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddAdmin{}

func NewMsgAddAdmin(sender sdk.AccAddress, address string, editOption bool) *MsgAddAdmin {
	fmt.Println(sender.String())
	fmt.Println(address)
	fmt.Println(editOption)
	return &MsgAddAdmin{
		Sender:       sender.String(),
		AdminAddress: address,
		EditOption:   editOption,
	}
}

func (msg MsgAddAdmin) Route() string {
	return RouterKey
}

func (msg MsgAddAdmin) Type() string {
	return "AddAdmin"
}

func (msg MsgAddAdmin) GetSigners() []sdk.AccAddress {
	if len(msg.Sender) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Sender)}
}

func (msg MsgAddAdmin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAddAdmin) ValidateBasic() error {
	if len(msg.AdminAddress) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Admin address can't be empty")
	}

	if len(msg.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Sender address can't be empty")
	}

	// keeper.Keeper.GetDistributor()

	_, err := sdk.AccAddressFromBech32(msg.AdminAddress)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Wrong admin address")
	}

	return nil
}

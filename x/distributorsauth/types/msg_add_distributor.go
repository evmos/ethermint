package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgAddDistributor{}

func NewMsgAddDistributor(sender sdk.AccAddress, address string, endDate uint64) *MsgAddDistributor {
	return &MsgAddDistributor{
		Sender:             sender.String(),
		DistributorAddress: address,
		EndDate:            endDate,
	}
}

func (msg MsgAddDistributor) Route() string {
	return RouterKey
}

func (msg MsgAddDistributor) Type() string {
	return "AddDistributor"
}

func (msg MsgAddDistributor) GetSigners() []sdk.AccAddress {
	if len(msg.Sender) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Sender)}
}

func (msg MsgAddDistributor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAddDistributor) ValidateBasic() error {
	if len(msg.DistributorAddress) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Distributor address can't be empty")
	}

	if len(msg.Sender) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Sender address can't be empty")
	}

	_, err := sdk.AccAddressFromBech32(msg.DistributorAddress)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Wrong distributor address")
	}

	return nil
}

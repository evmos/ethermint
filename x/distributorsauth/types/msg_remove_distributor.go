package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgRemoveDistributor{}

func NewMsgRemoveDistributor(fromAddress sdk.AccAddress, address string) *MsgRemoveDistributor {
	return &MsgRemoveDistributor{
		Sender:             fromAddress.String(),
		DistributorAddress: address,
	}
}

func (msg *MsgRemoveDistributor) Route() string {
	return RouterKey
}

func (msg *MsgRemoveDistributor) Type() string {
	return "RemoveDistributor"
}

func (msg MsgRemoveDistributor) GetSigners() []sdk.AccAddress {
	if len(msg.Sender) == 0 {
		return []sdk.AccAddress{}
	}

	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Sender)}
}

func (msg *MsgRemoveDistributor) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveDistributor) ValidateBasic() error {
	if msg.DistributorAddress == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Distributor address can't be empty")
	}

	if msg.Sender == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Sender address can't be empty")
	}

	_, err := sdk.AccAddressFromBech32(msg.DistributorAddress)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "Wrong distributor address")
	}

	return nil
}

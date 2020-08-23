package faucet

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/x/faucet/types"
)

// NewHandler returns a handler for faucet messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgFund:
			return handleMsgFund(ctx, keeper, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// handleMsgFund handles a message to fund an address
func handleMsgFund(ctx sdk.Context, keeper Keeper, msg types.MsgFund) (*sdk.Result, error) {
	err := keeper.Fund(ctx, msg.Amount, msg.Recipient)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeRecipient, msg.Recipient.String()),
		),
	)

	return &sdk.Result{
		Events: ctx.EventManager().Events(),
	}, nil
}

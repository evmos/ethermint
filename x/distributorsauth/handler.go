package distributorsauth

// import (
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
// 	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/keeper"
//   )

// func handleMsgAddDistributor(ctx sdk.Context, k keeper.Keeper, msg types.MsgAddDistributor) (*sdk.Result, error) {
// 	var post = types.Post{
// 		ID: 	 		msg.ID,
// 		FromAddress:	msg.FromAddress,
// 		Address:   		msg.Address,
// 		EndDate:   		msg.EndDate,
// 	}
// 	k.CreatePost(ctx, post)

// 	ID          string
// 	FromAddress sdk.AccAddress `json:"from" yaml:"from"` //`protobuf:"bytes,1,opt,name=from_address,json=fromAddress,proto3" json:"from_address,omitempty"`
// 	Address     sdk.AccAddress `json:"address" yaml:"address"`
// 	EndDate     *big.Int

// 	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
// }

import (
	"fmt"

	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/keeper"
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewHandler ...
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		// this line is used by starport scaffolding
		case *types.MsgAddDistributor:
			return handleMsgAddDistributor(ctx, k, msg)
		case *types.MsgRemoveDistributor:
			return handleMsgRemoveDistributor(ctx, k, msg)
		case *types.MsgAddAdmin:
			return handleMsgAddAdmin(ctx, k, msg)
		case *types.MsgRemoveAdmin:
			return handleMsgRemoveAdmin(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}

func handleMsgAddDistributor(ctx sdk.Context, k keeper.Keeper, msg *types.MsgAddDistributor) (*sdk.Result, error) {
	var DistributorInfo = types.DistributorInfo{
		Address: msg.DistributorAddress,
		EndDate: msg.EndDate, //msg.DistributorAddress,
	}

	// ctx.BlockTime()

	k.AddDistributor(ctx, DistributorInfo)

	return &sdk.Result{}, nil //&sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRemoveDistributor(ctx sdk.Context, k keeper.Keeper, msg *types.MsgRemoveDistributor) (*sdk.Result, error) {
	k.RemoveDistributor(ctx, msg.DistributorAddress)

	// return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	return &sdk.Result{}, nil
}

func handleMsgAddAdmin(ctx sdk.Context, k keeper.Keeper, msg *types.MsgAddAdmin) (*sdk.Result, error) {
	var Admin = types.Admin{
		Address:    msg.AdminAddress,
		EditOption: msg.EditOption,
	}

	// ctx.BlockTime()

	k.AddAdmin(ctx, Admin)

	return &sdk.Result{}, nil //&sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgRemoveAdmin(ctx sdk.Context, k keeper.Keeper, msg *types.MsgRemoveAdmin) (*sdk.Result, error) {
	k.RemoveAdmin(ctx, msg.AdminAddress)

	// return &sdk.Result{Events: ctx.EventManager().Events()}, nil
	return &sdk.Result{}, nil
}

func NewDistributorProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddDistributorProposal:
			return k.HandleAddDistributorProposal(ctx, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized distr proposal content type: %T", c)
		}
	}
}

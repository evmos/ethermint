package keeper

import (
	"context"

	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (s msgServer) AddDistributor(goCtx context.Context, msg *types.MsgAddDistributor) (*types.MsgAddDistributorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := s.checkSenderHaveAdminsWrights(ctx, msg.Sender, false)
	if err != nil {
		return nil, err
	}

	var DistributorInfo = types.DistributorInfo{
		Address: msg.DistributorAddress,
		EndDate: msg.EndDate,
	}

	ctx.BlockTime()

	s.Keeper.AddDistributor(ctx, DistributorInfo)

	return &types.MsgAddDistributorResponse{}, nil
}

func (s msgServer) RemoveDistributor(goCtx context.Context, msg *types.MsgRemoveDistributor) (*types.MsgRemoveDistributorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := s.checkSenderHaveAdminsWrights(ctx, msg.Sender, false)
	if err != nil {
		return nil, err
	}

	s.Keeper.RemoveDistributor(ctx, msg.DistributorAddress)

	return &types.MsgRemoveDistributorResponse{}, nil
}

func (s msgServer) AddAdmin(goCtx context.Context, msg *types.MsgAddAdmin) (*types.MsgAddAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := s.checkSenderHaveAdminsWrights(ctx, msg.Sender, true)
	if err != nil {
		return nil, err
	}

	var AdminInfo = types.Admin{
		Address:    msg.AdminAddress,
		EditOption: msg.EditOption,
	}

	s.Keeper.AddAdmin(ctx, AdminInfo)

	return &types.MsgAddAdminResponse{}, nil
}

func (s msgServer) RemoveAdmin(goCtx context.Context, msg *types.MsgRemoveAdmin) (*types.MsgRemoveAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := s.checkSenderHaveAdminsWrights(ctx, msg.Sender, true)
	if err != nil {
		return nil, err
	}

	s.Keeper.RemoveAdmin(ctx, msg.AdminAddress)

	return &types.MsgRemoveAdminResponse{}, nil
}

func checkSenderIsGovModule(address string) bool {
	return (address == authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func (s msgServer) checkSenderHaveAdminsWrights(ctx sdk.Context, address string, editable bool) error {
	if checkSenderIsGovModule(address) {
		return nil
	}

	admin, err := s.Keeper.GetAdmin(ctx, address)
	if err != nil {
		return sdkerrors.Wrap(types.ErrSenderIsNotAnAdmin, address)
	}

	if editable && !admin.EditOption {
		return sdkerrors.Wrap(types.ErrWrongAdminEditOption, address)
	}

	return nil
}

var _ types.MsgServer = msgServer{}

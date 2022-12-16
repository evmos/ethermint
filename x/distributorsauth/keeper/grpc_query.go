package keeper

import (
	"context"

	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// // NewQuerier returns an implementation of the distributorsauth QueryServer interface
// // for the provided Keeper.
// func NewQuerier(keeper Keeper) types.QueryServer {
// 	return &querier{Keeper: keeper}
// }

// Distributor queries distributor with address.
func (k Keeper) Distributor(
	goCtx context.Context,
	req *types.QueryDistributorRequest,
) (*types.QueryDistributorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// distrAddr, err := sdk.AccAddressFromBech32(req.DistributorAddr)
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	ctx := sdk.UnwrapSDKContext(goCtx)

	distributorInfo, err := k.GetDistributor(ctx, req.DistributorAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryDistributorResponse{
		Distributor: distributorInfo,
	}, nil
}

// Distributors queries request of all distributors.
func (k Keeper) Distributors(
	goCtx context.Context,
	req *types.QueryDistributorsRequest,
) (*types.QueryDistributorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// var distributors []types.DistributorInfo
	distributors, err := k.GetDistributors(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cant get admins")
	}

	return &types.QueryDistributorsResponse{
		Distributors: distributors,
	}, nil
}

// Admin queries admin with address.
func (k Keeper) Admin(
	goCtx context.Context,
	req *types.QueryAdminRequest,
) (*types.QueryAdminResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	adminInfo, err := k.GetAdmin(ctx, req.AdminAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryAdminResponse{
		Admin: adminInfo,
	}, nil
}

// Admins queries request of all admins.
func (k Keeper) Admins(
	goCtx context.Context,
	req *types.QueryAdminsRequest,
) (*types.QueryAdminsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// var distributors []types.DistributorInfo
	admins, err := k.GetAdmins(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "cant get admin")
	}

	return &types.QueryAdminsResponse{
		Admins: admins,
	}, nil
}

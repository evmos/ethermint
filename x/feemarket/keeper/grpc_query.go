package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tharsis/ethermint/x/feemarket/types"
)

var _ types.QueryServer = Keeper{}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// BaseFee implements the Query/BaseFee gRPC method
// return empty if base fee is not enable
func (k Keeper) BaseFee(c context.Context, _ *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	res := &types.QueryBaseFeeResponse{}

	params := k.GetParams(ctx)

	if params.NoBaseFee {
		return res, nil
	}

	baseFee := k.GetBaseFee(ctx)

	if baseFee != nil {
		aux := sdk.NewIntFromBigInt(baseFee)
		res.BaseFee = &aux
	}

	return res, nil
}

// BlockGas implements the Query/BlockGas gRPC method
func (k Keeper) BlockGas(c context.Context, _ *types.QueryBlockGasRequest) (*types.QueryBlockGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	gas := k.GetBlockGasUsed(ctx)

	return &types.QueryBlockGasResponse{
		Gas: int64(gas),
	}, nil
}

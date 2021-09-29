package keeper

import (
	"context"
	"math/big"

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
func (k Keeper) BaseFee(c context.Context, _ *types.QueryBaseFeeRequest) (*types.QueryBaseFeeResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	baseFee := k.GetBaseFee(ctx)
	if baseFee == nil {
		// TODO: should this be 0? 1? error?
		baseFee = big.NewInt(0)
	}

	return &types.QueryBaseFeeResponse{
		BaseFee: sdk.NewIntFromBigInt(baseFee),
	}, nil
}

// BlockGas implements the Query/BlockGas gRPC method
func (k Keeper) BlockGas(c context.Context, _ *types.QueryBlockGasRequest) (*types.QueryBlockGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	gas := k.GetBlockGasUsed(ctx)

	return &types.QueryBlockGasResponse{
		Gas: int64(gas),
	}, nil
}

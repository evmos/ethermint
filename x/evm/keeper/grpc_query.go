package keeper

import (
	"context"
	"math/big"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	ethcmn "github.com/ethereum/go-ethereum/common"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.QueryServer = Keeper{}

// Account implements the Query/Account gRPC method
func (k Keeper) Account(c context.Context, req *types.QueryAccountRequest) (*types.QueryAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := ethermint.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	so := k.CommitStateDB.GetOrNewStateObject(ethcmn.HexToAddress(req.Address))
	balance, err := ethermint.MarshalBigInt(so.Balance())
	if err != nil {
		return nil, err
	}

	return &types.QueryAccountResponse{
		Balance:  balance,
		CodeHash: so.CodeHash(),
		Nonce:    so.Nonce(),
	}, nil
}

func (k Keeper) CosmosAccount(c context.Context, req *types.QueryCosmosAccountRequest) (*types.QueryCosmosAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := ethermint.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	ethAddr := ethcmn.HexToAddress(req.Address)
	cosmosAddr := sdk.AccAddress(ethAddr.Bytes())

	account := k.accountKeeper.GetAccount(ctx, cosmosAddr)
	res := types.QueryCosmosAccountResponse{
		CosmosAddress: cosmosAddr.String(),
	}

	if account != nil {
		res.Sequence = account.GetSequence()
		res.AccountNumber = account.GetAccountNumber()
	}

	return &res, nil
}

// Balance implements the Query/Balance gRPC method
func (k Keeper) Balance(c context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := ethermint.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	balanceInt := k.CommitStateDB.GetBalance(ethcmn.HexToAddress(req.Address))
	balance, err := ethermint.MarshalBigInt(balanceInt)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			"failed to marshal big.Int to string",
		)
	}

	return &types.QueryBalanceResponse{
		Balance: balance,
	}, nil
}

// Storage implements the Query/Storage gRPC method
func (k Keeper) Storage(c context.Context, req *types.QueryStorageRequest) (*types.QueryStorageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := ethermint.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	if ethermint.IsEmptyHash(req.Key) {
		return nil, status.Errorf(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	address := ethcmn.HexToAddress(req.Address)
	key := ethcmn.HexToHash(req.Key)

	state := k.CommitStateDB.GetState(address, key)
	stateHex := state.Hex()

	if ethermint.IsEmptyHash(stateHex) {
		return nil, status.Error(
			codes.NotFound, "contract code not found for given address",
		)
	}

	return &types.QueryStorageResponse{
		Value: stateHex,
	}, nil
}

// Code implements the Query/Code gRPC method
func (k Keeper) Code(c context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := ethermint.ValidateAddress(req.Address); err != nil {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrZeroAddress.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	address := ethcmn.HexToAddress(req.Address)
	code := k.CommitStateDB.GetCode(address)

	return &types.QueryCodeResponse{
		Code: code,
	}, nil
}

// TxLogs implements the Query/TxLogs gRPC method
func (k Keeper) TxLogs(c context.Context, req *types.QueryTxLogsRequest) (*types.QueryTxLogsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if ethermint.IsEmptyHash(req.Hash) {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	hash := ethcmn.HexToHash(req.Hash)
	logs, err := k.CommitStateDB.GetLogs(hash)
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			err.Error(),
		)
	}

	return &types.QueryTxLogsResponse{
		Logs: types.NewTransactionLogsFromEth(hash, logs).Logs,
	}, nil
}

// BlockLogs implements the Query/BlockLogs gRPC method
func (k Keeper) BlockLogs(c context.Context, req *types.QueryBlockLogsRequest) (*types.QueryBlockLogsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if ethermint.IsEmptyHash(req.Hash) {
		return nil, status.Error(
			codes.InvalidArgument,
			types.ErrEmptyHash.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixLogs)
	txLogs := []types.TransactionLogs{}

	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, accumulate bool) (bool, error) {
		var txLog types.TransactionLogs
		k.cdc.MustUnmarshalBinaryBare(value, &txLog)

		if len(txLog.Logs) > 0 && txLog.Logs[0].BlockHash == req.Hash {
			if accumulate {
				txLogs = append(txLogs, txLog)
			}
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBlockLogsResponse{
		TxLogs:     txLogs,
		Pagination: pageRes,
	}, nil
}

// BlockBloom implements the Query/BlockBloom gRPC method
func (k Keeper) BlockBloom(c context.Context, _ *types.QueryBlockBloomRequest) (*types.QueryBlockBloomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	bloom, found := k.GetBlockBloom(ctx, ctx.BlockHeight())
	if !found {
		return nil, status.Error(
			codes.NotFound, types.ErrBloomNotFound.Error(),
		)
	}

	return &types.QueryBlockBloomResponse{
		Bloom: bloom.Bytes(),
	}, nil
}

// Params implements the Query/Params gRPC method
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// StaticCall implements Query/StaticCall gRPCP method
func (k Keeper) StaticCall(c context.Context, req *types.QueryStaticCallRequest) (*types.QueryStaticCallResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.CommitStateDB.WithContext(ctx)

	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := ethcmn.BytesToHash(txHash)

	var recipient *ethcmn.Address
	if len(req.Address) > 0 {
		addr := ethcmn.HexToAddress(req.Address)
		recipient = &addr
	}

	so := k.CommitStateDB.GetOrNewStateObject(*recipient)
	sender := ethcmn.HexToAddress("0xaDd00275E3d9d213654Ce5223f0FADE8b106b707")

	msg := types.NewMsgEthereumTx(
		chainIDEpoch, so.Nonce(), recipient, big.NewInt(0), 100000000, big.NewInt(0), req.Input, nil,
	)
	msg.From = sender.Hex()

	if err := msg.ValidateBasic(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ethMsg, err := msg.AsMessage()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	st := &types.StateTransition{
		Message:  ethMsg,
		Csdb:     k.CommitStateDB.WithContext(ctx),
		ChainID:  chainIDEpoch,
		TxHash:   &ethHash,
		Simulate: ctx.IsCheckTx(),
		Debug:    false,
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, status.Error(codes.Internal, types.ErrChainConfigNotFound.Error())
	}

	ret, err := st.StaticCall(ctx, config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStaticCallResponse{Data: ret}, nil
}

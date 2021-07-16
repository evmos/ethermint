package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/palantir/stacktrace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"
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

	addr := ethcmn.HexToAddress(req.Address)

	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	return &types.QueryAccountResponse{
		Balance:  k.GetBalance(addr).String(),
		CodeHash: k.GetCodeHash(addr).Hex(),
		Nonce:    k.GetNonce(addr),
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
	k.WithContext(ctx)

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

func (k Keeper) ValidatorAccount(c context.Context, req *types.QueryValidatorAccountRequest) (*types.QueryValidatorAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	consAddr, err := sdk.ConsAddressFromBech32(req.ConsAddress)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	if !found {
		return nil, nil
	}

	accAddr := sdk.AccAddress(validator.GetOperator())

	res := types.QueryValidatorAccountResponse{
		AccountAddress: accAddr.String(),
	}

	account := k.accountKeeper.GetAccount(ctx, accAddr)
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
	k.WithContext(ctx)

	balanceInt := k.GetBalance(ethcmn.HexToAddress(req.Address))

	return &types.QueryBalanceResponse{
		Balance: balanceInt.String(),
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
	k.WithContext(ctx)

	address := ethcmn.HexToAddress(req.Address)
	key := ethcmn.HexToHash(req.Key)

	state := k.GetState(address, key)
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
	k.WithContext(ctx)

	address := ethcmn.HexToAddress(req.Address)
	code := k.GetCode(address)

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
	k.WithContext(ctx)

	hash := ethcmn.HexToHash(req.Hash)
	logs := k.GetTxLogs(hash)

	return &types.QueryTxLogsResponse{
		Logs: types.NewLogsFromEth(logs),
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
		k.cdc.MustUnmarshal(value, &txLog)

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
		// if the bloom is not found, query the transient store at the current height
		k.ctx = ctx
		bloomInt := k.GetBlockBloomTransient()

		if bloomInt.Sign() == 0 {
			return nil, status.Error(
				codes.NotFound, sdkerrors.Wrapf(types.ErrBloomNotFound, "height: %d", ctx.BlockHeight()).Error(),
			)
		}

		bloom = ethtypes.BytesToBloom(bloomInt.Bytes())
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

	// ctx := sdk.UnwrapSDKContext(c)
	// k.WithContext(ctx)

	// // parse the chainID from a string to a base-10 integer
	// chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	// ethHash := ethcmn.BytesToHash(txHash)

	// var recipient *ethcmn.Address
	// if len(req.Address) > 0 {
	// 	addr := ethcmn.HexToAddress(req.Address)
	// 	recipient = &addr
	// }

	// so := k.GetOrNewStateObject(*recipient)
	// sender := ethcmn.HexToAddress("0xaDd00275E3d9d213654Ce5223f0FADE8b106b707")

	// msg := types.NewTx(
	// 	chainIDEpoch, so.Nonce(), recipient, big.NewInt(0), 100000000, big.NewInt(0), req.Input, nil,
	// )
	// msg.From = sender.Hex()

	// if err := msg.ValidateBasic(); err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// ethMsg, err := msg.AsMessage()
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// st := &types.StateTransition{
	// 	Message:  ethMsg,
	// 	Csdb:     k.WithContext(ctx),
	// 	ChainID:  chainIDEpoch,
	// 	TxHash:   &ethHash,
	// 	Simulate: ctx.IsCheckTx(),
	// 	Debug:    false,
	// }

	// config, found := k.GetChainConfig(ctx)
	// if !found {
	// 	return nil, status.Error(codes.Internal, types.ErrChainConfigNotFound.Error())
	// }

	// ret, err := st.StaticCall(ctx, config)
	// if err != nil {
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// return &types.QueryStaticCallResponse{Data: ret}, nil

	return nil, nil
}

// EthCall implements eth_call rpc api.
func (k Keeper) EthCall(c context.Context, req *types.EthCallRequest) (*types.MsgEthereumTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	var args types.CallArgs
	err := json.Unmarshal(req.Args, &args)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	msg := args.ToMessage(req.GasCap)

	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)

	coinbase, err := k.GetCoinbaseAddress()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	evm := k.NewEVM(msg, ethCfg, params, coinbase)
	// pass true means execute in query mode, which don't do actual gas refund.
	res, err := k.ApplyMessage(evm, msg, ethCfg, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

// EstimateGas implements eth_estimateGas rpc api.
func (k Keeper) EstimateGas(c context.Context, req *types.EthCallRequest) (*types.EstimateGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	var args types.CallArgs
	err := json.Unmarshal(req.Args, &args)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = ethparams.TxGas - 1
		hi  uint64
		cap uint64
	)

	// Determine the highest gas limit can be used during the estimation.
	if args.Gas != nil && uint64(*args.Gas) >= ethparams.TxGas {
		hi = uint64(*args.Gas)
	} else {
		// Query block gas limit
		params := ctx.ConsensusParams()
		if params != nil && params.Block != nil && params.Block.MaxGas > 0 {
			hi = uint64(params.Block.MaxGas)
		} else {
			hi = req.GasCap
		}
	}

	// TODO Recap the highest gas limit with account's available balance.

	// Recap the highest gas allowance with specified gascap.
	if req.GasCap != 0 && hi > req.GasCap {
		hi = req.GasCap
	}
	cap = hi

	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)

	coinbase, err := k.GetCoinbaseAddress()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *types.MsgEthereumTxResponse, error) {
		args.Gas = (*hexutil.Uint64)(&gas)

		// Execute the call in an isolated context
		sandboxCtx, _ := ctx.CacheContext()
		k.WithContext(sandboxCtx)

		msg := args.ToMessage(req.GasCap)
		evm := k.NewEVM(msg, ethCfg, params, coinbase)
		// pass true means execute in query mode, which don't do actual gas refund.
		rsp, err := k.ApplyMessage(evm, msg, ethCfg, true)

		k.WithContext(ctx)

		if err != nil {
			if errors.Is(stacktrace.RootCause(err), core.ErrIntrinsicGas) {
				return true, nil, nil // Special case, raise gas limit
			}
			return true, nil, err // Bail out
		}
		return len(rsp.VmError) > 0, rsp, nil
	}

	// Execute the binary search and hone in on an executable gas limit
	hi, err = types.BinSearch(lo, hi, executable)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		failed, result, err := executable(hi)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if failed {
			if result != nil && result.VmError != vm.ErrOutOfGas.Error() {
				if result.VmError == vm.ErrExecutionReverted.Error() {
					return nil, status.Error(codes.Internal, types.NewExecErrorWithReason(result.Ret).Error())
				}
				return nil, status.Error(codes.Internal, result.VmError)
			}
			// Otherwise, the specified gas cap is too low
			return nil, status.Error(codes.Internal, fmt.Sprintf("gas required exceeds allowance (%d)", cap))
		}
	}
	return &types.EstimateGasResponse{Gas: hi}, nil
}

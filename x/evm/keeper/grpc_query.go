package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/palantir/stacktrace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

var _ types.QueryServer = Keeper{}

const (
	defaultTraceTimeout = 5 * time.Second
)

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

	addr := common.HexToAddress(req.Address)

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

	ethAddr := common.HexToAddress(req.Address)
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

	balanceInt := k.GetBalance(common.HexToAddress(req.Address))

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

	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	address := common.HexToAddress(req.Address)
	key := common.HexToHash(req.Key)

	state := k.GetState(address, key)
	stateHex := state.Hex()

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

	address := common.HexToAddress(req.Address)
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

	hash := common.HexToHash(req.Hash)

	store := prefix.NewStore(k.Ctx().KVStore(k.storeKey), append(types.KeyPrefixLogs, hash.Bytes()...))

	var logs []*types.Log

	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		var log types.Log
		if err := k.cdc.Unmarshal(value, &log); err != nil {
			return err
		}
		logs = append(logs, &log)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTxLogsResponse{
		Logs:       logs,
		Pagination: pageRes,
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

	mapOrder := []string{}
	logs := make(map[string][]*types.Log)

	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_, value []byte, accumulate bool) (bool, error) {
		var txLog types.Log
		if err := k.cdc.Unmarshal(value, &txLog); err != nil {
			return false, err
		}

		if txLog.BlockHash == req.Hash {
			if accumulate {
				if len(logs[txLog.TxHash]) == 0 {
					mapOrder = append(mapOrder, txLog.TxHash)
				}

				logs[txLog.TxHash] = append(logs[txLog.TxHash], &txLog)
			}
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txsLogs := []types.TransactionLogs{}
	for _, txHash := range mapOrder {
		if len(logs[txHash]) > 0 {
			txsLogs = append(txsLogs, types.TransactionLogs{Hash: txHash, Logs: logs[txHash]})
		}
	}

	return &types.QueryBlockLogsResponse{
		TxLogs:     txsLogs,
		Pagination: pageRes,
	}, nil
}

// BlockBloom implements the Query/BlockBloom gRPC method
func (k Keeper) BlockBloom(c context.Context, req *types.QueryBlockBloomRequest) (*types.QueryBlockBloomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	bloom, found := k.GetBlockBloom(ctx, req.Height)
	if !found {
		// if the bloom is not found, query the transient store at the current height
		k.WithContext(ctx)
		bloomInt := k.GetBlockBloomTransient()

		if bloomInt.Sign() == 0 {
			return nil, status.Error(
				codes.NotFound, sdkerrors.Wrapf(types.ErrBloomNotFound, "height: %d", req.Height).Error(),
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

	coinbase, err := k.GetCoinbaseAddress(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	tracer := types.NewTracer(k.tracer, msg, ethCfg, k.Ctx().BlockHeight(), k.debug)
	evm := k.NewEVM(msg, ethCfg, params, coinbase, tracer)

	// pass true means execute in query mode, which don't do actual gas refund.
	res, err := k.ApplyMessage(evm, msg, ethCfg, true)
	k.ctxStack.RevertAll()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

// EstimateGas implements eth_estimateGas rpc api.
func (k Keeper) EstimateGas(c context.Context, req *types.EthCallRequest) (*types.EstimateGasResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)

	if req.GasCap < ethparams.TxGas {
		return nil, status.Error(codes.InvalidArgument, "gas cap cannot be lower than 21,000")
	}

	var args types.CallArgs
	err := json.Unmarshal(req.Args, &args)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  = ethparams.TxGas - 1
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

	coinbase, err := k.GetCoinbaseAddress(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) (bool, *types.MsgEthereumTxResponse, error) {
		args.Gas = (*hexutil.Uint64)(&gas)

		// Reset to the initial context
		k.WithContext(ctx)

		msg := args.ToMessage(req.GasCap)

		tracer := types.NewTracer(k.tracer, msg, ethCfg, k.Ctx().BlockHeight(), k.debug)
		evm := k.NewEVM(msg, ethCfg, params, coinbase, tracer)
		// pass true means execute in query mode, which don't do actual gas refund.
		rsp, err := k.ApplyMessage(evm, msg, ethCfg, true)

		k.ctxStack.RevertAll()

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
					return nil, types.NewExecErrorWithReason(result.Ret)
				}
				return nil, status.Error(codes.Internal, result.VmError)
			}
			// Otherwise, the specified gas cap is too low
			return nil, status.Error(codes.Internal, fmt.Sprintf("gas required exceeds allowance (%d)", cap))
		}
	}
	return &types.EstimateGasResponse{Gas: hi}, nil
}

// TraceTx configures a new tracer according to the provided configuration, and
// executes the given message in the provided environment. The return value will
// be tracer dependent.
func (k Keeper) TraceTx(c context.Context, req *types.QueryTraceTxRequest) (*types.QueryTraceTxResponse, error) {
	// Assemble the structured logger or the JavaScript tracer
	var (
		tracer     vm.Tracer
		err        error
		resultData []byte
	)

	ctx := sdk.UnwrapSDKContext(c)
	k.WithContext(ctx)
	params := k.GetParams(ctx)

	coinbase, err := k.GetCoinbaseAddress(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ethCfg := params.ChainConfig.EthereumConfig(k.eip155ChainID)
	signer := ethtypes.MakeSigner(ethCfg, big.NewInt(ctx.BlockHeight()))
	coreMessage, err := req.Msg.AsMessage(signer)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch {
	case req.TraceConfig != nil && req.TraceConfig.Tracer != "":
		timeout := defaultTraceTimeout
		// TODO change timeout to time.duration
		// Used string to comply with go ethereum
		if req.TraceConfig.Timeout != "" {
			if timeout, err = time.ParseDuration(req.TraceConfig.Timeout); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "timeout value: %s", err.Error())
			}
		}

		txContext := core.NewEVMTxContext(coreMessage)
		// Constuct the JavaScript tracer to execute with
		if tracer, err = tracers.New(req.TraceConfig.Tracer, txContext); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Handle timeouts and RPC cancellations
		deadlineCtx, cancel := context.WithTimeout(c, timeout)
		go func() {
			<-deadlineCtx.Done()
			if deadlineCtx.Err() == context.DeadlineExceeded {
				tracer.(*tracers.Tracer).Stop(errors.New("execution timeout"))
			}
		}()
		defer cancel()
	case req.TraceConfig != nil && req.TraceConfig.LogConfig != nil:
		logConfig := vm.LogConfig{
			DisableMemory:  req.TraceConfig.LogConfig.DisableMemory,
			Debug:          req.TraceConfig.LogConfig.Debug,
			DisableStorage: req.TraceConfig.LogConfig.DisableStorage,
			DisableStack:   req.TraceConfig.LogConfig.DisableStack,
		}
		tracer = vm.NewStructLogger(&logConfig)
	default:
		tracer = types.NewTracer(types.TracerStruct, coreMessage, ethCfg, ctx.BlockHeight(), true)
	}

	evm := k.NewEVM(coreMessage, ethCfg, params, coinbase, tracer)

	k.SetTxHashTransient(common.HexToHash(req.Msg.Hash))
	k.SetTxIndexTransient(uint64(req.TxIndex))

	res, err := k.ApplyMessage(evm, coreMessage, ethCfg, true)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Depending on the tracer type, format and return the trace result data.
	switch tracer := tracer.(type) {
	case *vm.StructLogger:
		// TODO Return proper returnValue
		result := types.ExecutionResult{
			Gas:         res.GasUsed,
			Failed:      res.Failed(),
			ReturnValue: "",
			StructLogs:  types.FormatLogs(tracer.StructLogs()),
		}

		resultData, err = json.Marshal(result)

		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	case *tracers.Tracer:
		result, err := tracer.GetResult()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		resultData, err = json.Marshal(result)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid tracer type")
	}

	return &types.QueryTraceTxResponse{
		Data: resultData,
	}, nil
}

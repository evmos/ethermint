package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ethermint/utils"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryProtocolVersion:
			return queryProtocolVersion(keeper)
		case types.QueryBalance:
			return queryBalance(ctx, path, keeper)
		case types.QueryBlockNumber:
			return queryBlockNumber(ctx, keeper)
		case types.QueryStorage:
			return queryStorage(ctx, path, keeper)
		case types.QueryCode:
			return queryCode(ctx, path, keeper)
		case types.QueryNonce:
			return queryNonce(ctx, path, keeper)
		case types.QueryHashToHeight:
			return queryHashToHeight(ctx, path, keeper)
		case types.QueryTxLogs:
			return queryTxLogs(ctx, path, keeper)
		case types.QueryLogsBloom:
			return queryBlockLogsBloom(ctx, path, keeper)
		case types.QueryLogs:
			return queryLogs(ctx, keeper)
		case types.QueryAccount:
			return queryAccount(ctx, path, keeper)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown query endpoint")
		}
	}
}

func queryProtocolVersion(keeper Keeper) ([]byte, error) {
	vers := version.ProtocolVersion

	bz, err := codec.MarshalJSONIndent(keeper.cdc, hexutil.Uint(vers))
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryBalance(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	balance := keeper.GetBalance(ctx, addr)

	res := types.QueryResBalance{Balance: utils.MarshalBigInt(balance)}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryBlockNumber(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	num := ctx.BlockHeight()
	bnRes := types.QueryResBlockNumber{Number: num}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, bnRes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryStorage(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	key := ethcmn.HexToHash(path[2])
	val := keeper.GetState(ctx, addr, key)
	res := types.QueryResStorage{Value: val.Bytes()}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryCode(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	code := keeper.GetCode(ctx, addr)
	res := types.QueryResCode{Code: code}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryNonce(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	nonce := keeper.GetNonce(ctx, addr)
	nRes := types.QueryResNonce{Nonce: nonce}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, nRes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryHashToHeight(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	blockHash := ethcmn.FromHex(path[1])
	blockNumber := keeper.GetBlockHashMapping(ctx, blockHash)

	res := types.QueryResBlockNumber{Number: blockNumber}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryBlockLogsBloom(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	num, err := strconv.ParseInt(path[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal block number: %w", err)
	}

	bloom, err := keeper.GetBlockBloomMapping(ctx, num)
	if err != nil {
		return nil, fmt.Errorf("failed to get block bloom mapping: %w", err)
	}

	res := types.QueryBloomFilter{Bloom: bloom}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryTxLogs(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	txHash := ethcmn.HexToHash(path[1])
	logs, err := keeper.GetLogs(ctx, txHash)
	if err != nil {
		return nil, err
	}

	res := types.QueryETHLogs{Logs: logs}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryLogs(ctx sdk.Context, keeper Keeper) ([]byte, error) {
	logs := keeper.AllLogs(ctx)

	res := types.QueryETHLogs{Logs: logs}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

func queryAccount(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	so := keeper.GetOrNewStateObject(ctx, addr)

	res := types.QueryResAccount{
		Balance:  utils.MarshalBigInt(so.Balance()),
		CodeHash: so.CodeHash(),
		Nonce:    so.Nonce(),
	}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

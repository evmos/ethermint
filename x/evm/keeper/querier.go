package keeper

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/utils"
	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryBalance:
			return queryBalance(ctx, path, keeper)
		case types.QueryBlockNumber:
			return queryBlockNumber(ctx, keeper)
		case types.QueryStorage:
			return queryStorage(ctx, path, keeper)
		case types.QueryCode:
			return queryCode(ctx, path, keeper)
		case types.QueryHashToHeight:
			return queryHashToHeight(ctx, path, keeper)
		case types.QueryTransactionLogs:
			return queryTransactionLogs(ctx, path, keeper)
		case types.QueryBloom:
			return queryBlockBloom(ctx, path, keeper)
		case types.QueryLogs:
			return queryLogs(ctx, keeper)
		case types.QueryAccount:
			return queryAccount(ctx, path, keeper)
		default:
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "unknown query endpoint")
		}
	}
}

func queryBalance(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	addr := ethcmn.HexToAddress(path[1])
	balance := keeper.GetBalance(ctx, addr)
	balanceStr, err := utils.MarshalBigInt(balance)
	if err != nil {
		return nil, err
	}

	res := types.QueryResBalance{Balance: balanceStr}
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

func queryHashToHeight(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	blockHash := ethcmn.FromHex(path[1])
	blockNumber, found := keeper.GetBlockHash(ctx, blockHash)
	if !found {
		return []byte{}, fmt.Errorf("block height not found for hash %s", path[1])
	}

	res := types.QueryResBlockNumber{Number: blockNumber}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryBlockBloom(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
	num, err := strconv.ParseInt(path[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal block height: %w", err)
	}

	bloom, found := keeper.GetBlockBloom(ctx.WithBlockHeight(num), num)
	if !found {
		return nil, fmt.Errorf("block bloom not found for height %d", num)
	}

	res := types.QueryBloomFilter{Bloom: bloom}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryTransactionLogs(ctx sdk.Context, path []string, keeper Keeper) ([]byte, error) {
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

	balance, err := utils.MarshalBigInt(so.Balance())
	if err != nil {
		return nil, err
	}

	res := types.QueryResAccount{
		Balance:  balance,
		CodeHash: so.CodeHash(),
		Nonce:    so.Nonce(),
	}
	bz, err := codec.MarshalJSONIndent(keeper.cdc, res)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}
	return bz, nil
}

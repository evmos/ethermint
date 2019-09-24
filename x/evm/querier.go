package evm

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/version"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Supported endpoints
const (
	QueryProtocolVersion = "protocolVersion"
	QueryBalance         = "balance"
	QueryBlockNumber     = "blockNumber"
	QueryStorage         = "storage"
	QueryCode            = "code"
	QueryNonce           = "nonce"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryProtocolVersion:
			return queryProtocolVersion(keeper)
		case QueryBalance:
			return queryBalance(ctx, path, keeper)
		case QueryBlockNumber:
			return queryBlockNumber(ctx, keeper)
		case QueryStorage:
			return queryStorage(ctx, path, keeper)
		case QueryCode:
			return queryCode(ctx, path, keeper)
		case QueryNonce:
			return queryNonce(ctx, path, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown query endpoint")
		}
	}
}

func queryProtocolVersion(keeper Keeper) ([]byte, sdk.Error) {
	vers := version.ProtocolVersion

	bigRes := hexutil.Uint(vers)
	res, err := codec.MarshalJSONIndent(keeper.cdc, bigRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryBalance(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	balance := keeper.GetBalance(ctx, addr)
	res, err := codec.MarshalJSONIndent(keeper.cdc, balance)
	if err != nil {
		panic("could not marshal result to JSON: ")
	}

	return res, nil
}

func queryBlockNumber(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	num := ctx.BlockHeight()
	hexUint := hexutil.Uint64(num)

	res, err := codec.MarshalJSONIndent(keeper.cdc, hexUint)

	if err != nil {
		panic("could not marshal result to JSON: " + err.Error())
	}

	return res, nil
}

func queryStorage(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	key := ethcmn.BytesToHash([]byte(path[2]))
	val := keeper.GetState(ctx, addr, key)
	bRes := hexutil.Bytes(val.Bytes())
	res, err := codec.MarshalJSONIndent(keeper.cdc, &bRes)
	if err != nil {
		panic("could not marshal result to JSON: " + err.Error())
	}
	return res, nil
}

func queryCode(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	code := keeper.GetCode(ctx, addr)
	res, err := codec.MarshalJSONIndent(keeper.cdc, code)
	if err != nil {
		panic("could not marshal result to JSON: " + err.Error())
	}

	return res, nil
}

func queryNonce(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	nonce := keeper.GetNonce(ctx, addr)
	nRes := hexutil.Uint64(nonce)
	res, err := codec.MarshalJSONIndent(keeper.cdc, nRes)
	if err != nil {
		panic("could not marshal result to JSON: " + err.Error())
	}

	return res, nil
}

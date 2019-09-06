package evm

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm/types"
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

	res, err := codec.MarshalJSONIndent(keeper.cdc, vers)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryBalance(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	balance := keeper.GetBalance(ctx, addr)
	hBalance := &hexutil.Big{}
	err := hBalance.UnmarshalText(balance.Bytes())
	if err != nil {
		panic("could not marshal big.Int to hexutil.Big")
	}

	bRes := types.QueryResBalance{Balance: hBalance}
	res, err := codec.MarshalJSONIndent(keeper.cdc, bRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryBlockNumber(ctx sdk.Context, keeper Keeper) ([]byte, sdk.Error) {
	num := ctx.BlockHeight()
	bnRes := types.QueryResBlockNumber{Number: big.NewInt(num)}
	res, err := codec.MarshalJSONIndent(keeper.cdc, bnRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryStorage(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	key := ethcmn.BytesToHash([]byte(path[2]))
	val := keeper.GetState(ctx, addr, key)
	bRes := types.QueryResStorage{Value: val.Bytes()}
	res, err := codec.MarshalJSONIndent(keeper.cdc, bRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

func queryCode(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	code := keeper.GetCode(ctx, addr)
	cRes := types.QueryResCode{Code: code}
	res, err := codec.MarshalJSONIndent(keeper.cdc, cRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

func queryNonce(ctx sdk.Context, path []string, keeper Keeper) ([]byte, sdk.Error) {
	addr := ethcmn.BytesToAddress([]byte(path[1]))
	nonce := keeper.GetNonce(ctx, addr)
	nRes := types.QueryResNonce{Nonce: nonce}
	res, err := codec.MarshalJSONIndent(keeper.cdc, nRes)
	if err != nil {
		panic("could not marshal result to JSON")
	}

	return res, nil
}

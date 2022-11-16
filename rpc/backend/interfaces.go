package backend

import (
	"context"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type TMSignClient interface {
	Block(ctx context.Context, height *int64) (*ctypes.ResultBlock, error)
	BlockByHash(ctx context.Context, hash []byte) (*ctypes.ResultBlock, error)
	BlockResults(ctx context.Context, height *int64) (*ctypes.ResultBlockResults, error)
	Header(ctx context.Context, height *int64) (*ctypes.ResultHeader, error)
	Commit(ctx context.Context, height *int64) (*ctypes.ResultCommit, error)
	Validators(ctx context.Context, height *int64, page, perPage *int) (*ctypes.ResultValidators, error)
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
}

package types

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"

	"google.golang.org/grpc/metadata"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// BlockNumber represents decoding hex string to block values
type BlockNumber int64

const (
	// LatestBlockNumber mapping from "latest" to 0 for tm query
	LatestBlockNumber = BlockNumber(0)

	// EarliestBlockNumber mapping from "earliest" to 1 for tm query (earliest query not supported)
	EarliestBlockNumber = BlockNumber(1)

	// PendingBlockNumber mapping from "pending" to -1 for tm query
	PendingBlockNumber = BlockNumber(-1)
)

// NewBlockNumber creates a new BlockNumber instance.
func NewBlockNumber(n *big.Int) BlockNumber {
	return BlockNumber(n.Int64())
}

// ContextWithHeight wraps a context with the a gRPC block height header. If the provided height is
// 0 or -1, it will return an empty context and the gRPC query will use the latest block height for querying.
func ContextWithHeight(height int64) context.Context {
	if height == LatestBlockNumber.Int64() || height == PendingBlockNumber.Int64() {
		return context.Background()
	}

	return metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", height))
}

// UnmarshalJSON parses the given JSON fragment into a BlockNumber. It supports:
// - "latest", "earliest" or "pending" as string arguments
// - the block number
// Returned errors:
// - an invalid block number error when the given argument isn't a known strings
// - an out of range error when the given block number is either too little or too large
func (bn *BlockNumber) UnmarshalJSON(data []byte) error {
	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	switch input {
	case "earliest":
		*bn = EarliestBlockNumber
		return nil
	case "latest":
		*bn = LatestBlockNumber
		return nil
	case "pending":
		*bn = PendingBlockNumber
		return nil
	}

	blckNum, err := hexutil.DecodeUint64(input)
	if err != nil {
		return err
	}
	if blckNum > math.MaxInt64 {
		return fmt.Errorf("blocknumber too high")
	}

	*bn = BlockNumber(blckNum)
	return nil
}

// Int64 converts block number to primitive type
func (bn BlockNumber) Int64() int64 {
	return int64(bn)
}

// TmHeight is a util function used for the Tendermint RPC client. It returns
// nil if the block number is "latest". Otherwise, it returns the pointer of the
// int64 value of the height.
func (bn BlockNumber) TmHeight() *int64 {
	if bn == LatestBlockNumber {
		return nil
	}
	height := bn.Int64()
	return &height
}

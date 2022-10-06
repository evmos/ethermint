package types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

// BlockNumber represents decoding hex string to block values
type BlockNumber int64

const (
	EthSafeBlockNumber      = BlockNumber(-4)
	EthFinalizedBlockNumber = BlockNumber(-3)
	EthPendingBlockNumber   = BlockNumber(-2)
	EthLatestBlockNumber    = BlockNumber(-1)
	EthEarliestBlockNumber  = BlockNumber(0)
)

const (
	BlockParamEarliest  = "earliest"
	BlockParamLatest    = "latest"
	BlockParamFinalized = "finalized"
	BlockParamPending   = "pending"
	BlockParamSafe      = "safe"
)

// NewBlockNumber creates a new BlockNumber instance.
func NewBlockNumber(n *big.Int) BlockNumber {
	if !n.IsInt64() {
		// default to latest block if it overflows
		return EthLatestBlockNumber
	}

	return BlockNumber(n.Int64())
}

// ContextWithHeight wraps a context with the a gRPC block height header. If the provided height is
// 0, it will return an empty context and the gRPC query will use the latest block height for querying.
// Note that all metadata are processed and removed by tendermint layer, so it wont be accessible at gRPC server level.
func ContextWithHeight(height int64) context.Context {
	if height == 0 {
		return context.Background()
	}

	return metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", height))
}

// UnmarshalJSON parses the given JSON fragment into a BlockNumber. It supports:
// - "latest", "finalized", "earliest" or "pending" as string arguments
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
	case BlockParamEarliest:
		*bn = EthEarliestBlockNumber
		return nil
	case BlockParamLatest:
		*bn = EthLatestBlockNumber
		return nil
	case BlockParamFinalized:
		*bn = EthFinalizedBlockNumber
		return nil
	case BlockParamPending:
		*bn = EthPendingBlockNumber
		return nil
	case BlockParamSafe:
		*bn = EthSafeBlockNumber
		return nil
	}

	blckNum, err := hexutil.DecodeUint64(input)
	if errors.Is(err, hexutil.ErrMissingPrefix) {
		blckNum = cast.ToUint64(input)
	} else if err != nil {
		return err
	}

	if blckNum > math.MaxInt64 {
		return fmt.Errorf("block number larger than int64")
	}
	*bn = BlockNumber(blckNum)

	return nil
}

// MarshalText implements encoding.TextMarshaler. It marshals:
// - "latest", "earliest" or "pending" as strings
// - other numbers as hex
func (bn BlockNumber) MarshalText() ([]byte, error) {
	switch bn {
	case EthEarliestBlockNumber:
		return []byte("earliest"), nil
	case EthLatestBlockNumber:
		return []byte("latest"), nil
	case EthPendingBlockNumber:
		return []byte("pending"), nil
	case EthFinalizedBlockNumber:
		return []byte("finalized"), nil
	case EthSafeBlockNumber:
		return []byte("safe"), nil
	default:
		return hexutil.Uint64(bn).MarshalText()
	}
}

// Int64 converts block number to primitive type
func (bn BlockNumber) Int64() int64 {
	if bn < 0 {
		return 0
	} else if bn == 0 {
		return 1
	}

	return int64(bn)
}

// TmHeight is a util function used for the Tendermint RPC client. It returns
// nil if the block number is "latest". Otherwise, it returns the pointer of the
// int64 value of the height.
func (bn BlockNumber) TmHeight() *int64 {
	if bn < 0 {
		return nil
	}

	height := bn.Int64()
	return &height
}

// BlockNumberOrHash represents a block number or a block hash.
type BlockNumberOrHash struct {
	BlockNumber      *BlockNumber `json:"blockNumber,omitempty"`
	BlockHash        *common.Hash `json:"blockHash,omitempty"`
	RequireCanonical bool         `json:"requireCanonical,omitempty"`
}

func (bnh *BlockNumberOrHash) UnmarshalJSON(data []byte) error {
	type erased BlockNumberOrHash
	e := erased{}
	err := json.Unmarshal(data, &e)
	if err == nil {
		return bnh.checkUnmarshal(BlockNumberOrHash(e))
	}
	var input string
	err = json.Unmarshal(data, &input)
	if err != nil {
		return err
	}
	err = bnh.decodeFromString(input)
	if err != nil {
		return err
	}

	return nil
}

func (bnh *BlockNumberOrHash) checkUnmarshal(e BlockNumberOrHash) error {
	if e.BlockNumber != nil && e.BlockHash != nil {
		return fmt.Errorf("cannot specify both BlockHash and BlockNumber, choose one or the other")
	}
	bnh.BlockNumber = e.BlockNumber
	bnh.BlockHash = e.BlockHash
	return nil
}

func (bnh *BlockNumberOrHash) decodeFromString(input string) error {
	switch input {
	case BlockParamEarliest:
		bn := EthEarliestBlockNumber
		bnh.BlockNumber = &bn
		return nil
	case BlockParamLatest:
		bn := EthLatestBlockNumber
		bnh.BlockNumber = &bn
		return nil
	case BlockParamPending:
		bn := EthPendingBlockNumber
		bnh.BlockNumber = &bn
		return nil
	case BlockParamFinalized:
		bn := EthFinalizedBlockNumber
		bnh.BlockNumber = &bn
		return nil
	case BlockParamSafe:
		bn := EthSafeBlockNumber
		bnh.BlockNumber = &bn
		return nil
	default:
		if len(input) == 66 {
			hash := common.Hash{}
			err := hash.UnmarshalText([]byte(input))
			if err != nil {
				return err
			}
			bnh.BlockHash = &hash
			return nil
		}

		blckNum, err := hexutil.DecodeUint64(input)
		if err != nil {
			return err
		}
		if blckNum > math.MaxInt64 {
			return fmt.Errorf("blocknumber too high")
		}
		bn := BlockNumber(blckNum)
		bnh.BlockNumber = &bn
		return nil
	}
}

func (bnh *BlockNumberOrHash) Number() (BlockNumber, bool) {
	if bnh.BlockNumber != nil {
		return *bnh.BlockNumber, true
	}
	return BlockNumber(0), false
}

func (bnh *BlockNumberOrHash) String() string {
	if bnh.BlockNumber != nil {
		return strconv.Itoa(int(*bnh.BlockNumber))
	}
	if bnh.BlockHash != nil {
		return bnh.BlockHash.String()
	}
	return "nil"
}

func (bnh *BlockNumberOrHash) Hash() (common.Hash, bool) {
	if bnh.BlockHash != nil {
		return *bnh.BlockHash, true
	}
	return common.Hash{}, false
}

func BlockNumberOrHashWithNumber(blockNr BlockNumber) BlockNumberOrHash {
	return BlockNumberOrHash{
		BlockNumber:      &blockNr,
		BlockHash:        nil,
		RequireCanonical: false,
	}
}

func BlockNumberOrHashWithHash(hash common.Hash, canonical bool) BlockNumberOrHash {
	return BlockNumberOrHash{
		BlockNumber:      nil,
		BlockHash:        &hash,
		RequireCanonical: canonical,
	}
}

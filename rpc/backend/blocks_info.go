package backend

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/pkg/errors"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Getting Blocks
//
// Retrieves information from a particular block in the blockchain.
// BlockNumber() (hexutil.Uint64, error)
// GetBlockByNumber(ethBlockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error)
// GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)

// BlockNumber returns the current block number in abci app state.
// Because abci app state could lag behind from tendermint latest block, it's more stable
// for the client to use the latest block number in abci app state than tendermint rpc.
func (b *Backend) BlockNumber() (hexutil.Uint64, error) {
	// do any grpc query, ignore the response and use the returned block height
	var header metadata.MD
	_, err := b.queryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	if err != nil {
		return hexutil.Uint64(0), err
	}

	blockHeightHeader := header.Get(grpctypes.GRPCBlockHeightHeader)
	if headerLen := len(blockHeightHeader); headerLen != 1 {
		return 0, fmt.Errorf("unexpected '%s' gRPC header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, headerLen, 1)
	}

	height, err := strconv.ParseUint(blockHeightHeader[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	return hexutil.Uint64(height), nil
}

// GetBlockByNumber returns the block identified by number.
func (b *Backend) GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, nil
	}

	// return if requested block height is greater than the current one
	if resBlock == nil || resBlock.Block == nil {
		return nil, nil
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("failed to fetch block result from Tendermint", "height", blockNum, "error", err.Error())
		return nil, nil
	}

	res, err := b.EthBlockFromTendermint(resBlock, blockRes, fullTx)
	if err != nil {
		b.logger.Debug("EthBlockFromTendermint failed", "height", blockNum, "error", err.Error())
		return nil, err
	}

	return res, nil
}

// GetBlockByHash returns the block identified by hash.
func (b *Backend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := b.GetTendermintBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	if resBlock == nil {
		// block not found
		return nil, nil
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("failed to fetch block result from Tendermint", "block-hash", hash.String(), "error", err.Error())
		return nil, nil
	}

	res, err := b.EthBlockFromTendermint(resBlock, blockRes, fullTx)
	if err != nil {
		b.logger.Debug("EthBlockFromTendermint failed", "hash", hash, "error", err.Error())
		return nil, err
	}

	return res, nil
}

// GetTendermintBlockByNumber returns a Tendermint formatted block for a given
// block number
func (b *Backend) GetTendermintBlockByNumber(blockNum rpctypes.BlockNumber) (*tmrpctypes.ResultBlock, error) {
	height := blockNum.Int64()
	if height <= 0 {
		// fetch the latest block number from the app state, more accurate than the tendermint block store state.
		n, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}
		height = int64(n)
	}
	resBlock, err := b.clientCtx.Client.Block(b.ctx, &height)
	if err != nil {
		b.logger.Debug("tendermint client failed to get block", "height", height, "error", err.Error())
		return nil, err
	}

	if resBlock.Block == nil {
		b.logger.Debug("GetTendermintBlockByNumber block not found", "height", height)
		return nil, nil
	}

	return resBlock, nil
}

// BlockBloom query block bloom filter from block results
func (b *Backend) BlockBloom(blockRes *tmrpctypes.ResultBlockResults) (ethtypes.Bloom, error) {
	for _, event := range blockRes.EndBlockEvents {
		if event.Type != evmtypes.EventTypeBlockBloom {
			continue
		}

		for _, attr := range event.Attributes {
			if bytes.Equal(attr.Key, bAttributeKeyEthereumBloom) {
				return ethtypes.BytesToBloom(attr.Value), nil
			}
		}
	}
	return ethtypes.Bloom{}, errors.New("block bloom event is not found")
}

// GetTendermintBlockResultByNumber returns a Tendermint-formatted block result by block number
func (b *Backend) GetTendermintBlockResultByNumber(height *int64) (*tmrpctypes.ResultBlockResults, error) {
	return b.clientCtx.Client.BlockResults(b.ctx, height)
}

// GetTendermintBlockByHash returns a Tendermint format block by block number
func (b *Backend) GetTendermintBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, blockHash.Bytes())
	if err != nil {
		b.logger.Debug("tendermint client failed to get block", "blockHash", blockHash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		b.logger.Debug("GetTendermintBlockByHash block not found", "blockHash", blockHash.Hex())
		return nil, nil
	}

	return resBlock, nil
}

// BlockByNumber returns the block identified by number.
func (b *Backend) BlockByNumber(blockNum rpctypes.BlockNumber) (*ethtypes.Block, error) {
	resBlock, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		// block not found
		return nil, fmt.Errorf("block not found for height %d", blockNum)
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, fmt.Errorf("block result not found for height %d", resBlock.Block.Height)
	}

	return b.EthBlockFromTm(resBlock, blockRes)
}

// BlockByHash returns the block identified by hash.
func (b *Backend) BlockByHash(hash common.Hash) (*ethtypes.Block, error) {
	resBlock, err := b.GetTendermintBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		return nil, fmt.Errorf("block not found for hash %s", hash)
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, fmt.Errorf("block result not found for hash %s", hash)
	}

	return b.EthBlockFromTm(resBlock, blockRes)
}

// GetBlockNumberByHash returns the block height of given block hash
func (b *Backend) GetBlockNumberByHash(blockHash common.Hash) (*big.Int, error) {
	resBlock, err := b.GetTendermintBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		return nil, errors.Errorf("block not found for hash %s", blockHash.Hex())
	}
	return big.NewInt(resBlock.Block.Height), nil
}

// getBlockNumber returns the BlockNumber from BlockNumberOrHash
func (b *Backend) GetBlockNumber(blockNrOrHash rpctypes.BlockNumberOrHash) (rpctypes.BlockNumber, error) {
	switch {
	case blockNrOrHash.BlockHash == nil && blockNrOrHash.BlockNumber == nil:
		return rpctypes.EthEarliestBlockNumber, fmt.Errorf("types BlockHash and BlockNumber cannot be both nil")
	case blockNrOrHash.BlockHash != nil:
		blockNumber, err := b.GetBlockNumberByHash(*blockNrOrHash.BlockHash)
		if err != nil {
			return rpctypes.EthEarliestBlockNumber, err
		}
		return rpctypes.NewBlockNumber(blockNumber), nil
	case blockNrOrHash.BlockNumber != nil:
		return *blockNrOrHash.BlockNumber, nil
	default:
		return rpctypes.EthEarliestBlockNumber, nil
	}
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (b *Backend) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	block, err := b.clientCtx.Client.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		b.logger.Debug("block not found", "hash", hash.Hex(), "error", err.Error())
		return nil
	}

	if block.Block == nil {
		b.logger.Debug("block not found", "hash", hash.Hex())
		return nil
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		return nil
	}

	ethMsgs := b.GetEthereumMsgsFromTendermintBlock(block, blockRes)
	n := hexutil.Uint(len(ethMsgs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (b *Backend) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	block, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		b.logger.Debug("block not found", "height", blockNum.Int64(), "error", err.Error())
		return nil
	}

	if block.Block == nil {
		b.logger.Debug("block not found", "height", blockNum.Int64())
		return nil
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		return nil
	}

	ethMsgs := b.GetEthereumMsgsFromTendermintBlock(block, blockRes)
	n := hexutil.Uint(len(ethMsgs))
	return &n
}

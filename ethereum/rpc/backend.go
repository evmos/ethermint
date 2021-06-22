package rpc

import (
	"context"
	"math/big"
	"regexp"

	"github.com/tharsis/ethermint/ethereum/rpc/types"

	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
	log "github.com/xlab/suplog"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Backend implements the functionality needed to filter changes.
// Implemented by EVMBackend.
type Backend interface {
	// Used by block filter; also used for polling
	BlockNumber() (hexutil.Uint64, error)
	HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)

	// returns the logs of a given block
	GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error)

	// Used by pending transaction filter
	PendingTransactions() ([]*types.RPCTransaction, error)

	// Used by log filter
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	BloomStatus() (uint64, uint64)
}

var _ Backend = (*EVMBackend)(nil)

// EVMBackend implements the Backend interface
type EVMBackend struct {
	ctx         context.Context
	clientCtx   client.Context
	queryClient *types.QueryClient // gRPC query client
	logger      log.Logger
}

// NewEVMBackend creates a new EVMBackend instance
func NewEVMBackend(clientCtx client.Context) *EVMBackend {
	return &EVMBackend{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: types.NewQueryClient(clientCtx),
		logger:      log.WithField("module", "evm-backend"),
	}
}

// BlockNumber returns the current block number.
func (e *EVMBackend) BlockNumber() (hexutil.Uint64, error) {
	// NOTE: using 0 as min and max height returns the blockchain info up to the latest block.
	info, err := e.clientCtx.Client.BlockchainInfo(e.ctx, 0, 0)
	if err != nil {
		return hexutil.Uint64(0), err
	}

	return hexutil.Uint64(info.LastHeight), nil
}

// GetBlockByNumber returns the block identified by number.
func (e *EVMBackend) GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber - 1)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			err := errors.Errorf("incorrect block height: %d", height)
			return nil, err
		} else if height > int64(currentBlockNumber) {
			return nil, nil
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		// e.logger.Debugf("GetBlockByNumber safely bumping down from %d to latest", height)
		if resBlock, err = e.clientCtx.Client.Block(e.ctx, nil); err != nil {
			e.logger.Debugln("GetBlockByNumber failed to get latest block")
			return nil, nil
		}
	}

	res, err := e.EthBlockFromTendermint(e.clientCtx, e.queryClient, resBlock.Block, fullTx)
	if err != nil {
		e.logger.WithError(err).Debugf("EthBlockFromTendermint failed with block %s", resBlock.Block.String())
	}

	return res, err
}

// GetBlockByHash returns the block identified by hash.
func (e *EVMBackend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		e.logger.Warningf("BlockByHash failed for %s", hash.Hex())
		return nil, err
	}

	return e.EthBlockFromTendermint(e.clientCtx, e.queryClient, resBlock.Block, fullTx)
}

// EthBlockFromTendermint returns a JSON-RPC compatible Ethereum block from a given Tendermint block.
func (e *EVMBackend) EthBlockFromTendermint(
	clientCtx client.Context,
	queryClient evmtypes.QueryClient,
	block *tmtypes.Block,
	fullTx bool,
) (map[string]interface{}, error) {

	gasUsed := uint64(0)

	ethRPCTxs := []interface{}{}

	for i, txBz := range block.Txs {
		// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
		// https://github.com/tendermint/tendermint/issues/6539
		hash := common.BytesToHash(txBz.Hash())

		tx, gas := types.DecodeTx(e.clientCtx, txBz)

		gasUsed += gas

		msg, isEthTx := tx.(*evmtypes.MsgEthereumTx)

		if fullTx {
			if !isEthTx {
				// TODO: eventually support Cosmos txs in the block
				continue
			}

			tx, err := types.NewTransactionFromData(
				msg.Data,
				common.HexToAddress(msg.From),
				hash,
				common.BytesToHash(block.Hash()),
				uint64(block.Height),
				uint64(i),
			)

			ethRPCTxs = append(ethRPCTxs, tx)

			if err != nil {
				e.logger.WithError(err).Debugln("NewTransactionFromData for receipt failed", "hash", hash.Hex)
				continue
			}
		} else {
			ethRPCTxs = append(ethRPCTxs, hash)
		}
	}

	blockBloomResp, err := queryClient.BlockBloom(types.ContextWithHeight(block.Height), &evmtypes.QueryBlockBloomRequest{})
	if err != nil {
		e.logger.WithError(err).Debugln("failed to query BlockBloom", "height", block.Height)

		blockBloomResp = &evmtypes.QueryBlockBloomResponse{Bloom: ethtypes.Bloom{}.Bytes()}
	}

	bloom := ethtypes.BytesToBloom(blockBloomResp.Bloom)
	formattedBlock := types.FormatBlock(block.Header, block.Size(), ethermint.DefaultRPCGasLimit, new(big.Int).SetUint64(gasUsed), ethRPCTxs, bloom)

	e.logger.Infoln(formattedBlock)
	return formattedBlock, nil
}

// HeaderByNumber returns the block header identified by height.
func (e *EVMBackend) HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber - 1)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			err := errors.Errorf("incorrect block height: %d", height)
			return nil, err
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		e.logger.Warningf("HeaderByNumber failed")
		return nil, err
	}

	req := &evmtypes.QueryBlockBloomRequest{}

	res, err := e.queryClient.BlockBloom(types.ContextWithHeight(resBlock.Block.Height), req)
	if err != nil {
		e.logger.Warningf("HeaderByNumber BlockBloom fail %d", resBlock.Block.Height)
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(res.Bloom)
	return ethHeader, nil
}

// HeaderByHash returns the block header identified by hash.
func (e *EVMBackend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, blockHash.Bytes())
	if err != nil {
		e.logger.Warningf("HeaderByHash fail")
		return nil, err
	}

	req := &evmtypes.QueryBlockBloomRequest{}

	res, err := e.queryClient.BlockBloom(types.ContextWithHeight(resBlock.Block.Height), req)
	if err != nil {
		e.logger.Warningf("HeaderByHash BlockBloom fail %d", resBlock.Block.Height)
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(res.Bloom)
	return ethHeader, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
// It returns an error if there's an encoding error.
// If no logs are found for the tx hash, the error is nil.
func (e *EVMBackend) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	req := &evmtypes.QueryTxLogsRequest{
		Hash: txHash.String(),
	}

	res, err := e.queryClient.TxLogs(e.ctx, req)
	if err != nil {
		e.logger.Warningf("TxLogs fail")
		return nil, err
	}

	return evmtypes.LogsToEthereum(res.Logs), nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *EVMBackend) PendingTransactions() ([]*types.RPCTransaction, error) {
	return []*types.RPCTransaction{}, nil
}

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (e *EVMBackend) GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	req := &evmtypes.QueryBlockLogsRequest{
		Hash: blockHash.String(),
	}

	res, err := e.queryClient.BlockLogs(e.ctx, req)
	if err != nil {
		e.logger.Warningf("BlockLogs fail")
		return nil, err
	}

	var blockLogs = [][]*ethtypes.Log{}
	for _, txLog := range res.TxLogs {
		blockLogs = append(blockLogs, txLog.EthLogs())
	}

	return blockLogs, nil
}

// This is very brittle, see: https://github.com/tendermint/tendermint/issues/4740
var regexpMissingHeight = regexp.MustCompile(`height \d+ (must be less than or equal to|is not available)`)

func (e *EVMBackend) GetLogsByNumber(blockNum types.BlockNumber) ([][]*ethtypes.Log, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber - 1)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			err := errors.Errorf("incorrect block height: %d", height)
			return nil, err
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		if regexpMissingHeight.MatchString(err.Error()) {
			return [][]*ethtypes.Log{}, nil
		}

		e.logger.Warningf("failed to query block at %d", height)
		return nil, err
	}

	return e.GetLogs(common.BytesToHash(resBlock.BlockID.Hash))
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (e *EVMBackend) BloomStatus() (uint64, uint64) {
	return 4096, 0
}

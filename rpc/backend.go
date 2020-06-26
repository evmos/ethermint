package rpc

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Backend implements the functionality needed to filter changes.
// Implemented by EthermintBackend.
type Backend interface {
	// Used by block filter; also used for polling
	BlockNumber() (hexutil.Uint64, error)
	GetBlockByNumber(blockNum BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	getEthBlockByNumber(height int64, fullTx bool) (map[string]interface{}, error)
	getGasLimit() (int64, error)

	// Used by pending transaction filter
	PendingTransactions() ([]*Transaction, error)

	// Used by log filter
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	// TODO: Bloom methods
}

// EthermintBackend implements Backend
type EthermintBackend struct {
	cliCtx   context.CLIContext
	gasLimit int64
}

// NewEthermintBackend creates a new EthermintBackend instance
func NewEthermintBackend(cliCtx context.CLIContext) *EthermintBackend {
	return &EthermintBackend{
		cliCtx:   cliCtx,
		gasLimit: int64(^uint32(0)),
	}
}

// BlockNumber returns the current block number.
func (e *EthermintBackend) BlockNumber() (hexutil.Uint64, error) {
	res, height, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/blockNumber", evmtypes.ModuleName), nil)
	if err != nil {
		return hexutil.Uint64(0), err
	}

	var out evmtypes.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)

	e.cliCtx.WithHeight(height)
	return hexutil.Uint64(out.Number), nil
}

// GetBlockByNumber returns the block identified by number.
func (e *EthermintBackend) GetBlockByNumber(blockNum BlockNumber, fullTx bool) (map[string]interface{}, error) {
	value := blockNum.Int64()
	return e.getEthBlockByNumber(value, fullTx)
}

// GetBlockByHash returns the block identified by hash.
func (e *EthermintBackend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryHashToHeight, hash.Hex()))
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryResBlockNumber
	if err := e.cliCtx.Codec.UnmarshalJSON(res, &out); err != nil {
		return nil, err
	}

	return e.getEthBlockByNumber(out.Number, fullTx)
}

func (e *EthermintBackend) getEthBlockByNumber(height int64, fullTx bool) (map[string]interface{}, error) {
	// Remove this check when 0 query is fixed ref: (https://github.com/tendermint/tendermint/issues/4014)
	var blkNumPtr *int64
	if height != 0 {
		blkNumPtr = &height
	}

	block, err := e.cliCtx.Client.Block(blkNumPtr)
	if err != nil {
		return nil, err
	}
	header := block.Block.Header

	gasLimit, err := e.getGasLimit()
	if err != nil {
		return nil, err
	}

	var (
		gasUsed      *big.Int
		transactions []common.Hash
	)

	if fullTx {
		// Populate full transaction data
		transactions, gasUsed, err = convertTransactionsToRPC(
			e.cliCtx, block.Block.Txs, common.BytesToHash(header.Hash()), uint64(header.Height),
		)
		if err != nil {
			return nil, err
		}
	} else {
		// TODO: Gas used not saved and cannot be calculated by hashes
		// Return slice of transaction hashes
		transactions = make([]common.Hash, len(block.Block.Txs))
		for i, tx := range block.Block.Txs {
			transactions[i] = common.BytesToHash(tx.Hash())
		}
	}

	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryBloom, strconv.FormatInt(block.Block.Height, 10)))
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryBloomFilter
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)

	return formatBlock(header, block.Block.Size(), gasLimit, gasUsed, transactions, out.Bloom), nil
}

// getGasLimit returns the gas limit per block set in genesis
func (e *EthermintBackend) getGasLimit() (int64, error) {
	// Retrieve from gasLimit variable cache
	if e.gasLimit != -1 {
		return e.gasLimit, nil
	}

	// Query genesis block if hasn't been retrieved yet
	genesis, err := e.cliCtx.Client.Genesis()
	if err != nil {
		return 0, err
	}

	// Save value to gasLimit cached value
	gasLimit := genesis.Genesis.ConsensusParams.Block.MaxGas
	if gasLimit == -1 {
		// Sets gas limit to max uint32 to not error with javascript dev tooling
		// This -1 value indicating no block gas limit is set to max uint64 with geth hexutils
		// which errors certain javascript dev tooling which only supports up to 53 bits
		gasLimit = int64(^uint32(0))
	}
	e.gasLimit = gasLimit
	return gasLimit, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (e *EthermintBackend) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	// do we need to use the block height somewhere?
	ctx := e.cliCtx

	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryTransactionLogs, txHash.Hex()), nil)
	if err != nil {
		return nil, err
	}

	out := new(evmtypes.QueryETHLogs)
	if err := e.cliCtx.Codec.UnmarshalJSON(res, &out); err != nil {
		return nil, err
	}

	return out.Logs, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *EthermintBackend) PendingTransactions() ([]*Transaction, error) {
	pendingTxs, err := e.cliCtx.Client.UnconfirmedTxs(100)
	if err != nil {
		return nil, err
	}

	transactions := make([]*Transaction, 0, 100)
	for _, tx := range pendingTxs.Txs {
		ethTx, err := bytesToEthTx(e.cliCtx, tx)
		if err != nil {
			return nil, err
		}

		// * Should check signer and reference against accounts the node manages in future
		rpcTx, err := newRPCTransaction(*ethTx, common.BytesToHash(tx.Hash()), common.Hash{}, nil, 0)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, rpcTx)
	}

	return transactions, nil
}

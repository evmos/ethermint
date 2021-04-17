package backend

import (
	"context"
	"errors"
	"os"

	"github.com/tendermint/tendermint/libs/log"

	rpctypes "github.com/cosmos/ethermint/rpc/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Backend implements the functionality needed to filter changes.
// Implemented by EthermintBackend.
type Backend interface {
	// Used by block filter; also used for polling
	BlockNumber() (hexutil.Uint64, error)
	LatestBlockNumber() (int64, error)
	HeaderByNumber(blockNum rpctypes.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)

	// returns the logs of a given block
	GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error)

	// Used by pending transaction filter
	PendingTransactions() ([]*rpctypes.Transaction, error)

	// Used by log filter
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	BloomStatus() (uint64, uint64)
}

var _ Backend = (*EthermintBackend)(nil)

// EthermintBackend implements the Backend interface
type EthermintBackend struct {
	ctx         context.Context
	clientCtx   client.Context
	queryClient *rpctypes.QueryClient // gRPC query client
	logger      log.Logger
}

// New creates a new EthermintBackend instance
func New(clientCtx client.Context) *EthermintBackend {
	return &EthermintBackend{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: rpctypes.NewQueryClient(clientCtx),
		logger:      log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "json-rpc"),
	}
}

// BlockNumber returns the current block number.
func (b *EthermintBackend) BlockNumber() (hexutil.Uint64, error) {
	// NOTE: using 0 as min and max height returns the blockchain info up to the latest block.
	info, err := b.clientCtx.Client.BlockchainInfo(b.ctx, 0, 0)
	if err != nil {
		return hexutil.Uint64(0), err
	}

	return hexutil.Uint64(info.LastHeight), nil
}

// GetBlockByNumber returns the block identified by number.
func (b *EthermintBackend) GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	var height *int64
	// NOTE: here pending and latest are defined as a nil height, which fetches the latest block
	if !(blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber) {
		height = blockNum.TmHeight()
	}

	resBlock, err := b.clientCtx.Client.Block(b.ctx, height)
	if err != nil {
		return nil, err
	}

	if resBlock.BlockID.IsZero() {
		return nil, errors.New("failed to query block by number: nil block returned")
	}

	return rpctypes.EthBlockFromTendermint(b.clientCtx, b.queryClient, resBlock.Block)
}

// GetBlockByHash returns the block identified by hash.
func (b *EthermintBackend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}

	return rpctypes.EthBlockFromTendermint(b.clientCtx, b.queryClient, resBlock.Block)
}

// HeaderByNumber returns the block header identified by height.
func (b *EthermintBackend) HeaderByNumber(blockNum rpctypes.BlockNumber) (*ethtypes.Header, error) {
	var height *int64
	// NOTE: here pending and latest are defined as a nil height, which fetches the latest header
	if !(blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber) {
		height = blockNum.TmHeight()
	}

	resBlock, err := b.clientCtx.Client.Block(b.ctx, height)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBlockBloomRequest{}

	res, err := b.queryClient.BlockBloom(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	ethHeader := rpctypes.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(res.Bloom)
	return ethHeader, nil
}

// HeaderByHash returns the block header identified by hash.
func (b *EthermintBackend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, blockHash.Bytes())
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBlockBloomRequest{}

	res, err := b.queryClient.BlockBloom(rpctypes.ContextWithHeight(resBlock.Block.Height), req)
	if err != nil {
		return nil, err
	}

	ethHeader := rpctypes.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(res.Bloom)
	return ethHeader, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
// It returns an error if there's an encoding error.
// If no logs are found for the tx hash, the error is nil.
func (b *EthermintBackend) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	req := &evmtypes.QueryTxLogsRequest{
		Hash: txHash.String(),
	}

	res, err := b.queryClient.TxLogs(b.ctx, req)
	if err != nil {
		return nil, err
	}

	return evmtypes.LogsToEthereum(res.Logs), nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (b *EthermintBackend) PendingTransactions() ([]*rpctypes.Transaction, error) {
	limit := 1000
	pendingTxs, err := b.clientCtx.Client.UnconfirmedTxs(b.ctx, &limit)
	if err != nil {
		return nil, err
	}

	transactions := make([]*rpctypes.Transaction, 0)
	for _, tx := range pendingTxs.Txs {
		ethTx, err := rpctypes.RawTxToEthTx(b.clientCtx, tx)
		if err != nil {
			// ignore non Ethermint EVM transactions
			continue
		}

		// TODO: check signer and reference against accounts the node manages
		rpcTx, err := rpctypes.NewTransaction(ethTx, common.BytesToHash(tx.Hash()), common.Hash{}, 0, 0)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, rpcTx)
	}
	return transactions, nil
}

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (b *EthermintBackend) GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	req := &evmtypes.QueryBlockLogsRequest{
		Hash: blockHash.String(),
	}

	res, err := b.queryClient.BlockLogs(b.ctx, req)
	if err != nil {
		return nil, err
	}

	var blockLogs = [][]*ethtypes.Log{}
	for _, txLog := range res.TxLogs {
		blockLogs = append(blockLogs, txLog.EthLogs())
	}

	return blockLogs, nil
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (b *EthermintBackend) BloomStatus() (uint64, uint64) {
	return 4096, 0
}

// LatestBlockNumber gets the latest block height in int64 format.
func (b *EthermintBackend) LatestBlockNumber() (int64, error) {
	// NOTE: using 0 as min and max height returns the blockchain info up to the latest block.
	info, err := b.clientCtx.Client.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		return 0, err
	}

	return info.LastHeight, nil
}

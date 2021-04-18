package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/xlab/suplog"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	rpctypes "github.com/cosmos/ethermint/ethereum/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

type DataError interface {
	Error() string          // returns the message
	ErrorData() interface{} // returns the error data
}

type dataError struct {
	msg  string
	data string
}

func (d *dataError) Error() string {
	return d.msg
}

func (d *dataError) ErrorData() interface{} {
	return d.data
}

type sdkTxLogs struct {
	Log string `json:"log"`
}

const logRevertedFlag = "transaction reverted"

func errRevertedWith(data []byte) DataError {
	return &dataError{
		msg:  "VM execution error.",
		data: fmt.Sprintf("0x%s", hex.EncodeToString(data)),
	}
}

// NewTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func NewTransaction(tx *evmtypes.MsgEthereumTx, txHash, blockHash common.Hash, blockNumber, index uint64) (*rpctypes.Transaction, error) {
	// Verify signature and retrieve sender address
	from, err := tx.VerifySig(tx.ChainID())
	if err != nil {
		return nil, err
	}

	rpcTx := &rpctypes.Transaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Data.GasLimit),
		GasPrice: (*hexutil.Big)(new(big.Int).SetBytes(tx.Data.Price)),
		Hash:     txHash,
		Input:    hexutil.Bytes(tx.Data.Payload),
		Nonce:    hexutil.Uint64(tx.Data.AccountNonce),
		To:       tx.To(),
		Value:    (*hexutil.Big)(new(big.Int).SetBytes(tx.Data.Amount)),
		V:        (*hexutil.Big)(new(big.Int).SetBytes(tx.Data.V)),
		R:        (*hexutil.Big)(new(big.Int).SetBytes(tx.Data.R)),
		S:        (*hexutil.Big)(new(big.Int).SetBytes(tx.Data.S)),
	}
	if rpcTx.To == nil {
		addr := common.HexToAddress("0x0000000000000000000000000000000000000000")
		rpcTx.To = &addr
	}

	if blockHash != (common.Hash{}) {
		rpcTx.BlockHash = &blockHash
		rpcTx.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		rpcTx.TransactionIndex = (*hexutil.Uint64)(&index)
	}

	return rpcTx, nil
}

// NewTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func NewTransactionFromData(
	txData *evmtypes.TxData,
	from common.Address,
	txHash, blockHash common.Hash,
	blockNumber, index uint64,
) (*rpctypes.Transaction, error) {

	var to *common.Address
	if len(txData.Recipient) > 0 {
		recipient := common.BytesToAddress(txData.Recipient)
		to = &recipient
	}

	rpcTx := &rpctypes.Transaction{
		From:     from,
		Gas:      hexutil.Uint64(txData.GasLimit),
		GasPrice: (*hexutil.Big)(new(big.Int).SetBytes(txData.Price)),
		Hash:     txHash,
		Input:    hexutil.Bytes(txData.Payload),
		Nonce:    hexutil.Uint64(txData.AccountNonce),
		To:       to,
		Value:    (*hexutil.Big)(new(big.Int).SetBytes(txData.Amount)),
		V:        (*hexutil.Big)(new(big.Int).SetBytes(txData.V)),
		R:        (*hexutil.Big)(new(big.Int).SetBytes(txData.R)),
		S:        (*hexutil.Big)(new(big.Int).SetBytes(txData.S)),
	}
	if rpcTx.To == nil {
		addr := common.HexToAddress("0x0000000000000000000000000000000000000000")
		rpcTx.To = &addr
	}

	if blockHash != (common.Hash{}) {
		rpcTx.BlockHash = &blockHash
		rpcTx.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		rpcTx.TransactionIndex = (*hexutil.Uint64)(&index)
	}

	return rpcTx, nil
}

// EthHeaderFromTendermint is an util function that returns an Ethereum Header
// from a tendermint Header.
func EthHeaderFromTendermint(header tmtypes.Header) *ethtypes.Header {
	return &ethtypes.Header{
		ParentHash:  common.BytesToHash(header.LastBlockID.Hash.Bytes()),
		UncleHash:   common.Hash{},
		Coinbase:    common.Address{},
		Root:        common.BytesToHash(header.AppHash),
		TxHash:      common.BytesToHash(header.DataHash),
		ReceiptHash: common.Hash{},
		Difficulty:  nil,
		Number:      big.NewInt(header.Height),
		Time:        uint64(header.Time.Unix()),
		Extra:       nil,
		MixDigest:   common.Hash{},
		Nonce:       ethtypes.BlockNonce{},
	}
}

// BlockMaxGasFromConsensusParams returns the gas limit for the latest block from the chain consensus params.
func BlockMaxGasFromConsensusParams(ctx context.Context, clientCtx client.Context, height *int64, logger suplog.Logger) (int64, error) {
	return ethermint.DefaultRPCGasLimit, nil
}

var zeroHash = hexutil.Bytes(make([]byte, 32))

func hashOrZero(data []byte) hexutil.Bytes {
	if len(data) == 0 {
		return zeroHash
	}

	return hexutil.Bytes(data)
}

func bigOrZero(i *big.Int) *hexutil.Big {
	if i == nil {
		return new(hexutil.Big)
	}

	return (*hexutil.Big)(i)
}

func formatBlock(
	header tmtypes.Header, size int, gasLimit int64,
	gasUsed *big.Int, transactions interface{}, bloom ethtypes.Bloom,
) map[string]interface{} {
	if len(header.DataHash) == 0 {
		header.DataHash = tmbytes.HexBytes(common.Hash{}.Bytes())
	}

	var txRoot interface{}

	txDescriptors, ok := transactions.([]interface{})
	if !ok || len(txDescriptors) == 0 {
		txRoot = ethtypes.EmptyRootHash
		transactions = []common.Hash{}
	} else {
		txRoot = hashOrZero(header.DataHash)
	}

	ret := map[string]interface{}{
		"parentHash":       hashOrZero(header.LastBlockID.Hash),
		"sha3Uncles":       ethtypes.EmptyUncleHash, // No uncles in Tendermint
		"miner":            common.Address{},
		"stateRoot":        hashOrZero(header.AppHash),
		"transactionsRoot": txRoot,
		"receiptsRoot":     zeroHash,
		"logsBloom":        hexutil.Encode(bloom.Bytes()),
		"difficulty":       new(hexutil.Big),
		"number":           hexutil.Uint64(header.Height),
		"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
		"gasUsed":          bigOrZero(gasUsed),
		"timestamp":        hexutil.Uint64(header.Time.Unix()),
		"extraData":        hexutil.Bytes([]byte{}),
		"mixHash":          zeroHash,
		"hash":             hashOrZero(header.Hash()),
		"nonce":            ethtypes.EncodeNonce(0),
		"totalDifficulty":  new(hexutil.Big),
		"size":             hexutil.Uint64(size),
		"transactions":     transactions,
		"uncles":           []string{},
	}

	return ret
}

// GetBlockCumulativeGas returns the cumulative gas used on a block up to a given
// transaction index. The returned gas used includes the gas from both the SDK and
// EVM module transactions.
func GetBlockCumulativeGas(clientCtx client.Context, block *tmtypes.Block, idx int) uint64 {
	var gasUsed uint64
	txDecoder := clientCtx.TxConfig.TxDecoder()

	for i := 0; i < idx && i < len(block.Txs); i++ {
		txi, err := txDecoder(block.Txs[i])
		if err != nil {
			continue
		}

		switch tx := txi.(type) {
		case *evmtypes.MsgEthereumTx:
			gasUsed += tx.GetGas()
		case sdk.FeeTx:
			gasUsed += tx.GetGas()
		}
	}

	return gasUsed
}

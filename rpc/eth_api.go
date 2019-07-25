package rpc

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/signer/core"
	"math/big"
)

// PublicEthAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthAPI struct {
	cliCtx context.CLIContext
}

// NewPublicEthAPI creates an instance of the public ETH Web3 API.
func NewPublicEthAPI(cliCtx context.CLIContext) *PublicEthAPI {
	return &PublicEthAPI{
		cliCtx: cliCtx,
	}
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicEthAPI) ProtocolVersion() string {
	return version.ProtocolVersion
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (e *PublicEthAPI) Syncing() interface{} {
	return false
}

// Coinbase returns this node's coinbase address. Not used in Ethermint.
func (e *PublicEthAPI) Coinbase() (addr common.Address) {
	return
}

// Mining returns whether or not this node is currently mining. Always false.
func (e *PublicEthAPI) Mining() bool {
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (e *PublicEthAPI) Hashrate() hexutil.Uint64 {
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (e *PublicEthAPI) GasPrice() *hexutil.Big {
	out := big.NewInt(0)
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (e *PublicEthAPI) Accounts() []common.Address {
	return nil
}

// BlockNumber returns the current block number.
func (e *PublicEthAPI) BlockNumber() *big.Int {
	res, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/blockNumber", types.ModuleName), nil)
	if err != nil {
		fmt.Printf("could not resolve: %s\n", err)
		return nil
	}

	var out types.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Number
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicEthAPI) GetBalance(address common.Address, blockNum rpc.BlockNumber) *hexutil.Big {
	res, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/balance/%s", types.ModuleName, address), nil)
	if err != nil {
		fmt.Printf("could not resolve: %s\n", err)
		return nil
	}

	var out types.QueryResBalance
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return (*hexutil.Big)(out.Balance)
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicEthAPI) GetStorageAt(address common.Address, key string, blockNum rpc.BlockNumber) hexutil.Bytes {
	res, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/storage/%s/%s", types.ModuleName, address, key), nil)
	if err != nil {
		fmt.Printf("could not resolve: %s\n", err)
		return nil
	}

	var out types.QueryResStorage
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Value[:]
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicEthAPI) GetTransactionCount(address common.Address, blockNum rpc.BlockNumber) hexutil.Uint64 {
	return 0
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicEthAPI) GetBlockTransactionCountByHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicEthAPI) GetBlockTransactionCountByNumber(blockNum rpc.BlockNumber) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockHash returns the number of uncles in the block idenfied by hash. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block idenfied by number. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockNumber(blockNum rpc.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicEthAPI) GetCode(address common.Address, blockNumber rpc.BlockNumber) hexutil.Bytes {
	res, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/code/%s", types.ModuleName, address), nil)
	if err != nil {
		fmt.Printf("could not resolve: %s\n", err)
		return nil
	}

	var out types.QueryResCode
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Code
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicEthAPI) Sign(address common.Address, data hexutil.Bytes) hexutil.Bytes {
	return nil
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicEthAPI) SendTransaction(args core.SendTxArgs) common.Hash {
	var h common.Hash
	return h
}

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicEthAPI) SendRawTransaction(data hexutil.Bytes) common.Hash {
	var h common.Hash
	return h
}

// CallArgs represents arguments to a smart contract call as provided by RPC clients.
type CallArgs struct {
	From     common.Address `json:"from"`
	To       common.Address `json:"to"`
	Gas      hexutil.Uint64 `json:"gas"`
	GasPrice hexutil.Big    `json:"gasPrice"`
	Value    hexutil.Big    `json:"value"`
	Data     hexutil.Bytes  `json:"data"`
}

// Call performs a raw contract call.
func (e *PublicEthAPI) Call(args CallArgs, blockNum rpc.BlockNumber) hexutil.Bytes {
	return nil
}

// EstimateGas estimates gas usage for the given smart contract call.
func (e *PublicEthAPI) EstimateGas(args CallArgs, blockNum rpc.BlockNumber) hexutil.Uint64 {
	return 0
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicEthAPI) GetBlockByHash(hash common.Hash, fullTx bool) map[string]interface{} {
	return nil
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicEthAPI) GetBlockByNumber(blockNum rpc.BlockNumber, fullTx bool) map[string]interface{} {
	return nil
}

// Transaction represents a transaction returned to RPC clients.
type Transaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicEthAPI) GetTransactionByHash(hash common.Hash) *Transaction {
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicEthAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) *Transaction {
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicEthAPI) GetTransactionByBlockNumberAndIndex(blockNumber rpc.BlockNumber, idx hexutil.Uint) *Transaction {
	return nil
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicEthAPI) GetTransactionReceipt(hash common.Hash) map[string]interface{} {
	return nil
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

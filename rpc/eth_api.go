package rpc

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"

	emintcrypto "github.com/cosmos/ethermint/crypto"
	emintkeys "github.com/cosmos/ethermint/keys"
	"github.com/cosmos/ethermint/rpc/args"
	"github.com/cosmos/ethermint/utils"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm"
	"github.com/cosmos/ethermint/x/evm/types"

	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authutils "github.com/cosmos/cosmos-sdk/x/auth/client/utils"

	"github.com/spf13/viper"
)

// PublicEthAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthAPI struct {
	cliCtx    context.CLIContext
	key       emintcrypto.PrivKeySecp256k1
	nonceLock *AddrLocker
	gasLimit  *int64
}

// NewPublicEthAPI creates an instance of the public ETH Web3 API.
func NewPublicEthAPI(cliCtx context.CLIContext, nonceLock *AddrLocker,
	key emintcrypto.PrivKeySecp256k1) *PublicEthAPI {

	return &PublicEthAPI{
		cliCtx:    cliCtx,
		key:       key,
		nonceLock: nonceLock,
	}
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicEthAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(version.ProtocolVersion)
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (e *PublicEthAPI) Syncing() (interface{}, error) {
	status, err := e.cliCtx.Client.Status()
	if err != nil {
		return false, err
	}

	if !status.SyncInfo.CatchingUp {
		return false, nil
	}

	return map[string]interface{}{
		// "startingBlock": nil, // NA
		"currentBlock": hexutil.Uint64(status.SyncInfo.LatestBlockHeight),
		// "highestBlock":  nil, // NA
		// "pulledStates":  nil, // NA
		// "knownStates":   nil, // NA
	}, nil
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
func (e *PublicEthAPI) Accounts() ([]common.Address, error) {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	keybase, err := emintkeys.NewKeyBaseFromHomeFlag()
	if err != nil {
		return addresses, err
	}

	infos, err := keybase.List()
	if err != nil {
		return addresses, err
	}

	for _, info := range infos {
		addressBytes := info.GetPubKey().Address().Bytes()
		addresses = append(addresses, common.BytesToAddress(addressBytes))
	}

	return addresses, nil
}

// BlockNumber returns the current block number.
func (e *PublicEthAPI) BlockNumber() (hexutil.Uint64, error) {
	res, _, err := e.cliCtx.QueryWithData(fmt.Sprintf("custom/%s/blockNumber", types.ModuleName), nil)
	if err != nil {
		return hexutil.Uint64(0), err
	}

	var out types.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return hexutil.Uint64(out.Number), nil
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicEthAPI) GetBalance(address common.Address, blockNum BlockNumber) (*hexutil.Big, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/balance/%s", types.ModuleName, address.Hex()), nil)
	if err != nil {
		return nil, err
	}

	var out types.QueryResBalance
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	val, err := utils.UnmarshalBigInt(out.Balance)
	if err != nil {
		return nil, err
	}

	return (*hexutil.Big)(val), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicEthAPI) GetStorageAt(address common.Address, key string, blockNum BlockNumber) (hexutil.Bytes, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/storage/%s/%s", types.ModuleName, address.Hex(), key), nil)
	if err != nil {
		return nil, err
	}

	var out types.QueryResStorage
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Value, nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicEthAPI) GetTransactionCount(address common.Address, blockNum BlockNumber) (*hexutil.Uint64, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/nonce/%s", types.ModuleName, address.Hex()), nil)
	if err != nil {
		return nil, err
	}

	var out types.QueryResNonce
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return (*hexutil.Uint64)(&out.Nonce), nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicEthAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryHashToHeight, hash.Hex()))
	if err != nil {
		// Return nil if block does not exist
		return nil
	}

	var out types.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return e.getBlockTransactionCountByNumber(out.Number)
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicEthAPI) GetBlockTransactionCountByNumber(blockNum BlockNumber) *hexutil.Uint {
	height := blockNum.Int64()
	return e.getBlockTransactionCountByNumber(height)
}

func (e *PublicEthAPI) getBlockTransactionCountByNumber(number int64) *hexutil.Uint {
	block, err := e.cliCtx.Client.Block(&number)
	if err != nil {
		// Return nil if block doesn't exist
		return nil
	}

	n := hexutil.Uint(block.Block.NumTxs)
	return &n
}

// GetUncleCountByBlockHash returns the number of uncles in the block idenfied by hash. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block idenfied by number. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockNumber(blockNum BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicEthAPI) GetCode(address common.Address, blockNumber BlockNumber) (hexutil.Bytes, error) {
	ctx := e.cliCtx.WithHeight(blockNumber.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/code/%s", types.ModuleName, address.Hex()), nil)
	if err != nil {
		return nil, err
	}

	var out types.QueryResCode
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Code, nil
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicEthAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	// TODO: Change this functionality to find an unlocked account by address
	if e.key == nil || !bytes.Equal(e.key.PubKey().Address().Bytes(), address.Bytes()) {
		return nil, keystore.ErrLocked
	}

	// Sign the requested hash with the wallet
	signature, err := e.key.Sign(data)
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}

	return signature, err
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicEthAPI) SendTransaction(args args.SendTxArgs) (common.Hash, error) {
	// TODO: Change this functionality to find an unlocked account by address
	if e.key == nil || !bytes.Equal(e.key.PubKey().Address().Bytes(), args.From.Bytes()) {
		return common.Hash{}, keystore.ErrLocked
	}

	// Mutex lock the address' nonce to avoid assigning it to multiple requests
	if args.Nonce == nil {
		e.nonceLock.LockAddr(args.From)
		defer e.nonceLock.UnlockAddr(args.From)
	}

	// Assemble transaction from fields
	tx, err := types.GenerateFromArgs(args, e.cliCtx)
	if err != nil {
		return common.Hash{}, err
	}

	// ChainID must be set as flag to send transaction
	chainID := viper.GetString(flags.FlagChainID)
	// parse the chainID from a string to a base-10 integer
	intChainID, ok := new(big.Int).SetString(chainID, 10)
	if !ok {
		return common.Hash{}, fmt.Errorf(
			fmt.Sprintf("Invalid chainID: %s, must be integer format", chainID))
	}

	// Sign transaction
	tx.Sign(intChainID, e.key.ToECDSA())

	// Encode transaction by default Tx encoder
	txEncoder := authutils.GetTxEncoder(e.cliCtx.Codec)
	txBytes, err := txEncoder(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// Broadcast transaction
	res, err := e.cliCtx.BroadcastTx(txBytes)
	// If error is encountered on the node, the broadcast will not return an error
	// TODO: Remove res log
	fmt.Println(res.RawLog)
	if err != nil {
		return common.Hash{}, err
	}

	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicEthAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	tx := new(types.EthereumTxMsg)

	// RLP decode raw transaction bytes
	if err := rlp.DecodeBytes(data, tx); err != nil {
		// Return nil is for when gasLimit overflows uint64
		return common.Hash{}, nil
	}

	// Encode transaction by default Tx encoder
	txEncoder := authutils.GetTxEncoder(e.cliCtx.Codec)
	txBytes, err := txEncoder(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// TODO: Possibly log the contract creation address (if recipient address is nil) or tx data
	res, err := e.cliCtx.BroadcastTx(txBytes)
	// If error is encountered on the node, the broadcast will not return an error
	// TODO: Remove res log
	fmt.Println(res.RawLog)
	if err != nil {
		return common.Hash{}, err
	}

	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
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
func (e *PublicEthAPI) Call(args CallArgs, blockNum BlockNumber) hexutil.Bytes {
	return nil
}

// EstimateGas estimates gas usage for the given smart contract call.
func (e *PublicEthAPI) EstimateGas(args CallArgs, blockNum BlockNumber) hexutil.Uint64 {
	return 0
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicEthAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryHashToHeight, hash.Hex()))
	if err != nil {
		return nil, err
	}

	var out types.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return e.getEthBlockByNumber(out.Number, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicEthAPI) GetBlockByNumber(blockNum BlockNumber, fullTx bool) (map[string]interface{}, error) {
	value := blockNum.Int64()
	return e.getEthBlockByNumber(value, fullTx)
}

func (e *PublicEthAPI) getEthBlockByNumber(value int64, fullTx bool) (map[string]interface{}, error) {
	// Remove this check when 0 query is fixed ref: (https://github.com/tendermint/tendermint/issues/4014)
	var blkNumPtr *int64
	if value != 0 {
		blkNumPtr = &value
	}

	block, err := e.cliCtx.Client.Block(blkNumPtr)
	if err != nil {
		return nil, err
	}
	header := block.BlockMeta.Header

	gasLimit, err := e.getGasLimit()
	if err != nil {
		return nil, err
	}

	var gasUsed *big.Int
	var transactions []interface{}

	if fullTx {
		// Populate full transaction data
		transactions, gasUsed = convertTransactionsToRPC(e.cliCtx, block.Block.Txs,
			common.BytesToHash(header.Hash()), uint64(header.Height))
	} else {
		// TODO: Gas used not saved and cannot be calculated by hashes
		// Return slice of transaction hashes
		transactions = make([]interface{}, len(block.Block.Txs))
		for i, tx := range block.Block.Txs {
			transactions[i] = common.BytesToHash(tx.Hash())
		}
	}

	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryLogsBloom, strconv.FormatInt(block.Block.Height, 10)))
	if err != nil {
		return nil, err
	}

	var out types.QueryBloomFilter
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)

	return formatBlock(header, block.Block.Size(), gasLimit, gasUsed, transactions, out.Bloom), nil
}

func formatBlock(
	header tmtypes.Header, size int, gasLimit int64,
	gasUsed *big.Int, transactions []interface{}, bloom ethtypes.Bloom,
) map[string]interface{} {
	return map[string]interface{}{
		"number":           hexutil.Uint64(header.Height),
		"hash":             hexutil.Bytes(header.Hash()),
		"parentHash":       hexutil.Bytes(header.LastBlockID.Hash),
		"nonce":            nil, // PoW specific
		"sha3Uncles":       nil, // No uncles in Tendermint
		"logsBloom":        bloom,
		"transactionsRoot": hexutil.Bytes(header.DataHash),
		"stateRoot":        hexutil.Bytes(header.AppHash),
		"miner":            hexutil.Bytes(header.ValidatorsHash),
		"difficulty":       nil,
		"totalDifficulty":  nil,
		"extraData":        nil,
		"size":             hexutil.Uint64(size),
		"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
		"gasUsed":          (*hexutil.Big)(gasUsed),
		"timestamp":        hexutil.Uint64(header.Time.Unix()),
		"transactions":     transactions,
		"uncles":           nil,
	}
}

func convertTransactionsToRPC(cliCtx context.CLIContext, txs []tmtypes.Tx, blockHash common.Hash, height uint64) ([]interface{}, *big.Int) {
	transactions := make([]interface{}, len(txs))
	gasUsed := big.NewInt(0)
	for i, tx := range txs {
		ethTx, err := bytesToEthTx(cliCtx, tx)
		if err != nil {
			continue
		}
		// TODO: Remove gas usage calculation if saving gasUsed per block
		gasUsed.Add(gasUsed, ethTx.Fee())
		transactions[i] = newRPCTransaction(ethTx, blockHash, &height, uint64(i))
	}
	return transactions, gasUsed
}

// Transaction represents a transaction returned to RPC clients.
type Transaction struct {
	BlockHash        *common.Hash    `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex *hexutil.Uint64 `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

func bytesToEthTx(cliCtx context.CLIContext, bz []byte) (*types.EthereumTxMsg, error) {
	var stdTx sdk.Tx
	err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(bz, &stdTx)
	ethTx, ok := stdTx.(*types.EthereumTxMsg)
	if !ok || err != nil {
		return nil, fmt.Errorf("Invalid transaction type, must be an amino encoded Ethereum transaction")
	}
	return ethTx, nil
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.EthereumTxMsg, blockHash common.Hash, blockNumber *uint64, index uint64) *Transaction {
	// Verify signature and retrieve sender address
	from, _ := tx.VerifySig(tx.ChainID())

	result := &Transaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Data.GasLimit),
		GasPrice: (*hexutil.Big)(tx.Data.Price),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data.Payload),
		Nonce:    hexutil.Uint64(tx.Data.AccountNonce),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Data.Amount),
		V:        (*hexutil.Big)(tx.Data.V),
		R:        (*hexutil.Big)(tx.Data.R),
		S:        (*hexutil.Big)(tx.Data.S),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(*blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}
	return result
}

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicEthAPI) GetTransactionByHash(hash common.Hash) (*Transaction, error) {
	tx, err := e.cliCtx.Client.Tx(hash.Bytes(), false)
	if err != nil {
		// Return nil for transaction when not found
		return nil, nil
	}

	// Can either cache or just leave this out if not necessary
	block, err := e.cliCtx.Client.Block(&tx.Height)
	if err != nil {
		return nil, err
	}
	blockHash := common.BytesToHash(block.BlockMeta.Header.Hash())

	ethTx, err := bytesToEthTx(e.cliCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	height := uint64(tx.Height)
	return newRPCTransaction(ethTx, blockHash, &height, uint64(tx.Index)), nil
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicEthAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*Transaction, error) {
	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryHashToHeight, hash.Hex()))
	if err != nil {
		return nil, err
	}

	var out types.QueryResBlockNumber
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return e.getTransactionByBlockNumberAndIndex(out.Number, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicEthAPI) GetTransactionByBlockNumberAndIndex(blockNum BlockNumber, idx hexutil.Uint) (*Transaction, error) {
	value := blockNum.Int64()
	return e.getTransactionByBlockNumberAndIndex(value, idx)
}

func (e *PublicEthAPI) getTransactionByBlockNumberAndIndex(number int64, idx hexutil.Uint) (*Transaction, error) {
	block, err := e.cliCtx.Client.Block(&number)
	if err != nil {
		return nil, err
	}
	header := block.BlockMeta.Header

	txs := block.Block.Txs
	if uint64(idx) >= uint64(len(txs)) {
		return nil, nil
	}
	ethTx, err := bytesToEthTx(e.cliCtx, txs[idx])
	if err != nil {
		return nil, err
	}

	height := uint64(header.Height)
	transaction := newRPCTransaction(ethTx, common.BytesToHash(header.Hash()), &height, uint64(idx))
	return transaction, nil
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicEthAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	tx, err := e.cliCtx.Client.Tx(hash.Bytes(), false)
	if err != nil {
		// Return nil for transaction when not found
		return nil, nil
	}

	// Query block for consensus hash
	block, err := e.cliCtx.Client.Block(&tx.Height)
	if err != nil {
		return nil, err
	}
	blockHash := common.BytesToHash(block.BlockMeta.Header.Hash())

	// Convert tx bytes to eth transaction
	ethTx, err := bytesToEthTx(e.cliCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	from, _ := ethTx.VerifySig(ethTx.ChainID())

	// Set status codes based on tx result
	var status hexutil.Uint
	if tx.TxResult.IsOK() {
		status = hexutil.Uint(1)
	} else {
		status = hexutil.Uint(0)
	}

	res, _, err := e.cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryTxLogs, hash.Hex()))
	if err != nil {
		return nil, err
	}

	var logs types.QueryETHLogs
	e.cliCtx.Codec.MustUnmarshalJSON(res, &logs)

	txData := tx.TxResult.GetData()
	var bloomFilter ethtypes.Bloom
	var contractAddress common.Address
	if len(txData) >= 20 {
		// TODO: change hard coded indexing of bytes
		bloomFilter = ethtypes.BytesToBloom(txData[20:])
		contractAddress = common.BytesToAddress(txData[:20])
	}

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(tx.Height),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(tx.Index),
		"from":              from,
		"to":                ethTx.To(),
		"gasUsed":           hexutil.Uint64(tx.TxResult.GasUsed),
		"cumulativeGasUsed": nil, // ignore until needed
		"contractAddress":   nil,
		"logs":              logs.Logs,
		"logsBloom":         bloomFilter,
		"status":            status,
	}

	if contractAddress != (common.Address{}) {
		// TODO: change hard coded indexing of first 20 bytes
		fields["contractAddress"] = contractAddress
	}

	return fields, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *PublicEthAPI) PendingTransactions() ([]*Transaction, error) {
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
		rpcTx := newRPCTransaction(ethTx, common.Hash{}, nil, 0)
		transactions = append(transactions, rpcTx)
	}

	return transactions, nil
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// AccountResult struct for account proof
type AccountResult struct {
	Address      common.Address  `json:"address"`
	AccountProof []string        `json:"accountProof"`
	Balance      *hexutil.Big    `json:"balance"`
	CodeHash     common.Hash     `json:"codeHash"`
	Nonce        hexutil.Uint64  `json:"nonce"`
	StorageHash  common.Hash     `json:"storageHash"`
	StorageProof []StorageResult `json:"storageProof"`
}

// StorageResult defines the format for storage proof return
type StorageResult struct {
	Key   string       `json:"key"`
	Value *hexutil.Big `json:"value"`
	Proof []string     `json:"proof"`
}

// GetProof returns an account object with proof and any storage proofs
func (e *PublicEthAPI) GetProof(address common.Address, storageKeys []string, block BlockNumber) (*AccountResult, error) {
	opts := client.ABCIQueryOptions{Height: int64(block), Prove: true}
	path := fmt.Sprintf("custom/%s/%s/%s", types.ModuleName, evm.QueryAccount, address.Hex())
	pRes, err := e.cliCtx.Client.ABCIQueryWithOptions(path, nil, opts)
	if err != nil {
		return nil, err
	}

	// TODO: convert TM merkle proof to []string if needed in future
	// proof := pRes.Response.GetProof()

	account := new(types.QueryAccount)
	e.cliCtx.Codec.MustUnmarshalJSON(pRes.Response.GetValue(), &account)

	storageProofs := make([]StorageResult, len(storageKeys))
	for i, k := range storageKeys {
		// Get value for key
		vPath := fmt.Sprintf("custom/%s/%s/%s/%s", types.ModuleName, evm.QueryStorage, address, k)
		vRes, err := e.cliCtx.Client.ABCIQueryWithOptions(vPath, nil, opts)
		if err != nil {
			return nil, err
		}
		value := new(types.QueryResStorage)
		e.cliCtx.Codec.MustUnmarshalJSON(vRes.Response.GetValue(), &value)

		storageProofs[i] = StorageResult{
			Key:   k,
			Value: (*hexutil.Big)(common.BytesToHash(value.Value).Big()),
			Proof: []string{""},
		}
	}

	return &AccountResult{
		Address:      address,
		AccountProof: []string{""}, // This shouldn't be necessary (different proof formats)
		Balance:      (*hexutil.Big)(utils.MustUnmarshalBigInt(account.Balance)),
		CodeHash:     common.BytesToHash(account.CodeHash),
		Nonce:        hexutil.Uint64(account.Nonce),
		StorageHash:  common.Hash{}, // Ethermint doesn't have a storage hash
		StorageProof: storageProofs,
	}, nil
}

// getGasLimit returns the gas limit per block set in genesis
func (e *PublicEthAPI) getGasLimit() (int64, error) {
	// Retrieve from gasLimit variable cache
	if e.gasLimit != nil {
		return *e.gasLimit, nil
	}

	// Query genesis block if hasn't been retrieved yet
	genesis, err := e.cliCtx.Client.Genesis()
	if err != nil {
		return 0, err
	}

	// Save value to gasLimit cached value
	gasLimit := genesis.Genesis.ConsensusParams.Block.MaxGas
	e.gasLimit = &gasLimit
	return gasLimit, nil
}

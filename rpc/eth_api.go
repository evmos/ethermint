package rpc

import (
	"bytes"
	"fmt"
	"math/big"

	emintcrypto "github.com/cosmos/ethermint/crypto"
	emintkeys "github.com/cosmos/ethermint/keys"
	"github.com/cosmos/ethermint/rpc/args"
	"github.com/cosmos/ethermint/version"
	"github.com/cosmos/ethermint/x/evm/types"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	var qRes uint64
	e.cliCtx.Codec.MustUnmarshalJSON(res, &qRes)

	return hexutil.Uint64(qRes), nil
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicEthAPI) GetBalance(address common.Address, blockNum BlockNumber) (*hexutil.Big, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/balance/%s", types.ModuleName, address), nil)
	if err != nil {
		return nil, err
	}

	var out *big.Int
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)

	return (*hexutil.Big)(out), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicEthAPI) GetStorageAt(address common.Address, key string, blockNum BlockNumber) (hexutil.Bytes, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/storage/%s/%s", types.ModuleName, address, key), nil)
	if err != nil {
		return nil, err
	}

	var out []byte
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out, nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicEthAPI) GetTransactionCount(address common.Address, blockNum BlockNumber) (*hexutil.Uint64, error) {
	ctx := e.cliCtx.WithHeight(blockNum.Int64())
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/nonce/%s", types.ModuleName, address), nil)
	if err != nil {
		return nil, err
	}

	var out *hexutil.Uint64
	e.cliCtx.Codec.MustUnmarshalJSON(res, out)
	return out, nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicEthAPI) GetBlockTransactionCountByHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicEthAPI) GetBlockTransactionCountByNumber(blockNum BlockNumber) (hexutil.Uint, error) {
	node, err := e.cliCtx.GetNode()
	if err != nil {
		return 0, err
	}

	height := blockNum.Int64()
	block, err := node.Block(&height)
	if err != nil {
		return 0, err
	}

	return hexutil.Uint(block.Block.NumTxs), nil
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
	res, _, err := ctx.QueryWithData(fmt.Sprintf("custom/%s/code/%s", types.ModuleName, address), nil)
	if err != nil {
		return nil, err
	}

	var out []byte
	e.cliCtx.Codec.MustUnmarshalJSON(res, &out)
	return out, nil
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
func (e *PublicEthAPI) GetBlockByHash(hash common.Hash, fullTx bool) map[string]interface{} {
	return nil
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicEthAPI) GetBlockByNumber(blockNum BlockNumber, fullTx bool) (map[string]interface{}, error) {
	value := blockNum.Int64()

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
			common.BytesToHash(header.ConsensusHash.Bytes()), uint64(header.Height))
	} else {
		// TODO: Gas used not saved and cannot be calculated by hashes
		// Return slice of transaction hashes
		transactions = make([]interface{}, len(block.Block.Txs))
		for i, tx := range block.Block.Txs {
			transactions[i] = common.BytesToHash(tx.Hash())
		}
	}

	return formatBlock(header, block.Block.Size(), gasLimit, gasUsed, transactions), nil
}

func formatBlock(
	header tmtypes.Header, size int, gasLimit int64,
	gasUsed *big.Int, transactions []interface{},
) map[string]interface{} {
	return map[string]interface{}{
		"number":           hexutil.Uint64(header.Height),
		"hash":             hexutil.Bytes(header.ConsensusHash),
		"parentHash":       hexutil.Bytes(header.LastBlockID.Hash),
		"nonce":            nil, // PoW specific
		"sha3Uncles":       nil, // No uncles in Tendermint
		"logsBloom":        "",  // TODO: Complete with #55
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
		var stdTx sdk.Tx
		err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(tx, &stdTx)
		ethTx, ok := stdTx.(*types.EthereumTxMsg)
		if !ok || err != nil {
			continue
		}
		// TODO: Remove gas usage calculation if saving gasUsed per block
		gasUsed.Add(gasUsed, ethTx.Fee())
		transactions[i] = newRPCTransaction(ethTx, blockHash, height, uint64(i))
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

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.EthereumTxMsg, blockHash common.Hash, blockNumber uint64, index uint64) *Transaction {
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
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}
	return result
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
func (e *PublicEthAPI) GetTransactionByBlockNumberAndIndex(blockNumber BlockNumber, idx hexutil.Uint) *Transaction {
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

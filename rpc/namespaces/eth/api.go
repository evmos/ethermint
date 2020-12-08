package eth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/spf13/viper"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/crypto/hd"
	"github.com/cosmos/ethermint/rpc/backend"
	rpctypes "github.com/cosmos/ethermint/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/utils"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	clientcontext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// PublicEthereumAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthereumAPI struct {
	ctx          context.Context
	clientCtx    clientcontext.CLIContext
	chainIDEpoch *big.Int
	logger       log.Logger
	backend      backend.Backend
	keys         []ethsecp256k1.PrivKey // unlocked keys
	nonceLock    *rpctypes.AddrLocker
	keyringLock  sync.Mutex
}

// NewAPI creates an instance of the public ETH Web3 API.
func NewAPI(
	clientCtx clientcontext.CLIContext, backend backend.Backend, nonceLock *rpctypes.AddrLocker,
	keys ...ethsecp256k1.PrivKey,
) *PublicEthereumAPI {

	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	api := &PublicEthereumAPI{
		ctx:          context.Background(),
		clientCtx:    clientCtx,
		chainIDEpoch: epoch,
		logger:       log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "json-rpc", "namespace", "eth"),
		backend:      backend,
		keys:         keys,
		nonceLock:    nonceLock,
	}

	if err := api.GetKeyringInfo(); err != nil {
		api.logger.Error("failed to get keybase info", "error", err)
	}

	return api
}

// GetKeyringInfo checks if the keyring is present on the client context. If not, it creates a new
// instance and sets it to the client context for later usage.
func (api *PublicEthereumAPI) GetKeyringInfo() error {
	api.keyringLock.Lock()
	defer api.keyringLock.Unlock()

	if api.clientCtx.Keybase != nil {
		return nil
	}

	keybase, err := keys.NewKeyring(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		api.clientCtx.Input,
		hd.EthSecp256k1Options()...,
	)
	if err != nil {
		return err
	}

	api.clientCtx.Keybase = keybase
	return nil
}

// ClientCtx returns the Cosmos SDK client context.
func (api *PublicEthereumAPI) ClientCtx() clientcontext.CLIContext {
	return api.clientCtx
}

// GetKeys returns the Cosmos SDK client context.
func (api *PublicEthereumAPI) GetKeys() []ethsecp256k1.PrivKey {
	return api.keys
}

// SetKeys sets the given key slice to the set of private keys
func (api *PublicEthereumAPI) SetKeys(keys []ethsecp256k1.PrivKey) {
	api.keys = keys
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (api *PublicEthereumAPI) ProtocolVersion() hexutil.Uint {
	api.logger.Debug("eth_protocolVersion")
	return hexutil.Uint(ethermint.ProtocolVersion)
}

// ChainId returns the chain's identifier in hex format
func (api *PublicEthereumAPI) ChainId() (hexutil.Uint, error) { // nolint
	api.logger.Debug("eth_chainId")
	return hexutil.Uint(uint(api.chainIDEpoch.Uint64())), nil
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (api *PublicEthereumAPI) Syncing() (interface{}, error) {
	api.logger.Debug("eth_syncing")

	status, err := api.clientCtx.Client.Status()
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

// Coinbase is the address that staking rewards will be send to (alias for Etherbase).
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	api.logger.Debug("eth_coinbase")

	node, err := api.clientCtx.GetNode()
	if err != nil {
		return common.Address{}, err
	}

	status, err := node.Status()
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(status.ValidatorInfo.Address.Bytes()), nil
}

// Mining returns whether or not this node is currently mining. Always false.
func (api *PublicEthereumAPI) Mining() bool {
	api.logger.Debug("eth_mining")
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	api.logger.Debug("eth_hashrate")
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (api *PublicEthereumAPI) GasPrice() *hexutil.Big {
	api.logger.Debug("eth_gasPrice")
	out := big.NewInt(0)
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (api *PublicEthereumAPI) Accounts() ([]common.Address, error) {
	api.logger.Debug("eth_accounts")
	api.keyringLock.Lock()
	defer api.keyringLock.Unlock()

	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := api.clientCtx.Keybase.List()
	if err != nil {
		return addresses, err
	}

	for _, info := range infos {
		addressBytes := info.GetPubKey().Address().Bytes()
		addresses = append(addresses, common.BytesToAddress(addressBytes))
	}

	return addresses, nil
}

// rpctypes.BlockNumber returns the current block number.
func (api *PublicEthereumAPI) BlockNumber() (hexutil.Uint64, error) {
	api.logger.Debug("eth_blockNumber")
	return api.backend.BlockNumber()
}

// GetBalance returns the provided account's balance up to the provided block number.
func (api *PublicEthereumAPI) GetBalance(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Big, error) {
	api.logger.Debug("eth_getBalance", "address", address, "block number", blockNum)
	clientCtx := api.clientCtx.WithHeight(blockNum.Int64())
	res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/balance/%s", evmtypes.ModuleName, address.Hex()), nil)
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryResBalance
	api.clientCtx.Codec.MustUnmarshalJSON(res, &out)
	val, err := utils.UnmarshalBigInt(out.Balance)
	if err != nil {
		return nil, err
	}

	return (*hexutil.Big)(val), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (api *PublicEthereumAPI) GetStorageAt(address common.Address, key string, blockNum rpctypes.BlockNumber) (hexutil.Bytes, error) {
	api.logger.Debug("eth_getStorageAt", "address", address, "key", key, "block number", blockNum)
	clientCtx := api.clientCtx.WithHeight(blockNum.Int64())
	res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/storage/%s/%s", evmtypes.ModuleName, address.Hex(), key), nil)
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryResStorage
	api.clientCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Value, nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (api *PublicEthereumAPI) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	api.logger.Debug("eth_getTransactionCount", "address", address, "block number", blockNum)
	clientCtx := api.clientCtx.WithHeight(blockNum.Int64())

	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := authtypes.NewAccountRetriever(clientCtx)

	err := accRet.EnsureExists(from)
	if err != nil {
		// account doesn't exist yet, return 0
		n := hexutil.Uint64(0)
		return &n, nil
	}

	_, nonce, err := accRet.GetAccountNumberSequence(from)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (api *PublicEthereumAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	api.logger.Debug("eth_getBlockTransactionCountByHash", "hash", hash)
	res, _, err := api.clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryHashToHeight, hash.Hex()))
	if err != nil {
		return nil
	}

	var out evmtypes.QueryResBlockNumber
	if err := api.clientCtx.Codec.UnmarshalJSON(res, &out); err != nil {
		return nil
	}

	resBlock, err := api.clientCtx.Client.Block(&out.Number)
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (api *PublicEthereumAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	api.logger.Debug("eth_getBlockTransactionCountByNumber", "block number", blockNum)
	height := blockNum.Int64()
	resBlock, err := api.clientCtx.Client.Block(&height)
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetUncleCountByBlockHash returns the number of uncles in the block idenfied by hash. Always zero.
func (api *PublicEthereumAPI) GetUncleCountByBlockHash(_ common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block idenfied by number. Always zero.
func (api *PublicEthereumAPI) GetUncleCountByBlockNumber(_ rpctypes.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (api *PublicEthereumAPI) GetCode(address common.Address, blockNumber rpctypes.BlockNumber) (hexutil.Bytes, error) {
	api.logger.Debug("eth_getCode", "address", address, "block number", blockNumber)
	clientCtx := api.clientCtx.WithHeight(blockNumber.Int64())
	res, _, err := clientCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryCode, address.Hex()), nil)
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryResCode
	api.clientCtx.Codec.MustUnmarshalJSON(res, &out)
	return out.Code, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (api *PublicEthereumAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	api.logger.Debug("eth_getTransactionLogs", "hash", txHash)
	return api.backend.GetTransactionLogs(txHash)
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (api *PublicEthereumAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	api.logger.Debug("eth_sign", "address", address, "data", data)
	// TODO: Change this functionality to find an unlocked account by address

	key, exist := rpctypes.GetKeyByAddress(api.keys, address)
	if !exist {
		return nil, keystore.ErrLocked
	}

	// Sign the requested hash with the wallet
	signature, err := key.Sign(data)
	if err != nil {
		return nil, err
	}

	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SendTransaction sends an Ethereum transaction.
func (api *PublicEthereumAPI) SendTransaction(args rpctypes.SendTxArgs) (common.Hash, error) {
	api.logger.Debug("eth_sendTransaction", "args", args)
	// TODO: Change this functionality to find an unlocked account by address

	key, exist := rpctypes.GetKeyByAddress(api.keys, args.From)
	if !exist {
		api.logger.Debug("failed to find key in keyring", "key", args.From)
		return common.Hash{}, keystore.ErrLocked
	}

	// Mutex lock the address' nonce to avoid assigning it to multiple requests
	if args.Nonce == nil {
		api.nonceLock.LockAddr(args.From)
		defer api.nonceLock.UnlockAddr(args.From)
	}

	// Assemble transaction from fields
	tx, err := api.generateFromArgs(args)
	if err != nil {
		api.logger.Debug("failed to generate tx", "error", err)
		return common.Hash{}, err
	}

	// Sign transaction
	if err := tx.Sign(api.chainIDEpoch, key.ToECDSA()); err != nil {
		api.logger.Debug("failed to sign tx", "error", err)
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txEncoder := authclient.GetTxEncoder(api.clientCtx.Codec)
	txBytes, err := txEncoder(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	res, err := api.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return common.Hash{}, err
	}

	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (api *PublicEthereumAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	api.logger.Debug("eth_sendRawTransaction", "data", data)
	tx := new(evmtypes.MsgEthereumTx)

	// RLP decode raw transaction bytes
	if err := rlp.DecodeBytes(data, tx); err != nil {
		// Return nil is for when gasLimit overflows uint64
		return common.Hash{}, nil
	}

	// Encode transaction by default Tx encoder
	txEncoder := authclient.GetTxEncoder(api.clientCtx.Codec)
	txBytes, err := txEncoder(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// TODO: Possibly log the contract creation address (if recipient address is nil) or tx data
	// If error is encountered on the node, the broadcast will not return an error
	res, err := api.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return common.Hash{}, err
	}

	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
}

// Call performs a raw contract call.
func (api *PublicEthereumAPI) Call(args rpctypes.CallArgs, blockNr rpctypes.BlockNumber, _ *map[common.Address]rpctypes.Account) (hexutil.Bytes, error) {
	api.logger.Debug("eth_call", "args", args, "block number", blockNr)
	simRes, err := api.doCall(args, blockNr, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return []byte{}, err
	}

	data, err := evmtypes.DecodeResultData(simRes.Result.Data)
	if err != nil {
		return []byte{}, err
	}

	return (hexutil.Bytes)(data.Ret), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (api *PublicEthereumAPI) doCall(
	args rpctypes.CallArgs, blockNr rpctypes.BlockNumber, globalGasCap *big.Int,
) (*sdk.SimulationResponse, error) {
	// Set height for historical queries
	clientCtx := api.clientCtx

	if blockNr.Int64() != 0 {
		clientCtx = api.clientCtx.WithHeight(blockNr.Int64())
	}

	// Set sender address or use a default if none specified
	var addr common.Address

	if args.From == nil {
		addrs, err := api.Accounts()
		if err == nil && len(addrs) > 0 {
			addr = addrs[0]
		}
	} else {
		addr = *args.From
	}

	// Set default gas & gas price if none were set
	// Change this to uint64(math.MaxUint64 / 2) if gas cap can be configured
	gas := uint64(ethermint.DefaultRPCGasLimit)
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != nil && globalGasCap.Uint64() < gas {
		api.logger.Debug("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
		gas = globalGasCap.Uint64()
	}

	// Set gas price using default or parameter if passed in
	gasPrice := new(big.Int).SetUint64(ethermint.DefaultGasPrice)
	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	// Set value for transaction
	value := new(big.Int)
	if args.Value != nil {
		value = args.Value.ToInt()
	}

	// Set Data if provided
	var data []byte
	if args.Data != nil {
		data = []byte(*args.Data)
	}

	// Set destination address for call
	var toAddr sdk.AccAddress
	if args.To != nil {
		toAddr = sdk.AccAddress(args.To.Bytes())
	}

	// Create new call message
	msg := evmtypes.NewMsgEthermint(0, &toAddr, sdk.NewIntFromBigInt(value), gas,
		sdk.NewIntFromBigInt(gasPrice), data, sdk.AccAddress(addr.Bytes()))

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Generate tx to be used to simulate (signature isn't needed)
	var stdSig authtypes.StdSignature
	tx := authtypes.NewStdTx([]sdk.Msg{msg}, authtypes.StdFee{}, []authtypes.StdSignature{stdSig}, "")

	txEncoder := authclient.GetTxEncoder(clientCtx.Codec)
	txBytes, err := txEncoder(tx)
	if err != nil {
		return nil, err
	}

	// Transaction simulation through query
	res, _, err := clientCtx.QueryWithData("app/simulate", txBytes)
	if err != nil {
		return nil, err
	}

	var simResponse sdk.SimulationResponse
	if err := clientCtx.Codec.UnmarshalBinaryBare(res, &simResponse); err != nil {
		return nil, err
	}

	return &simResponse, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
// It adds 1,000 gas to the returned value instead of using the gas adjustment
// param from the SDK.
func (api *PublicEthereumAPI) EstimateGas(args rpctypes.CallArgs) (hexutil.Uint64, error) {
	api.logger.Debug("eth_estimateGas", "args", args)
	simResponse, err := api.doCall(args, 0, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return 0, err
	}

	// TODO: change 1000 buffer for more accurate buffer (eg: SDK's gasAdjusted)
	estimatedGas := simResponse.GasInfo.GasUsed
	gas := estimatedGas + 1000

	return hexutil.Uint64(gas), nil
}

// GetBlockByHash returns the block identified by hash.
func (api *PublicEthereumAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	api.logger.Debug("eth_getBlockByHash", "hash", hash, "full", fullTx)
	return api.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (api *PublicEthereumAPI) GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	api.logger.Debug("eth_getBlockByNumber", "number", blockNum, "full", fullTx)
	return api.backend.GetBlockByNumber(blockNum, fullTx)
}

// GetTransactionByHash returns the transaction identified by hash.
func (api *PublicEthereumAPI) GetTransactionByHash(hash common.Hash) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByHash", "hash", hash)
	tx, err := api.clientCtx.Client.Tx(hash.Bytes(), false)
	if err != nil {
		// Return nil for transaction when not found
		return nil, nil
	}

	// Can either cache or just leave this out if not necessary
	block, err := api.clientCtx.Client.Block(&tx.Height)
	if err != nil {
		return nil, err
	}

	blockHash := common.BytesToHash(block.Block.Header.Hash())

	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	height := uint64(tx.Height)
	return rpctypes.NewTransaction(ethTx, common.BytesToHash(tx.Tx.Hash()), blockHash, height, uint64(tx.Index))
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (api *PublicEthereumAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByHashAndIndex", "hash", hash, "index", idx)
	res, _, err := api.clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryHashToHeight, hash.Hex()))
	if err != nil {
		return nil, err
	}

	var out evmtypes.QueryResBlockNumber
	api.clientCtx.Codec.MustUnmarshalJSON(res, &out)

	resBlock, err := api.clientCtx.Client.Block(&out.Number)
	if err != nil {
		return nil, err
	}

	return api.getTransactionByBlockAndIndex(resBlock.Block, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (api *PublicEthereumAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)
	height := blockNum.Int64()
	resBlock, err := api.clientCtx.Client.Block(&height)
	if err != nil {
		return nil, err
	}

	return api.getTransactionByBlockAndIndex(resBlock.Block, idx)
}

func (api *PublicEthereumAPI) getTransactionByBlockAndIndex(block *tmtypes.Block, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	// return if index out of bounds
	if uint64(idx) >= uint64(len(block.Txs)) {
		return nil, nil
	}

	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, block.Txs[idx])
	if err != nil {
		// return nil error if the transaction is not a MsgEthereumTx
		return nil, nil
	}

	height := uint64(block.Height)
	txHash := common.BytesToHash(block.Txs[idx].Hash())
	blockHash := common.BytesToHash(block.Header.Hash())
	return rpctypes.NewTransaction(ethTx, txHash, blockHash, height, uint64(idx))
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (api *PublicEthereumAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	api.logger.Debug("eth_getTransactionReceipt", "hash", hash)
	tx, err := api.clientCtx.Client.Tx(hash.Bytes(), false)
	if err != nil {
		// Return nil for transaction when not found
		return nil, nil
	}

	// Query block for consensus hash
	block, err := api.clientCtx.Client.Block(&tx.Height)
	if err != nil {
		return nil, err
	}

	blockHash := common.BytesToHash(block.Block.Header.Hash())

	// Convert tx bytes to eth transaction
	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	from, err := ethTx.VerifySig(ethTx.ChainID())
	if err != nil {
		return nil, err
	}

	// Set status codes based on tx result
	var status hexutil.Uint
	if tx.TxResult.IsOK() {
		status = hexutil.Uint(1)
	} else {
		status = hexutil.Uint(0)
	}

	txData := tx.TxResult.GetData()

	data, err := evmtypes.DecodeResultData(txData)
	if err != nil {
		status = 0 // transaction failed
	}

	if len(data.Logs) == 0 {
		data.Logs = []*ethtypes.Log{}
	}

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": nil, // ignore until needed
		"logsBloom":         data.Bloom,
		"logs":              data.Logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": data.ContractAddress,
		"gasUsed":         hexutil.Uint64(tx.TxResult.GasUsed),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        blockHash,
		"blockNumber":      hexutil.Uint64(tx.Height),
		"transactionIndex": hexutil.Uint64(tx.Index),

		// sender and receiver (contract or EOA) addreses
		"from": from,
		"to":   ethTx.To(),
	}

	return receipt, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (api *PublicEthereumAPI) PendingTransactions() ([]*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getPendingTransactions")
	return api.backend.PendingTransactions()
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (api *PublicEthereumAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (api *PublicEthereumAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetProof returns an account object with proof and any storage proofs
func (api *PublicEthereumAPI) GetProof(address common.Address, storageKeys []string, block rpctypes.BlockNumber) (*rpctypes.AccountResult, error) {
	api.logger.Debug("eth_getProof", "address", address, "keys", storageKeys, "number", block)

	clientCtx := api.clientCtx.WithHeight(int64(block))
	path := fmt.Sprintf("custom/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryAccount, address.Hex())

	// query eth account at block height
	resBz, _, err := clientCtx.Query(path)
	if err != nil {
		return nil, err
	}

	var account evmtypes.QueryResAccount
	clientCtx.Codec.MustUnmarshalJSON(resBz, &account)

	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))
	opts := client.ABCIQueryOptions{Height: int64(block), Prove: true}
	for i, k := range storageKeys {
		// Get value for key
		vPath := fmt.Sprintf("custom/%s/%s/%s/%s", evmtypes.ModuleName, evmtypes.QueryStorage, address, k)
		vRes, err := api.clientCtx.Client.ABCIQueryWithOptions(vPath, nil, opts)
		if err != nil {
			return nil, err
		}

		var value evmtypes.QueryResStorage
		clientCtx.Codec.MustUnmarshalJSON(vRes.Response.GetValue(), &value)

		// check for proof
		proof := vRes.Response.GetProof()
		proofStr := new(merkle.Proof).String()
		if proof != nil {
			proofStr = proof.String()
		}

		storageProofs[i] = rpctypes.StorageResult{
			Key:   k,
			Value: (*hexutil.Big)(common.BytesToHash(value.Value).Big()),
			Proof: []string{proofStr},
		}
	}

	req := abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", auth.StoreKey),
		Data:   auth.AddressStoreKey(sdk.AccAddress(address.Bytes())),
		Height: int64(block),
		Prove:  true,
	}

	res, err := clientCtx.QueryABCI(req)
	if err != nil {
		return nil, err
	}

	// check for proof
	accountProof := res.GetProof()
	accProofStr := new(merkle.Proof).String()
	if accountProof != nil {
		accProofStr = accountProof.String()
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: []string{accProofStr},
		Balance:      (*hexutil.Big)(utils.MustUnmarshalBigInt(account.Balance)),
		CodeHash:     common.BytesToHash(account.CodeHash),
		Nonce:        hexutil.Uint64(account.Nonce),
		StorageHash:  common.Hash{}, // Ethermint doesn't have a storage hash
		StorageProof: storageProofs,
	}, nil
}

// generateFromArgs populates tx message with args (used in RPC API)
func (api *PublicEthereumAPI) generateFromArgs(args rpctypes.SendTxArgs) (*evmtypes.MsgEthereumTx, error) {
	var (
		nonce    uint64
		gasLimit uint64
		err      error
	)

	amount := (*big.Int)(args.Value)
	gasPrice := (*big.Int)(args.GasPrice)

	if args.GasPrice == nil {

		// Set default gas price
		// TODO: Change to min gas price from context once available through server/daemon
		gasPrice = big.NewInt(ethermint.DefaultGasPrice)
	}

	if args.Nonce == nil {
		// Get nonce (sequence) from account
		from := sdk.AccAddress(args.From.Bytes())
		accRet := authtypes.NewAccountRetriever(api.clientCtx)

		if api.clientCtx.Keybase == nil {
			return nil, fmt.Errorf("clientCtx.Keybase is nil")
		}

		_, nonce, err = accRet.GetAccountNumberSequence(from)
		if err != nil {
			return nil, err
		}
	} else {
		nonce = (uint64)(*args.Nonce)
	}

	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return nil, errors.New(`both "data" and "input" are set and not equal. Please use "input" to pass transaction call data`)
	}

	// Sets input to either Input or Data, if both are set and not equal error above returns
	var input []byte
	if args.Input != nil {
		input = *args.Input
	} else if args.Data != nil {
		input = *args.Data
	}

	if args.To == nil && len(input) == 0 {
		// Contract creation
		return nil, fmt.Errorf("contract creation without any data provided")
	}

	if args.Gas == nil {
		callArgs := rpctypes.CallArgs{
			From:     &args.From,
			To:       args.To,
			Gas:      args.Gas,
			GasPrice: args.GasPrice,
			Value:    args.Value,
			Data:     args.Data,
		}
		gl, err := api.EstimateGas(callArgs)
		if err != nil {
			return nil, err
		}
		gasLimit = uint64(gl)
	} else {
		gasLimit = (uint64)(*args.Gas)
	}
	msg := evmtypes.NewMsgEthereumTx(nonce, args.To, amount, gasLimit, gasPrice, input)

	return &msg, nil
}

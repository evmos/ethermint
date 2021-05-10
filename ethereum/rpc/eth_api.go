package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"sync"

	"github.com/cosmos/ethermint/ethereum/rpc/types"
	"github.com/gogo/protobuf/jsonpb"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	rpctypes "github.com/cosmos/ethermint/ethereum/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

// PublicEthAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthAPI struct {
	ctx          context.Context
	clientCtx    client.Context
	queryClient  *types.QueryClient
	chainIDEpoch *big.Int
	logger       log.Logger
	backend      Backend
	nonceLock    *types.AddrLocker
	keyringLock  sync.Mutex
}

// NewPublicEthAPI creates an instance of the public ETH Web3 API.
func NewPublicEthAPI(
	clientCtx client.Context,
	backend Backend,
	nonceLock *types.AddrLocker,
) *PublicEthAPI {
	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	api := &PublicEthAPI{
		ctx:          context.Background(),
		clientCtx:    clientCtx,
		queryClient:  types.NewQueryClient(clientCtx),
		chainIDEpoch: epoch,
		logger:       log.WithField("module", "json-rpc"),
		backend:      backend,
		nonceLock:    nonceLock,
	}

	return api
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicEthAPI) ProtocolVersion() hexutil.Uint {
	e.logger.Debugln("eth_protocolVersion")
	return hexutil.Uint(evmtypes.ProtocolVersion)
}

// ChainId returns the chain's identifier in hex format
func (e *PublicEthAPI) ChainId() (hexutil.Uint, error) { // nolint
	e.logger.Debugln("eth_chainId")
	return hexutil.Uint(uint(e.chainIDEpoch.Uint64())), nil
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (e *PublicEthAPI) Syncing() (interface{}, error) {
	e.logger.Debugln("eth_syncing")

	status, err := e.clientCtx.Client.Status(e.ctx)
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
func (e *PublicEthAPI) Coinbase() (common.Address, error) {
	e.logger.Debugln("eth_coinbase")

	node, err := e.clientCtx.GetNode()
	if err != nil {
		return common.Address{}, err
	}

	status, err := node.Status(e.ctx)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(status.ValidatorInfo.Address.Bytes()), nil
}

// Mining returns whether or not this node is currently mining. Always false.
func (e *PublicEthAPI) Mining() bool {
	e.logger.Debugln("eth_mining")
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (e *PublicEthAPI) Hashrate() hexutil.Uint64 {
	e.logger.Debugln("eth_hashrate")
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (e *PublicEthAPI) GasPrice() *hexutil.Big {
	e.logger.Debugln("eth_gasPrice")
	out := big.NewInt(0)
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (e *PublicEthAPI) Accounts() ([]common.Address, error) {
	e.logger.Debugln("eth_accounts")

	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := e.clientCtx.Keyring.List()
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
	//e.logger.Debugln("eth_blockNumber")
	return e.backend.BlockNumber()
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicEthAPI) GetBalance(address common.Address, blockNum types.BlockNumber) (*hexutil.Big, error) { // nolint: interfacer
	e.logger.Debugln("eth_getBalance", "address", address.String(), "block number", blockNum)

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Balance(types.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, err := ethermint.UnmarshalBigInt(res.Balance)
	if err != nil {
		return nil, err
	}

	return (*hexutil.Big)(val), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicEthAPI) GetStorageAt(address common.Address, key string, blockNum types.BlockNumber) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debugln("eth_getStorageAt", "address", address.Hex(), "key", key, "block number", blockNum)

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := e.queryClient.Storage(types.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicEthAPI) GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error) {
	e.logger.Debugln("eth_getTransactionCount", "address", address.Hex(), "block number", blockNum)

	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := e.clientCtx.AccountRetriever

	err := accRet.EnsureExists(e.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		n := hexutil.Uint64(0)
		return &n, nil
	}

	_, nonce, err := accRet.GetAccountNumberSequence(e.clientCtx, from)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicEthAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	e.logger.Debugln("eth_getBlockTransactionCountByHash", "hash", hash.Hex())

	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicEthAPI) GetBlockTransactionCountByNumber(blockNum types.BlockNumber) *hexutil.Uint {
	e.logger.Debugln("eth_getBlockTransactionCountByNumber", "block number", blockNum)
	resBlock, err := e.clientCtx.Client.Block(e.ctx, blockNum.TmHeight())
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetUncleCountByBlockHash returns the number of uncles in the block identified by hash. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block identified by number. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockNumber(blockNum types.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicEthAPI) GetCode(address common.Address, blockNumber types.BlockNumber) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debugln("eth_getCode", "address", address.Hex(), "block number", blockNumber)

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Code(types.ContextWithHeight(blockNumber.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (e *PublicEthAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	e.logger.Debugln("eth_getTransactionLogs", "hash", txHash)
	return e.backend.GetTransactionLogs(txHash)
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicEthAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	e.logger.Debugln("eth_sign", "address", address.Hex(), "data", common.Bytes2Hex(data))
	return nil, errors.New("eth_sign not supported")
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicEthAPI) SendTransaction(args types.SendTxArgs) (common.Hash, error) {
	e.logger.Debugln("eth_sendTransaction", "args", args)
	return common.Hash{}, errors.New("eth_sendTransaction not supported")
}

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicEthAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	e.logger.Debugln("eth_sendRawTransaction", "data_len", len(data))
	ethereumTx := new(evmtypes.MsgEthereumTx)

	// RLP decode raw transaction bytes
	if err := rlp.DecodeBytes(data, ethereumTx); err != nil {
		e.logger.WithError(err).Errorln("transaction RLP decode failed")

		// Return nil is for when gasLimit overflows uint64
		return common.Hash{}, err
	}

	builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		e.logger.Panicln("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		e.logger.Panicln("codectypes.NewAnyWithValue failed to pack an obvious value")
	}

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(ethereumTx.GetMsgs()...)
	if err != nil {
		e.logger.Panicln("builder.SetMsgs failed")
	}

	fees := sdk.NewCoins(ethermint.NewPhotonCoin(sdk.NewIntFromBigInt(ethereumTx.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(ethereumTx.GetGas())

	// Encode transaction by default Tx encoder
	txBytes, err := e.clientCtx.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		e.logger.WithError(err).Errorln("failed to encode Eth tx using default encoder")
		return common.Hash{}, err
	}

	txHash := common.BytesToHash(tmhash.Sum(txBytes))
	asyncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastAsync)

	if _, err := asyncCtx.BroadcastTx(txBytes); err != nil {
		e.logger.WithError(err).Errorln("failed to broadcast Eth tx")
		return txHash, err
	}

	return txHash, nil
}

// Call performs a raw contract call.
func (e *PublicEthAPI) Call(args types.CallArgs, blockNr types.BlockNumber, _ *types.StateOverride) (hexutil.Bytes, error) {
	//e.logger.Debugln("eth_call", "args", args, "block number", blockNr)
	simRes, err := e.doCall(args, blockNr, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return []byte{}, err
	} else if len(simRes.Result.Log) > 0 {
		var logs []types.SDKTxLogs
		if err := json.Unmarshal([]byte(simRes.Result.Log), &logs); err != nil {
			e.logger.WithError(err).Errorln("failed to unmarshal simRes.Result.Log")
		}

		if len(logs) > 0 && logs[0].Log == types.LogRevertedFlag {
			data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
			if err != nil {
				e.logger.WithError(err).Warningln("call result decoding failed")
				return []byte{}, err
			}
			return []byte{}, types.ErrRevertedWith(data.Ret)
		}
	}

	data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
	if err != nil {
		e.logger.WithError(err).Warningln("call result decoding failed")
		return []byte{}, err
	}

	return (hexutil.Bytes)(data.Ret), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (e *PublicEthAPI) doCall(
	args types.CallArgs, blockNr types.BlockNumber, globalGasCap *big.Int,
) (*sdk.SimulationResponse, error) {
	// Set default gas & gas price if none were set
	// Change this to uint64(math.MaxUint64 / 2) if gas cap can be configured
	gas := uint64(ethermint.DefaultRPCGasLimit)
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != nil && globalGasCap.Uint64() < gas {
		e.logger.Debugln("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
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

	var accessList *ethtypes.AccessList
	if args.AccessList != nil {
		accessList = args.AccessList
	}

	// Set destination address for call
	var fromAddr sdk.AccAddress
	if args.From != nil {
		fromAddr = sdk.AccAddress(args.From.Bytes())
	} else {
		fromAddr = sdk.AccAddress(common.Address{}.Bytes())
	}

	_, seq, err := e.clientCtx.AccountRetriever.GetAccountNumberSequence(e.clientCtx, fromAddr)
	if err != nil {
		return nil, err
	}

	// Create new call message
	msg := evmtypes.NewMsgEthereumTx(e.chainIDEpoch, seq, args.To, value, gas, gasPrice, data, accessList)
	msg.From = fromAddr.String()

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Create a TxBuilder
	txBuilder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		log.Panicln("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		log.Panicln("codectypes.NewAnyWithValue failed to pack an obvious value")
	}
	txBuilder.SetExtensionOptions(option)

	if err := txBuilder.SetMsgs(msg); err != nil {
		log.Panicln("builder.SetMsgs failed")
	}

	fees := sdk.NewCoins(ethermint.NewPhotonCoin(sdk.NewIntFromBigInt(msg.Fee())))
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(gas)

	//doc about generate and transform tx into json, protobuf bytes
	//https://github.com/cosmos/cosmos-sdk/blob/master/docs/run-node/txs.md
	txBytes, err := e.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	// simulate by calling ABCI Query
	query := abci.RequestQuery{
		Path:   "/app/simulate",
		Data:   txBytes,
		Height: blockNr.Int64(),
	}

	queryResult, err := e.clientCtx.QueryABCI(query)
	if err != nil {
		return nil, err
	}

	var simResponse sdk.SimulationResponse
	err = jsonpb.Unmarshal(strings.NewReader(string(queryResult.Value)), &simResponse)
	if err != nil {
		return nil, err
	}

	return &simResponse, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
// It adds 1,000 gas to the returned value instead of using the gas adjustment
// param from the SDK.
func (e *PublicEthAPI) EstimateGas(args types.CallArgs) (hexutil.Uint64, error) {
	e.logger.Debugln("eth_estimateGas")

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	simRes, err := e.doCall(args, 0, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return 0, err
	}

	if len(simRes.Result.Log) > 0 {
		var logs []types.SDKTxLogs
		if err := json.Unmarshal([]byte(simRes.Result.Log), &logs); err != nil {
			e.logger.WithError(err).Errorln("failed to unmarshal simRes.Result.Log")
			return 0, err
		}

		if len(logs) > 0 && logs[0].Log == types.LogRevertedFlag {
			data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
			if err != nil {
				e.logger.WithError(err).Warningln("call result decoding failed")
				return 0, err
			}

			return 0, types.ErrRevertedWith(data.Ret)
		}
	}

	// TODO: change 1000 buffer for more accurate buffer (eg: SDK's gasAdjusted)
	estimatedGas := simRes.GasInfo.GasUsed
	gas := estimatedGas + 200000

	return hexutil.Uint64(gas), nil
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicEthAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debugln("eth_getBlockByHash", "hash", hash.Hex(), "full", fullTx)
	return e.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicEthAPI) GetBlockByNumber(ethBlockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	// e.logger.Debugln("eth_getBlockByNumber", "number", ethBlockNum, "full", fullTx)
	return e.backend.GetBlockByNumber(ethBlockNum, fullTx)
}

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicEthAPI) GetTransactionByHash(hash common.Hash) (*types.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByHash", "hash", hash.Hex())

	resp, err := e.queryClient.TxReceipt(e.ctx, &evmtypes.QueryTxReceiptRequest{
		Hash: hash.Hex(),
	})
	if err != nil {
		e.logger.Debugf("failed to get tx info for %s: %s", hash.Hex(), err.Error())
		return nil, nil
	}

	return types.NewTransactionFromData(
		resp.Receipt.Data,
		common.BytesToAddress(resp.Receipt.From),
		common.BytesToHash(resp.Receipt.Hash),
		common.BytesToHash(resp.Receipt.BlockHash),
		resp.Receipt.BlockHeight,
		resp.Receipt.Index,
	)
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicEthAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*types.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByHashAndIndex", "hash", hash.Hex(), "index", idx)

	resp, err := e.queryClient.TxReceiptsByBlockHash(e.ctx, &evmtypes.QueryTxReceiptsByBlockHashRequest{
		Hash: hash.Hex(),
	})
	if err != nil {
		err = errors.Wrap(err, "failed to query tx receipts by block hash")
		return nil, err
	}

	return e.getReceiptByIndex(resp.Receipts, hash, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicEthAPI) GetTransactionByBlockNumberAndIndex(blockNum types.BlockNumber, idx hexutil.Uint) (*types.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)

	resp, err := e.queryClient.TxReceiptsByBlockHeight(e.ctx, &evmtypes.QueryTxReceiptsByBlockHeightRequest{
		Height: blockNum.Int64(),
	})
	if err != nil {
		err = errors.Wrap(err, "failed to query tx receipts by block height")
		return nil, err
	}

	return e.getReceiptByIndex(resp.Receipts, common.Hash{}, idx)
}

func (e *PublicEthAPI) getReceiptByIndex(receipts []*evmtypes.TxReceipt, blockHash common.Hash, idx hexutil.Uint) (*types.RPCTransaction, error) {
	// return if index out of bounds
	if uint64(idx) >= uint64(len(receipts)) {
		return nil, nil
	}

	receipt := receipts[idx]

	if (blockHash != common.Hash{}) {
		if !bytes.Equal(receipt.BlockHash, blockHash.Bytes()) {
			err := errors.Errorf("receipt found but block hashes don't match %s != %s",
				common.Bytes2Hex(receipt.BlockHash),
				blockHash.Hex(),
			)

			return nil, err
		}
	}

	return types.NewTransactionFromData(
		receipt.Data,
		common.BytesToAddress(receipt.From),
		common.BytesToHash(receipt.Hash),
		blockHash,
		receipt.BlockHeight,
		uint64(idx),
	)
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicEthAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	e.logger.Debugln("eth_getTransactionReceipt", "hash", hash.Hex())

	ctx := types.ContextWithHeight(int64(0))
	tx, err := e.queryClient.TxReceipt(ctx, &evmtypes.QueryTxReceiptRequest{
		Hash: hash.Hex(),
	})
	if err != nil {
		e.logger.Debugf("failed to get tx receipt for %s: %s", hash.Hex(), err.Error())
		return nil, nil
	}

	// Query block for consensus hash
	height := int64(tx.Receipt.BlockHeight)
	block, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		e.logger.Warningln("didnt find block for tx height", height)
		return nil, err
	}

	cumulativeGasUsed := uint64(tx.Receipt.Result.GasUsed)
	if tx.Receipt.Index != 0 {
		cumulativeGasUsed += rpctypes.GetBlockCumulativeGas(e.clientCtx, block.Block, int(tx.Receipt.Index))
	}

	var status hexutil.Uint
	if tx.Receipt.Result.Reverted {
		status = hexutil.Uint(0)
	} else {
		status = hexutil.Uint(1)
	}

	toHex := common.Address{}
	if len(tx.Receipt.Data.To) > 0 {
		toHex = common.BytesToAddress(tx.Receipt.Data.To)
	}

	contractAddress := common.HexToAddress(tx.Receipt.Result.ContractAddress)
	logsBloom := hexutil.Encode(ethtypes.BytesToBloom(tx.Receipt.Result.Bloom).Bytes())
	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         logsBloom,
		"logs":              tx.Receipt.Result.TxLogs.EthLogs(),

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash.Hex(),
		"contractAddress": contractAddress.Hex(),
		"gasUsed":         hexutil.Uint64(tx.Receipt.Result.GasUsed),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        common.BytesToHash(block.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(tx.Receipt.BlockHeight),
		"transactionIndex": hexutil.Uint64(tx.Receipt.Index),

		// sender and receiver (contract or EOA) addreses
		"from": common.BytesToAddress(tx.Receipt.From),
		"to":   toHex,
	}

	return receipt, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *PublicEthAPI) PendingTransactions() ([]*types.RPCTransaction, error) {
	e.logger.Debugln("eth_getPendingTransactions")
	return e.backend.PendingTransactions()
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetProof returns an account object with proof and any storage proofs
func (e *PublicEthAPI) GetProof(address common.Address, storageKeys []string, blockNumber types.BlockNumber) (*types.AccountResult, error) {
	height := blockNumber.Int64()
	e.logger.Debugln("eth_getProof", "address", address.Hex(), "keys", storageKeys, "number", height)

	ctx := types.ContextWithHeight(height)
	clientCtx := e.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]types.StorageResult, len(storageKeys))
	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := e.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
		if err != nil {
			return nil, err
		}

		// check for proof
		var proofStr string
		if proof != nil {
			proofStr = proof.String()
		}

		storageProofs[i] = types.StorageResult{
			Key:   key,
			Value: (*hexutil.Big)(new(big.Int).SetBytes(valueBz)),
			Proof: []string{proofStr},
		}
	}

	// query EVM account
	req := &evmtypes.QueryAccountRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := authtypes.AddressStoreKey(sdk.AccAddress(address.Bytes()))
	_, proof, err := e.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	// check for proof
	var accProofStr string
	if proof != nil {
		accProofStr = proof.String()
	}

	balance, err := ethermint.UnmarshalBigInt(res.Balance)
	if err != nil {
		return nil, err
	}

	return &types.AccountResult{
		Address:      address,
		AccountProof: []string{accProofStr},
		Balance:      (*hexutil.Big)(balance),
		CodeHash:     common.BytesToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Ethermint doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

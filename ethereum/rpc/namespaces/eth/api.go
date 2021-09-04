package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/tharsis/ethermint/crypto/hd"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// PublicAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicAPI struct {
	ctx          context.Context
	clientCtx    client.Context
	queryClient  *rpctypes.QueryClient
	chainIDEpoch *big.Int
	logger       log.Logger
	backend      backend.Backend
	nonceLock    *rpctypes.AddrLocker
}

// NewPublicAPI creates an instance of the public ETH Web3 API.
func NewPublicAPI(
	logger log.Logger,
	clientCtx client.Context,
	backend backend.Backend,
	nonceLock *rpctypes.AddrLocker,
) *PublicAPI {
	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	algos, _ := clientCtx.Keyring.SupportedAlgorithms()

	if !algos.Contains(hd.EthSecp256k1) {
		kr, err := keyring.New(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			clientCtx.KeyringDir,
			clientCtx.Input,
			hd.EthSecp256k1Option(),
		)

		if err != nil {
			panic(err)
		}

		clientCtx = clientCtx.WithKeyring(kr)
	}

	api := &PublicAPI{
		ctx:          context.Background(),
		clientCtx:    clientCtx,
		queryClient:  rpctypes.NewQueryClient(clientCtx),
		chainIDEpoch: epoch,
		logger:       logger.With("client", "json-rpc"),
		backend:      backend,
		nonceLock:    nonceLock,
	}

	return api
}

// ClientCtx returns client context
func (e *PublicAPI) ClientCtx() client.Context {
	return e.clientCtx
}

func (e *PublicAPI) QueryClient() *rpctypes.QueryClient {
	return e.queryClient
}

func (e *PublicAPI) Ctx() context.Context {
	return e.ctx
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicAPI) ProtocolVersion() hexutil.Uint {
	e.logger.Debug("eth_protocolVersion")
	return hexutil.Uint(ethermint.ProtocolVersion)
}

// ChainId returns the chain's identifier in hex format
func (e *PublicAPI) ChainId() (hexutil.Uint, error) { // nolint
	e.logger.Debug("eth_chainId")
	return hexutil.Uint(uint(e.chainIDEpoch.Uint64())), nil
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (e *PublicAPI) Syncing() (interface{}, error) {
	e.logger.Debug("eth_syncing")

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
func (e *PublicAPI) Coinbase() (string, error) {
	e.logger.Debug("eth_coinbase")

	coinbase, err := e.backend.GetCoinbase()
	if err != nil {
		return "", err
	}
	ethAddr := common.BytesToAddress(coinbase.Bytes())
	return ethAddr.Hex(), nil
}

// Mining returns whether or not this node is currently mining. Always false.
func (e *PublicAPI) Mining() bool {
	e.logger.Debug("eth_mining")
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (e *PublicAPI) Hashrate() hexutil.Uint64 {
	e.logger.Debug("eth_hashrate")
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (e *PublicAPI) GasPrice() *hexutil.Big {
	e.logger.Debug("eth_gasPrice")
	out := new(big.Int).SetUint64(e.backend.RPCGasCap())
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (e *PublicAPI) Accounts() ([]common.Address, error) {
	e.logger.Debug("eth_accounts")

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
func (e *PublicAPI) BlockNumber() (hexutil.Uint64, error) {
	e.logger.Debug("eth_blockNumber")
	return e.backend.BlockNumber()
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicAPI) GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) { // nolint: interfacer
	e.logger.Debug("eth_getBalance", "address", address.String(), "block number or hash", blockNrOrHash)

	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Balance(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, ok := sdk.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return (*hexutil.Big)(val.BigInt()), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicAPI) GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debug("eth_getStorageAt", "address", address.Hex(), "key", key, "block number or hash", blockNrOrHash)

	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := e.queryClient.Storage(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicAPI) GetTransactionCount(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Uint64, error) {
	e.logger.Debug("eth_getTransactionCount", "address", address.Hex(), "block number or hash", blockNrOrHash)
	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	return e.backend.GetTransactionCount(address, blockNum)
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	e.logger.Debug("eth_getBlockTransactionCountByHash", "hash", hash.Hex())

	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		e.logger.Debug("block not found", "hash", hash.Hex(), "error", err.Error())
		return nil
	}

	if resBlock.Block == nil {
		e.logger.Debug("block not found", "hash", hash.Hex())
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	e.logger.Debug("eth_getBlockTransactionCountByNumber", "height", blockNum.Int64())
	resBlock, err := e.clientCtx.Client.Block(e.ctx, blockNum.TmHeight())
	if err != nil {
		e.logger.Debug("block not found", "height", blockNum.Int64(), "error", err.Error())
		return nil
	}

	if resBlock.Block == nil {
		e.logger.Debug("block not found", "height", blockNum.Int64())
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetUncleCountByBlockHash returns the number of uncles in the block identified by hash. Always zero.
func (e *PublicAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block identified by number. Always zero.
func (e *PublicAPI) GetUncleCountByBlockNumber(blockNum rpctypes.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicAPI) GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debug("eth_getCode", "address", address.Hex(), "block number or hash", blockNrOrHash)

	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Code(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (e *PublicAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	e.logger.Debug("eth_getTransactionLogs", "hash", txHash)
	return e.backend.GetTransactionLogs(txHash)
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	e.logger.Debug("eth_sign", "address", address.Hex(), "data", common.Bytes2Hex(data))

	from := sdk.AccAddress(address.Bytes())

	_, err := e.clientCtx.Keyring.KeyByAddress(from)
	if err != nil {
		e.logger.Error("failed to find key in keyring", "address", address.String())
		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	// Sign the requested hash with the wallet
	signature, _, err := e.clientCtx.Keyring.SignByAddress(from, data)
	if err != nil {
		e.logger.Error("keyring.SignByAddress failed", "address", address.Hex())
		return nil, err
	}

	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicAPI) SendTransaction(args rpctypes.SendTxArgs) (common.Hash, error) {
	e.logger.Debug("eth_sendTransaction", "args", args.String())
	return e.backend.SendTransaction(args)
}

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	e.logger.Debug("eth_sendRawTransaction", "length", len(data))

	// RLP decode raw transaction bytes
	tx, err := e.clientCtx.TxConfig.TxDecoder()(data)
	if err != nil {
		e.logger.Error("transaction decoding failed", "error", err.Error())

		return common.Hash{}, err
	}

	ethereumTx, isEthTx := tx.(*evmtypes.MsgEthereumTx)
	if !isEthTx {
		e.logger.Debug("invalid transaction type", "type", fmt.Sprintf("%T", tx))
		return common.Hash{}, fmt.Errorf("invalid transaction type %T", tx)
	}

	if err := ethereumTx.ValidateBasic(); err != nil {
		e.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		e.logger.Error("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		e.logger.Error("codectypes.NewAnyWithValue failed to pack an obvious value", "error", err.Error())
	}

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(tx.GetMsgs()...)
	if err != nil {
		e.logger.Error("builder.SetMsgs failed", "error", err.Error())
	}

	// Query params to use the EVM denomination
	res, err := e.queryClient.QueryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		e.logger.Error("failed to query evm params", "error", err.Error())
		return common.Hash{}, err
	}

	txData, err := evmtypes.UnpackTxData(ethereumTx.Data)
	if err != nil {
		e.logger.Error("failed to unpack tx data", "error", err.Error())
		return common.Hash{}, err
	}

	fees := sdk.Coins{sdk.NewCoin(res.Params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(ethereumTx.GetGas())

	// Encode transaction by default Tx encoder
	txBytes, err := e.clientCtx.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		e.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	txHash := ethereumTx.AsTransaction().Hash()

	syncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if err != nil || rsp.Code != 0 {
		if err == nil {
			err = errors.New(rsp.RawLog)
		}
		e.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	return txHash, nil
}

// Call performs a raw contract call.
func (e *PublicAPI) Call(args evmtypes.CallArgs, blockNrOrHash rpctypes.BlockNumberOrHash, _ *rpctypes.StateOverride) (hexutil.Bytes, error) {
	e.logger.Debug("eth_call", "args", args.String(), "block number or hash", blockNrOrHash)

	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	data, err := e.doCall(args, blockNum)
	if err != nil {
		return []byte{}, err
	}

	return (hexutil.Bytes)(data.Ret), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (e *PublicAPI) doCall(
	args evmtypes.CallArgs, blockNr rpctypes.BlockNumber,
) (*evmtypes.MsgEthereumTxResponse, error) {
	bz, err := json.Marshal(&args)
	if err != nil {
		return nil, err
	}
	req := evmtypes.EthCallRequest{Args: bz, GasCap: e.backend.RPCGasCap()}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	res, err := e.queryClient.EthCall(rpctypes.ContextWithHeight(blockNr.Int64()), &req)
	if err != nil {
		return nil, err
	}

	if res.Failed() {
		if res.VmError != vm.ErrExecutionReverted.Error() {
			return nil, status.Error(codes.Internal, res.VmError)
		}
		return nil, evmtypes.NewExecErrorWithReason(res.Ret)
	}

	return res, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (e *PublicAPI) EstimateGas(args evmtypes.CallArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error) {
	e.logger.Debug("eth_estimateGas")
	return e.backend.EstimateGas(args, blockNrOptional)
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debug("eth_getBlockByHash", "hash", hash.Hex(), "full", fullTx)
	return e.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicAPI) GetBlockByNumber(ethBlockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debug("eth_getBlockByNumber", "number", ethBlockNum, "full", fullTx)
	return e.backend.GetBlockByNumber(ethBlockNum, fullTx)
}

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicAPI) GetTransactionByHash(hash common.Hash) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByHash", "hash", hash.Hex())
	return e.backend.GetTransactionByHash(hash)
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByBlockHashAndIndex", "hash", hash.Hex(), "index", idx)

	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		e.logger.Debug("block not found", "hash", hash.Hex(), "error", err.Error())
		return nil, nil
	}

	if resBlock.Block == nil {
		e.logger.Debug("block not found", "hash", hash.Hex())
		return nil, nil
	}

	i := int(idx)
	if i >= len(resBlock.Block.Txs) {
		e.logger.Debug("block txs index out of bound", "index", i)
		return nil, nil
	}

	txBz := resBlock.Block.Txs[i]
	tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
	if err != nil {
		e.logger.Debug("decoding failed", "error", err.Error())
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, err := evmtypes.UnwrapEthereumMsg(&tx)
	if err != nil {
		e.logger.Debug("invalid tx", "error", err.Error())
		return nil, err
	}

	return rpctypes.NewTransactionFromMsg(
		msg,
		hash,
		uint64(resBlock.Block.Height),
		uint64(idx),
		e.chainIDEpoch,
	)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)

	resBlock, err := e.clientCtx.Client.Block(e.ctx, blockNum.TmHeight())
	if err != nil {
		e.logger.Debug("block not found", "height", blockNum.Int64(), "error", err.Error())
		return nil, nil
	}

	if resBlock.Block == nil {
		e.logger.Debug("block not found", "height", blockNum.Int64())
		return nil, nil
	}

	i := int(idx)
	if i >= len(resBlock.Block.Txs) {
		e.logger.Debug("block txs index out of bound", "index", i)
		return nil, nil
	}

	txBz := resBlock.Block.Txs[i]
	tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
	if err != nil {
		e.logger.Debug("decoding failed", "error", err.Error())
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, err := evmtypes.UnwrapEthereumMsg(&tx)
	if err != nil {
		e.logger.Debug("invalid tx", "error", err.Error())
		return nil, err
	}

	return rpctypes.NewTransactionFromMsg(
		msg,
		common.BytesToHash(resBlock.Block.Hash()),
		uint64(resBlock.Block.Height),
		uint64(idx),
		e.chainIDEpoch,
	)
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	e.logger.Debug("eth_getTransactionReceipt", "hash", hash.Hex())

	res, err := e.backend.GetTxByEthHash(hash)
	if err != nil {
		e.logger.Debug("tx not found", "hash", hash.Hex(), "error", err.Error())
		return nil, nil
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &res.Height)
	if err != nil {
		e.logger.Debug("block not found", "height", res.Height, "error", err.Error())
		return nil, nil
	}

	tx, err := e.clientCtx.TxConfig.TxDecoder()(res.Tx)
	if err != nil {
		e.logger.Debug("decoding failed", "error", err.Error())
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, err := evmtypes.UnwrapEthereumMsg(&tx)
	if err != nil {
		e.logger.Debug("invalid tx", "error", err.Error())
		return nil, err
	}

	txData, err := evmtypes.UnpackTxData(msg.Data)
	if err != nil {
		e.logger.Error("failed to unpack tx data", "error", err.Error())
		return nil, err
	}

	cumulativeGasUsed := uint64(0)
	blockRes, err := e.clientCtx.Client.BlockResults(e.ctx, &res.Height)
	if err != nil {
		e.logger.Debug("failed to retrieve block results", "height", res.Height, "error", err.Error())
		return nil, nil
	}

	for i := 0; i <= int(res.Index) && i < len(blockRes.TxsResults); i++ {
		cumulativeGasUsed += uint64(blockRes.TxsResults[i].GasUsed)
	}

	// Get the transaction result from the log
	var status hexutil.Uint
	if strings.Contains(res.TxResult.GetLog(), evmtypes.AttributeKeyEthereumTxFailed) {
		status = hexutil.Uint(ethtypes.ReceiptStatusFailed)
	} else {
		status = hexutil.Uint(ethtypes.ReceiptStatusSuccessful)
	}

	from, err := msg.GetSender(e.chainIDEpoch)
	if err != nil {
		return nil, err
	}

	resLogs, err := e.queryClient.TxLogs(e.ctx, &evmtypes.QueryTxLogsRequest{Hash: hash.Hex()})
	if err != nil {
		e.logger.Debug("logs not found", "hash", hash.Hex(), "error", err.Error())
		resLogs = &evmtypes.QueryTxLogsResponse{Logs: []*evmtypes.Log{}}
	}

	logs := evmtypes.LogsToEthereum(resLogs.Logs)

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         ethtypes.BytesToBloom(ethtypes.LogsBloom(logs)),
		"logs":              logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(res.TxResult.GasUsed),
		"type":            hexutil.Uint(txData.TxType()),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        common.BytesToHash(resBlock.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(res.Height),
		"transactionIndex": hexutil.Uint64(res.Index),

		// sender and receiver (contract or EOA) addreses
		"from": from,
		"to":   txData.GetTo(),
	}

	if logs == nil {
		receipt["logs"] = [][]*ethtypes.Log{}
	}

	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if txData.GetTo() == nil {
		receipt["contractAddress"] = crypto.CreateAddress(from, txData.GetNonce())
	}

	return receipt, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *PublicAPI) PendingTransactions() ([]*rpctypes.RPCTransaction, error) {
	e.logger.Debug("eth_getPendingTransactions")

	txs, err := e.backend.PendingTransactions()
	if err != nil {
		return nil, err
	}

	result := make([]*rpctypes.RPCTransaction, 0, len(txs))
	for _, tx := range txs {
		msg, err := evmtypes.UnwrapEthereumMsg(tx)
		if err != nil {
			// not valid ethereum tx
			continue
		}

		rpctx, err := rpctypes.NewTransactionFromMsg(
			msg,
			common.Hash{},
			uint64(0),
			uint64(0),
			e.chainIDEpoch,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, rpctx)
	}

	return result, nil
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetProof returns an account object with proof and any storage proofs
func (e *PublicAPI) GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error) {
	e.logger.Debug("eth_getProof", "address", address.Hex(), "keys", storageKeys, "block number or hash", blockNrOrHash)

	blockNum, err := e.getBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}
	height := blockNum.Int64()

	ctx := rpctypes.ContextWithHeight(height)
	clientCtx := e.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))

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

		storageProofs[i] = rpctypes.StorageResult{
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

	balance, ok := sdk.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: []string{accProofStr},
		Balance:      (*hexutil.Big)(balance.BigInt()),
		CodeHash:     common.HexToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Ethermint doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// getBlockNumber returns the BlockNumber from BlockNumberOrHash
func (e *PublicAPI) getBlockNumber(blockNrOrHash rpctypes.BlockNumberOrHash) (rpctypes.BlockNumber, error) {
	switch {
	case blockNrOrHash.BlockHash == nil && blockNrOrHash.BlockNumber == nil:
		return rpctypes.EthEarliestBlockNumber, fmt.Errorf("types BlockHash and BlockNumber cannot be both nil")
	case blockNrOrHash.BlockHash != nil:
		blockHeader, err := e.backend.HeaderByHash(*blockNrOrHash.BlockHash)
		if err != nil {
			return rpctypes.EthEarliestBlockNumber, err
		}
		return rpctypes.NewBlockNumber(blockHeader.Number), nil
	case blockNrOrHash.BlockNumber != nil:
		return *blockNrOrHash.BlockNumber, nil
	default:
		return rpctypes.EthEarliestBlockNumber, nil
	}
}

package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/params"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/eth/filters"
	"github.com/tharsis/ethermint/rpc/ethereum/types"
	"github.com/tharsis/ethermint/server/config"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

// Backend implements the functionality shared within namespaces.
// Implemented by EVMBackend.
type Backend interface {
	// General Ethereum API
	RPCGasCap() uint64 // global gas cap for eth_call over rpc: DoS protection
	RPCMinGasPrice() int64
	SuggestGasTipCap() (*big.Int, error)

	// Blockchain API
	BlockNumber() (hexutil.Uint64, error)
	GetTendermintBlockByNumber(blockNum types.BlockNumber) (*tmrpctypes.ResultBlock, error)
	GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	CurrentHeader() *ethtypes.Header
	HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	PendingTransactions() ([]*sdk.Tx, error)
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error)
	SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error)
	GetCoinbase() (sdk.AccAddress, error)
	GetTransactionByHash(txHash common.Hash) (*types.RPCTransaction, error)
	GetTxByEthHash(txHash common.Hash) (*tmrpctypes.ResultTx, error)
	EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error)
	BaseFee() (*big.Int, error)

	// Filter API
	BloomStatus() (uint64, uint64)
	GetLogs(hash common.Hash) ([][]*ethtypes.Log, error)
	GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error)
	GetFilteredBlocks(from int64, to int64, filter [][]filters.BloomIV, filterAddresses bool) ([]int64, error)

	ChainConfig() *params.ChainConfig
	SetTxDefaults(args evmtypes.TransactionArgs) (evmtypes.TransactionArgs, error)
}

var _ Backend = (*EVMBackend)(nil)

var bAttributeKeyEthereumBloom = []byte(evmtypes.AttributeKeyEthereumBloom)

// EVMBackend implements the Backend interface
type EVMBackend struct {
	ctx         context.Context
	clientCtx   client.Context
	queryClient *types.QueryClient // gRPC query client
	logger      log.Logger
	chainID     *big.Int
	cfg         config.Config
}

// NewEVMBackend creates a new EVMBackend instance
func NewEVMBackend(ctx *server.Context, logger log.Logger, clientCtx client.Context) *EVMBackend {
	chainID, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	appConf := config.GetConfig(ctx.Viper)

	return &EVMBackend{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: types.NewQueryClient(clientCtx),
		logger:      logger.With("module", "evm-backend"),
		chainID:     chainID,
		cfg:         appConf,
	}
}

// BlockNumber returns the current block number in abci app state.
// Because abci app state could lag behind from tendermint latest block, it's more stable
// for the client to use the latest block number in abci app state than tendermint rpc.
func (e *EVMBackend) BlockNumber() (hexutil.Uint64, error) {
	// do any grpc query, ignore the response and use the returned block height
	var header metadata.MD
	_, err := e.queryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	if err != nil {
		return hexutil.Uint64(0), err
	}

	blockHeightHeader := header.Get(grpctypes.GRPCBlockHeightHeader)
	if headerLen := len(blockHeightHeader); headerLen != 1 {
		return 0, fmt.Errorf("unexpected '%s' gRPC header length; got %d, expected: %d", grpctypes.GRPCBlockHeightHeader, headerLen, 1)
	}

	height, err := strconv.ParseUint(blockHeightHeader[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse block height: %w", err)
	}

	return hexutil.Uint64(height), nil
}

// GetBlockByNumber returns the block identified by number.
func (e *EVMBackend) GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := e.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	// return if requested block height is greater than the current one
	if resBlock == nil || resBlock.Block == nil {
		return nil, nil
	}

	res, err := e.EthBlockFromTendermint(resBlock.Block, fullTx)
	if err != nil {
		e.logger.Debug("EthBlockFromTendermint failed", "height", blockNum, "error", err.Error())
		return nil, err
	}

	return res, nil
}

// GetBlockByHash returns the block identified by hash.
func (e *EVMBackend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		e.logger.Debug("BlockByHash block not found", "hash", hash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock.Block == nil {
		e.logger.Debug("BlockByHash block not found", "hash", hash.Hex())
		return nil, nil
	}

	return e.EthBlockFromTendermint(resBlock.Block, fullTx)
}

// GetTendermintBlockByNumber returns a Tendermint format block by block number
func (e *EVMBackend) GetTendermintBlockByNumber(blockNum types.BlockNumber) (*tmrpctypes.ResultBlock, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			return nil, errors.Errorf("cannot fetch a negative block height: %d", height)
		}
		if height > int64(currentBlockNumber) {
			return nil, nil
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		if resBlock, err = e.clientCtx.Client.Block(e.ctx, nil); err != nil {
			e.logger.Debug("tendermint client failed to get latest block", "height", height, "error", err.Error())
			return nil, nil
		}
	}

	if resBlock.Block == nil {
		e.logger.Debug("GetBlockByNumber block not found", "height", height)
		return nil, nil
	}

	return resBlock, nil
}

// BlockBloom query block bloom filter from block results
func (e *EVMBackend) BlockBloom(height *int64) (ethtypes.Bloom, error) {
	result, err := e.clientCtx.Client.BlockResults(e.ctx, height)
	if err != nil {
		return ethtypes.Bloom{}, err
	}
	for _, event := range result.EndBlockEvents {
		if event.Type != evmtypes.EventTypeBlockBloom {
			continue
		}

		for _, attr := range event.Attributes {
			if bytes.Equal(attr.Key, bAttributeKeyEthereumBloom) {
				return ethtypes.BytesToBloom(attr.Value), nil
			}
		}
	}
	return ethtypes.Bloom{}, errors.New("block bloom event is not found")
}

// EthBlockFromTendermint returns a JSON-RPC compatible Ethereum block from a given Tendermint block and its block result.
func (e *EVMBackend) EthBlockFromTendermint(
	block *tmtypes.Block,
	fullTx bool,
) (map[string]interface{}, error) {
	ethRPCTxs := []interface{}{}

	ctx := types.ContextWithHeight(block.Height)

	baseFee, err := e.BaseFee()
	if err != nil {
		return nil, err
	}

	for i, txBz := range block.Txs {
		tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			e.logger.Debug("failed to decode transaction in block", "height", block.Height, "error", err.Error())
			continue
		}

		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			tx := ethMsg.AsTransaction()

			if !fullTx {
				hash := tx.Hash()
				ethRPCTxs = append(ethRPCTxs, hash)
				continue
			}

			rpcTx, err := types.NewRPCTransaction(
				tx,
				common.BytesToHash(block.Hash()),
				uint64(block.Height),
				uint64(i),
				baseFee,
			)
			if err != nil {
				e.logger.Debug("NewTransactionFromData for receipt failed", "hash", tx.Hash().Hex(), "error", err.Error())
				continue
			}
			ethRPCTxs = append(ethRPCTxs, rpcTx)
		}
	}

	bloom, err := e.BlockBloom(&block.Height)
	if err != nil {
		e.logger.Debug("failed to query BlockBloom", "height", block.Height, "error", err.Error())
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(block.Header.ProposerAddress).String(),
	}

	res, err := e.queryClient.ValidatorAccount(ctx, req)
	if err != nil {
		e.logger.Debug(
			"failed to query validator operator address",
			"height", block.Height,
			"cons-address", req.ConsAddress,
			"error", err.Error(),
		)
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(res.AccountAddress)
	if err != nil {
		return nil, err
	}

	validatorAddr := common.BytesToAddress(addr)

	gasLimit, err := types.BlockMaxGasFromConsensusParams(ctx, e.clientCtx)
	if err != nil {
		e.logger.Error("failed to query consensus params", "error", err.Error())
	}

	resBlockResult, err := e.clientCtx.Client.BlockResults(e.ctx, &block.Height)
	if err != nil {
		e.logger.Debug("EthBlockFromTendermint block result not found", "height", block.Height, "error", err.Error())
		return nil, err
	}

	gasUsed := uint64(0)

	for _, txsResult := range resBlockResult.TxsResults {
		gasUsed += uint64(txsResult.GetGasUsed())
	}

	formattedBlock := types.FormatBlock(
		block.Header, block.Size(),
		gasLimit, new(big.Int).SetUint64(gasUsed),
		ethRPCTxs, bloom, validatorAddr, baseFee,
	)
	return formattedBlock, nil
}

// CurrentHeader returns the latest block header
func (e *EVMBackend) CurrentHeader() *ethtypes.Header {
	header, _ := e.HeaderByNumber(types.EthLatestBlockNumber)
	return header
}

// HeaderByNumber returns the block header identified by height.
func (e *EVMBackend) HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			return nil, errors.Errorf("incorrect block height: %d", height)
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		e.logger.Debug("HeaderByNumber failed")
		return nil, err
	}

	bloom, err := e.BlockBloom(&resBlock.Block.Height)
	if err != nil {
		e.logger.Debug("HeaderByNumber BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := e.BaseFee()
	if err != nil {
		e.logger.Debug("HeaderByNumber BaseFee failed", "height", resBlock.Block.Height, "error", err.Error())
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header, baseFee)
	ethHeader.Bloom = bloom
	return ethHeader, nil
}

// HeaderByHash returns the block header identified by hash.
func (e *EVMBackend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, blockHash.Bytes())
	if err != nil {
		e.logger.Debug("HeaderByHash failed", "hash", blockHash.Hex())
		return nil, err
	}

	if resBlock.Block == nil {
		return nil, errors.Errorf("block not found for hash %s", blockHash.Hex())
	}

	bloom, err := e.BlockBloom(&resBlock.Block.Height)
	if err != nil {
		e.logger.Debug("HeaderByHash BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := e.BaseFee()
	if err != nil {
		e.logger.Debug("HeaderByHash BaseFee failed", "height", resBlock.Block.Height, "error", err.Error())
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header, baseFee)
	ethHeader.Bloom = bloom
	return ethHeader, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
// It returns an error if there's an encoding error.
// If no logs are found for the tx hash, the error is nil.
func (e *EVMBackend) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	tx, err := e.GetTxByEthHash(txHash)
	if err != nil {
		return nil, err
	}

	return TxLogsFromEvents(tx.TxResult.Events)
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *EVMBackend) PendingTransactions() ([]*sdk.Tx, error) {
	res, err := e.clientCtx.Client.UnconfirmedTxs(e.ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*sdk.Tx, 0, len(res.Txs))
	for _, txBz := range res.Txs {
		tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			return nil, err
		}
		result = append(result, &tx)
	}

	return result, nil
}

// GetLogsByHeight returns all the logs from all the ethereum transactions in a block.
func (e *EVMBackend) GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	blockRes, err := e.clientCtx.Client.BlockResults(e.ctx, height)
	if err != nil {
		return nil, err
	}

	blockLogs := [][]*ethtypes.Log{}
	for _, txResult := range blockRes.TxsResults {
		logs, err := TxLogsFromEvents(txResult.Events)
		if err != nil {
			return nil, err
		}

		blockLogs = append(blockLogs, logs)
	}

	return blockLogs, nil
}

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (e *EVMBackend) GetLogs(hash common.Hash) ([][]*ethtypes.Log, error) {
	block, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}
	return e.GetLogsByHeight(&block.Block.Header.Height)
}

func (e *EVMBackend) GetLogsByNumber(blockNum types.BlockNumber) ([][]*ethtypes.Log, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := e.BlockNumber()

	switch blockNum {
	case types.EthLatestBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthEarliestBlockNumber:
		height = 1
	default:
		if blockNum < 0 {
			return nil, errors.Errorf("incorrect block height: %d", height)
		}
	}

	return e.GetLogsByHeight(&height)
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (e *EVMBackend) BloomStatus() (uint64, uint64) {
	return 4096, 0
}

// GetCoinbase is the address that staking rewards will be send to (alias for Etherbase).
func (e *EVMBackend) GetCoinbase() (sdk.AccAddress, error) {
	node, err := e.clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	status, err := node.Status(e.ctx)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(status.ValidatorInfo.Address).String(),
	}

	res, err := e.queryClient.ValidatorAccount(e.ctx, req)
	if err != nil {
		return nil, err
	}

	address, _ := sdk.AccAddressFromBech32(res.AccountAddress)
	return address, nil
}

// GetTransactionByHash returns the Ethereum format transaction identified by Ethereum transaction hash
func (e *EVMBackend) GetTransactionByHash(txHash common.Hash) (*types.RPCTransaction, error) {
	res, err := e.GetTxByEthHash(txHash)
	if err != nil {
		// try to find tx in mempool
		txs, err := e.PendingTransactions()
		if err != nil {
			e.logger.Debug("tx not found", "hash", txHash.Hex(), "error", err.Error())
			return nil, nil
		}

		for _, tx := range txs {
			msg, err := evmtypes.UnwrapEthereumMsg(tx)
			if err != nil {
				// not ethereum tx
				continue
			}

			if msg.Hash == txHash.Hex() {
				rpctx, err := types.NewTransactionFromMsg(
					msg,
					common.Hash{},
					uint64(0),
					uint64(0),
					e.chainID,
				)
				if err != nil {
					return nil, err
				}
				return rpctx, nil
			}
		}

		e.logger.Debug("tx not found", "hash", txHash.Hex())
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

	return types.NewTransactionFromMsg(
		msg,
		common.BytesToHash(resBlock.Block.Hash()),
		uint64(res.Height),
		uint64(res.Index),
		e.chainID,
	)
}

// GetTxByEthHash uses `/tx_query` to find transaction by ethereum tx hash
// TODO: Don't need to convert once hashing is fixed on Tendermint
// https://github.com/tendermint/tendermint/issues/6539
func (e *EVMBackend) GetTxByEthHash(hash common.Hash) (*tmrpctypes.ResultTx, error) {
	query := fmt.Sprintf("%s.%s='%s'", evmtypes.TypeMsgEthereumTx, evmtypes.AttributeKeyEthereumTxHash, hash.Hex())
	resTxs, err := e.clientCtx.Client.TxSearch(e.ctx, query, false, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if len(resTxs.Txs) == 0 {
		return nil, errors.Errorf("ethereum tx not found for hash %s", hash.Hex())
	}
	return resTxs.Txs[0], nil
}

func (e *EVMBackend) SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	_, err := e.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	if err != nil {
		e.logger.Error("failed to find key in keyring", "address", args.From, "error", err.Error())
		return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	args, err = e.SetTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	msg := args.ToTransaction()
	if err := msg.ValidateBasic(); err != nil {
		e.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	bn, err := e.BlockNumber()
	if err != nil {
		e.logger.Debug("failed to fetch latest block number", "error", err.Error())
		return common.Hash{}, err
	}

	signer := ethtypes.MakeSigner(e.ChainConfig(), new(big.Int).SetUint64(uint64(bn)))

	// Sign transaction
	if err := msg.Sign(signer, e.clientCtx.Keyring); err != nil {
		e.logger.Debug("failed to sign tx", "error", err.Error())
		return common.Hash{}, err
	}

	// Assemble transaction from fields
	builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		e.logger.Error("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		e.logger.Error("codectypes.NewAnyWithValue failed to pack an obvious value", "error", err.Error())
		return common.Hash{}, err
	}

	builder.SetExtensionOptions(option)
	if err = builder.SetMsgs(msg); err != nil {
		e.logger.Error("builder.SetMsgs failed", "error", err.Error())
	}

	// Query params to use the EVM denomination
	res, err := e.queryClient.QueryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		e.logger.Error("failed to query evm params", "error", err.Error())
		return common.Hash{}, err
	}

	txData, err := evmtypes.UnpackTxData(msg.Data)
	if err != nil {
		e.logger.Error("failed to unpack tx data", "error", err.Error())
		return common.Hash{}, err
	}

	fees := sdk.Coins{sdk.NewCoin(res.Params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())

	// Encode transaction by default Tx encoder
	txEncoder := e.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		e.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	txHash := msg.AsTransaction().Hash()

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if err != nil || rsp.Code != 0 {
		if err == nil {
			err = errors.New(rsp.RawLog)
		}
		e.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	// Return transaction hash
	return txHash, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (e *EVMBackend) EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error) {
	blockNr := types.EthPendingBlockNumber
	if blockNrOptional != nil {
		blockNr = *blockNrOptional
	}

	bz, err := json.Marshal(&args)
	if err != nil {
		return 0, err
	}

	baseFee, err := e.BaseFee()
	if err != nil {
		return 0, err
	}

	var bf *sdk.Int
	if baseFee != nil {
		aux := sdk.NewIntFromBigInt(baseFee)
		bf = &aux
	}

	req := evmtypes.EthCallRequest{
		Args:    bz,
		GasCap:  e.RPCGasCap(),
		BaseFee: bf,
	}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	res, err := e.queryClient.EstimateGas(types.ContextWithHeight(blockNr.Int64()), &req)
	if err != nil {
		return 0, err
	}
	return hexutil.Uint64(res.Gas), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *EVMBackend) GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error) {
	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := e.clientCtx.AccountRetriever

	err := accRet.EnsureExists(e.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		n := hexutil.Uint64(0)
		return &n, nil
	}

	includePending := blockNum == types.EthPendingBlockNumber
	nonce, err := e.getAccountNonce(address, includePending, blockNum.Int64(), e.logger)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
}

// RPCGasCap is the global gas cap for eth-call variants.
func (e *EVMBackend) RPCGasCap() uint64 {
	return e.cfg.JSONRPC.GasCap
}

// RPCMinGasPrice returns the minimum gas price for a transaction obtained from
// the node config. If set value is 0, it will default to 20.

func (e *EVMBackend) RPCMinGasPrice() int64 {
	evmParams, err := e.queryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		return ethermint.DefaultGasPrice
	}

	minGasPrice := e.cfg.GetMinGasPrices()
	amt := minGasPrice.AmountOf(evmParams.Params.EvmDenom).TruncateInt64()
	if amt == 0 {
		return ethermint.DefaultGasPrice
	}

	return amt
}

// ChainConfig return the ethereum chain configuration
func (e *EVMBackend) ChainConfig() *params.ChainConfig {
	params, err := e.queryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		return nil
	}

	return params.Params.ChainConfig.EthereumConfig(e.chainID)
}

// SuggestGasTipCap returns the suggested tip cap
func (e *EVMBackend) SuggestGasTipCap() (*big.Int, error) {
	out := new(big.Int).SetInt64(e.RPCMinGasPrice())
	return out, nil
}

// BaseFee returns the base fee tracked by the Fee Market module. If the base fee is not enabled,
// it returns the initial base fee amount.
func (e *EVMBackend) BaseFee() (*big.Int, error) {
	res, err := e.queryClient.FeeMarket.BaseFee(e.ctx, &feemarkettypes.QueryBaseFeeRequest{})
	if err != nil {
		return nil, err
	}

	if res.BaseFee != nil {
		return res.BaseFee.BigInt(), nil
	}

	resParams, err := e.queryClient.FeeMarket.Params(e.ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	baseFee := big.NewInt(resParams.Params.InitialBaseFee)
	return baseFee, nil
}

// GetFilteredBlocks returns the block height list match the given bloom filters.
func (e *EVMBackend) GetFilteredBlocks(
	from int64,
	to int64,
	filters [][]filters.BloomIV,
	filterAddresses bool,
) ([]int64, error) {
	matchedBlocks := make([]int64, 0)

BLOCKS:
	for height := from; height <= to; height++ {
		if err := e.ctx.Err(); err != nil {
			e.logger.Error("EVMBackend context error", "err", err)
			return nil, err
		}

		h := height
		bloom, err := e.BlockBloom(&h)
		if err != nil {
			e.logger.Error("retrieve header failed", "blockHeight", height, "err", err)
			return nil, err
		}

		for i, filter := range filters {
			// filter the header bloom with the addresses
			if filterAddresses && i == 0 {
				if !checkMatches(bloom, filter) {
					continue BLOCKS
				}

				// the filter doesn't have any topics
				if len(filters) == 1 {
					matchedBlocks = append(matchedBlocks, height)
					continue BLOCKS
				}
				continue
			}

			// filter the bloom with topics
			if len(filter) > 0 && !checkMatches(bloom, filter) {
				continue BLOCKS
			}
		}
		matchedBlocks = append(matchedBlocks, height)
	}

	return matchedBlocks, nil
}

// checkMatches revised the function from
// https://github.com/ethereum/go-ethereum/blob/401354976bb44f0ad4455ca1e0b5c0dc31d9a5f5/core/types/bloom9.go#L88
func checkMatches(bloom ethtypes.Bloom, filter []filters.BloomIV) bool {
	for _, bloomIV := range filter {
		if bloomIV.V[0] == bloomIV.V[0]&bloom[bloomIV.I[0]] &&
			bloomIV.V[1] == bloomIV.V[1]&bloom[bloomIV.I[1]] &&
			bloomIV.V[2] == bloomIV.V[2]&bloom[bloomIV.I[2]] {
			return true
		}
	}
	return false
}

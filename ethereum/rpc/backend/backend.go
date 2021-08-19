package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/accounts/keystore"

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

	"github.com/tharsis/ethermint/ethereum/rpc/types"
	"github.com/tharsis/ethermint/server/config"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// Backend implements the functionality shared within namespaces.
// Implemented by EVMBackend.
type Backend interface {
	BlockNumber() (hexutil.Uint64, error)
	GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	PendingTransactions() ([]*sdk.Tx, error)
	GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error)
	GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error)
	SendTransaction(args types.SendTxArgs) (common.Hash, error)
	GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error)
	BloomStatus() (uint64, uint64)
	GetCoinbase() (sdk.AccAddress, error)
	EstimateGas(args evmtypes.CallArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error)
	RPCGasCap() uint64
}

var _ Backend = (*EVMBackend)(nil)

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

	appConf, err := config.ParseConfig(ctx.Viper)
	if err != nil {
		panic(err)
	}

	return &EVMBackend{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: types.NewQueryClient(clientCtx),
		logger:      logger.With("module", "evm-backend"),
		chainID:     chainID,
		cfg:         *appConf,
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
			err := errors.Errorf("incorrect block height: %d", height)
			return nil, err
		} else if height > int64(currentBlockNumber) {
			return nil, nil
		}
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		// e.logger.Debug("GetBlockByNumber safely bumping down from %d to latest", height)
		if resBlock, err = e.clientCtx.Client.Block(e.ctx, nil); err != nil {
			e.logger.Debug("GetBlockByNumber failed to get latest block", "error", err.Error())
			return nil, nil
		}
	}

	if resBlock.Block == nil {
		e.logger.Debug("GetBlockByNumber block not found", "height", height)
		return nil, nil
	}

	res, err := e.EthBlockFromTendermint(resBlock.Block, fullTx)
	if err != nil {
		e.logger.Debug("EthBlockFromTendermint failed", "height", height, "error", err.Error())
	}

	return res, err
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

// EthBlockFromTendermint returns a JSON-RPC compatible Ethereum block from a given Tendermint block.
func (e *EVMBackend) EthBlockFromTendermint(
	block *tmtypes.Block,
	fullTx bool,
) (map[string]interface{}, error) {

	gasUsed := uint64(0)

	ethRPCTxs := []interface{}{}

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

			// Todo: gasUsed does not consider the refund gas so it is incorrect, we need to extract it from the result
			gasUsed += ethMsg.GetGas()
			hash := ethMsg.AsTransaction().Hash()
			if !fullTx {
				ethRPCTxs = append(ethRPCTxs, hash)
				continue
			}

			// get full transaction from message data
			from, err := ethMsg.GetSender(e.chainID)
			if err != nil {
				e.logger.Debug("failed to get sender from already included transaction", "hash", hash.Hex(), "error", err.Error())
				from = common.HexToAddress(ethMsg.From)
			}

			txData, err := evmtypes.UnpackTxData(ethMsg.Data)
			if err != nil {
				e.logger.Debug("decoding failed", "error", err.Error())
				return nil, fmt.Errorf("failed to unpack tx data: %w", err)
			}

			ethTx, err := types.NewTransactionFromData(
				txData,
				from,
				hash,
				common.BytesToHash(block.Hash()),
				uint64(block.Height),
				uint64(i),
			)
			if err != nil {
				e.logger.Debug("NewTransactionFromData for receipt failed", "hash", hash.Hex(), "error", err.Error())
				continue
			}
			ethRPCTxs = append(ethRPCTxs, ethTx)
		}
	}

	blockBloomResp, err := e.queryClient.BlockBloom(types.ContextWithHeight(block.Height), &evmtypes.QueryBlockBloomRequest{Height: block.Height})
	if err != nil {
		e.logger.Debug("failed to query BlockBloom", "height", block.Height, "error", err.Error())

		blockBloomResp = &evmtypes.QueryBlockBloomResponse{Bloom: ethtypes.Bloom{}.Bytes()}
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(block.Header.ProposerAddress).String(),
	}

	res, err := e.queryClient.ValidatorAccount(e.ctx, req)
	if err != nil {
		e.logger.Debug("failed to query validator operator address", "cons-address", req.ConsAddress, "error", err.Error())
		return nil, err
	}

	addr, err := sdk.AccAddressFromBech32(res.AccountAddress)
	if err != nil {
		return nil, err
	}

	validatorAddr := common.BytesToAddress(addr)

	bloom := ethtypes.BytesToBloom(blockBloomResp.Bloom)

	gasLimit, err := types.BlockMaxGasFromConsensusParams(types.ContextWithHeight(block.Height), e.clientCtx)
	if err != nil {
		e.logger.Error("failed to query consensus params", "error", err.Error())
	}
	formattedBlock := types.FormatBlock(block.Header, block.Size(), gasLimit, new(big.Int).SetUint64(gasUsed), ethRPCTxs, bloom, validatorAddr)
	return formattedBlock, nil
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

	req := &evmtypes.QueryBlockBloomRequest{Height: resBlock.Block.Height}

	blockBloomResp, err := e.queryClient.BlockBloom(types.ContextWithHeight(resBlock.Block.Height), req)
	if err != nil {
		e.logger.Debug("HeaderByNumber BlockBloom failed", "height", resBlock.Block.Height)
		blockBloomResp = &evmtypes.QueryBlockBloomResponse{Bloom: ethtypes.Bloom{}.Bytes()}
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(blockBloomResp.Bloom)
	return ethHeader, nil
}

// HeaderByHash returns the block header identified by hash.
func (e *EVMBackend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, blockHash.Bytes())
	if err != nil {
		e.logger.Debug("HeaderByHash failed", "hash", blockHash.Hex())
		return nil, err
	}

	req := &evmtypes.QueryBlockBloomRequest{Height: resBlock.Block.Height}

	blockBloomResp, err := e.queryClient.BlockBloom(types.ContextWithHeight(resBlock.Block.Height), req)
	if err != nil {
		e.logger.Debug("HeaderByHash BlockBloom failed", "height", resBlock.Block.Height)
		blockBloomResp = &evmtypes.QueryBlockBloomResponse{Bloom: ethtypes.Bloom{}.Bytes()}
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header)
	ethHeader.Bloom = ethtypes.BytesToBloom(blockBloomResp.Bloom)
	return ethHeader, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
// It returns an error if there's an encoding error.
// If no logs are found for the tx hash, the error is nil.
func (e *EVMBackend) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	req := &evmtypes.QueryTxLogsRequest{
		Hash: txHash.String(),
	}

	res, err := e.queryClient.TxLogs(e.ctx, req)
	if err != nil {
		e.logger.Debug("TxLogs failed", "tx-hash", req.Hash)
		return nil, err
	}

	return evmtypes.LogsToEthereum(res.Logs), nil
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

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (e *EVMBackend) GetLogs(blockHash common.Hash) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	req := &evmtypes.QueryBlockLogsRequest{
		Hash: blockHash.String(),
	}

	res, err := e.queryClient.BlockLogs(e.ctx, req)
	if err != nil {
		e.logger.Debug("BlockLogs failed", "hash", req.Hash)
		return nil, err
	}

	var blockLogs = [][]*ethtypes.Log{}
	for _, txLog := range res.TxLogs {
		blockLogs = append(blockLogs, txLog.EthLogs())
	}

	return blockLogs, nil
}

// This is very brittle, see: https://github.com/tendermint/tendermint/issues/4740
var regexpMissingHeight = regexp.MustCompile(`height \d+ (must be less than or equal to|is not available)`)

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

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &height)
	if err != nil {
		if regexpMissingHeight.MatchString(err.Error()) {
			return [][]*ethtypes.Log{}, nil
		}

		e.logger.Debug("failed to query block", "height", height)
		return nil, err
	}

	return e.GetLogs(common.BytesToHash(resBlock.BlockID.Hash))
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

func (e *EVMBackend) SendTransaction(args types.SendTxArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	_, err := e.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	if err != nil {
		e.logger.Error("failed to find key in keyring", "address", args.From, "error", err.Error())
		return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	args, err = e.setTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	msg := args.ToTransaction()

	if err := msg.ValidateBasic(); err != nil {
		e.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	// TODO: get from chain config
	signer := ethtypes.LatestSignerForChainID(args.ChainID.ToInt())

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
	err = builder.SetMsgs(msg)
	if err != nil {
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
func (e *EVMBackend) EstimateGas(args evmtypes.CallArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error) {
	blockNr := types.EthPendingBlockNumber
	if blockNrOptional != nil {
		blockNr = *blockNrOptional
	}

	bz, err := json.Marshal(&args)
	if err != nil {
		return 0, err
	}

	req := evmtypes.EthCallRequest{Args: bz, GasCap: e.RPCGasCap()}

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

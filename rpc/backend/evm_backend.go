package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdkcrypto "github.com/cosmos/cosmos-sdk/crypto"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/ethereum/eip712"
	"github.com/evmos/ethermint/rpc/types"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

var bAttributeKeyEthereumBloom = []byte(evmtypes.AttributeKeyEthereumBloom)

// ClientCtx returns client context
func (b *Backend) ClientCtx() client.Context {
	return b.clientCtx
}

// BlockByNumber returns the block identified by number.
func (b *Backend) BlockByNumber(blockNum types.BlockNumber) (*ethtypes.Block, error) {
	resBlock, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		// block not found
		return nil, fmt.Errorf("block not found for height %d", blockNum)
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, fmt.Errorf("block result not found for height %d", resBlock.Block.Height)
	}

	return b.EthBlockFromTm(resBlock, blockRes)
}

// BlockByHash returns the block identified by hash.
func (b *Backend) BlockByHash(hash common.Hash) (*ethtypes.Block, error) {
	resBlock, err := b.GetTendermintBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		return nil, fmt.Errorf("block not found for hash %s", hash)
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, fmt.Errorf("block result not found for hash %s", hash)
	}

	return b.EthBlockFromTm(resBlock, blockRes)
}

func (b *Backend) EthBlockFromTm(resBlock *tmrpctypes.ResultBlock, blockRes *tmrpctypes.ResultBlockResults) (*ethtypes.Block, error) {
	block := resBlock.Block
	height := block.Height
	bloom, err := b.BlockBloom(blockRes)
	if err != nil {
		b.logger.Debug("HeaderByNumber BlockBloom failed", "height", height)
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle error for pruned node and log
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", height, "error", err)
	}

	ethHeader := types.EthHeaderFromTendermint(block.Header, bloom, baseFee)

	resBlockResult, err := b.GetTendermintBlockResultByNumber(&block.Height)
	if err != nil {
		return nil, err
	}

	msgs := b.GetEthereumMsgsFromTendermintBlock(resBlock, resBlockResult)

	txs := make([]*ethtypes.Transaction, len(msgs))
	for i, ethMsg := range msgs {
		txs[i] = ethMsg.AsTransaction()
	}

	// TODO: add tx receipts
	ethBlock := ethtypes.NewBlock(ethHeader, txs, nil, nil, nil)
	return ethBlock, nil
}

// GetTendermintBlockByNumber returns a Tendermint formatted block for a given
// block number
func (b *Backend) GetTendermintBlockByNumber(blockNum types.BlockNumber) (*tmrpctypes.ResultBlock, error) {
	height := blockNum.Int64()
	if height <= 0 {
		// fetch the latest block number from the app state, more accurate than the tendermint block store state.
		n, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}
		height = int64(n)
	}
	resBlock, err := b.clientCtx.Client.Block(b.ctx, &height)
	if err != nil {
		b.logger.Debug("tendermint client failed to get block", "height", height, "error", err.Error())
		return nil, err
	}

	if resBlock.Block == nil {
		b.logger.Debug("GetTendermintBlockByNumber block not found", "height", height)
		return nil, nil
	}

	return resBlock, nil
}

// GetTendermintBlockResultByNumber returns a Tendermint-formatted block result by block number
func (b *Backend) GetTendermintBlockResultByNumber(height *int64) (*tmrpctypes.ResultBlockResults, error) {
	return b.clientCtx.Client.BlockResults(b.ctx, height)
}

// GetTendermintBlockByHash returns a Tendermint format block by block number
func (b *Backend) GetTendermintBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, blockHash.Bytes())
	if err != nil {
		b.logger.Debug("tendermint client failed to get block", "blockHash", blockHash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		b.logger.Debug("GetTendermintBlockByHash block not found", "blockHash", blockHash.Hex())
		return nil, nil
	}

	return resBlock, nil
}

// BlockBloom query block bloom filter from block results
func (b *Backend) BlockBloom(blockRes *tmrpctypes.ResultBlockResults) (ethtypes.Bloom, error) {
	for _, event := range blockRes.EndBlockEvents {
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

// EthBlockFromTendermint returns a JSON-RPC compatible Ethereum block from a
// given Tendermint block and its block result.
func (b *Backend) EthBlockFromTendermint(
	resBlock *tmrpctypes.ResultBlock,
	blockRes *tmrpctypes.ResultBlockResults,
	fullTx bool,
) (map[string]interface{}, error) {
	ethRPCTxs := []interface{}{}
	block := resBlock.Block

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", block.Height, "error", err)
	}

	msgs := b.GetEthereumMsgsFromTendermintBlock(resBlock, blockRes)
	for txIndex, ethMsg := range msgs {
		if !fullTx {
			hash := common.HexToHash(ethMsg.Hash)
			ethRPCTxs = append(ethRPCTxs, hash)
			continue
		}

		tx := ethMsg.AsTransaction()
		rpcTx, err := types.NewRPCTransaction(
			tx,
			common.BytesToHash(block.Hash()),
			uint64(block.Height),
			uint64(txIndex),
			baseFee,
		)
		if err != nil {
			b.logger.Debug("NewTransactionFromData for receipt failed", "hash", tx.Hash().Hex(), "error", err.Error())
			continue
		}
		ethRPCTxs = append(ethRPCTxs, rpcTx)
	}

	bloom, err := b.BlockBloom(blockRes)
	if err != nil {
		b.logger.Debug("failed to query BlockBloom", "height", block.Height, "error", err.Error())
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(block.Header.ProposerAddress).String(),
	}

	var validatorAccAddr sdk.AccAddress

	ctx := types.ContextWithHeight(block.Height)
	res, err := b.queryClient.ValidatorAccount(ctx, req)
	if err != nil {
		b.logger.Debug(
			"failed to query validator operator address",
			"height", block.Height,
			"cons-address", req.ConsAddress,
			"error", err.Error(),
		)
		// use zero address as the validator operator address
		validatorAccAddr = sdk.AccAddress(common.Address{}.Bytes())
	} else {
		validatorAccAddr, err = sdk.AccAddressFromBech32(res.AccountAddress)
		if err != nil {
			return nil, err
		}
	}

	validatorAddr := common.BytesToAddress(validatorAccAddr)

	gasLimit, err := types.BlockMaxGasFromConsensusParams(ctx, b.clientCtx, block.Height)
	if err != nil {
		b.logger.Error("failed to query consensus params", "error", err.Error())
	}

	gasUsed := uint64(0)

	for _, txsResult := range blockRes.TxsResults {
		// workaround for cosmos-sdk bug. https://github.com/cosmos/cosmos-sdk/issues/10832
		if ShouldIgnoreGasUsed(txsResult) {
			// block gas limit has exceeded, other txs must have failed with same reason.
			break
		}
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
func (b *Backend) CurrentHeader() *ethtypes.Header {
	header, _ := b.HeaderByNumber(types.EthLatestBlockNumber)
	return header
}

// HeaderByNumber returns the block header identified by height.
func (b *Backend) HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error) {
	resBlock, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	if resBlock == nil {
		return nil, errors.Errorf("block not found for height %d", blockNum)
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, fmt.Errorf("block result not found for height %d", resBlock.Block.Height)
	}

	bloom, err := b.BlockBloom(blockRes)
	if err != nil {
		b.logger.Debug("HeaderByNumber BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", resBlock.Block.Height, "error", err)
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header, bloom, baseFee)
	return ethHeader, nil
}

// GetBlockNumberByHash returns the block height of given block hash
func (b *Backend) GetBlockNumberByHash(blockHash common.Hash) (*big.Int, error) {
	resBlock, err := b.GetTendermintBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		return nil, errors.Errorf("block not found for hash %s", blockHash.Hex())
	}
	return big.NewInt(resBlock.Block.Height), nil
}

// HeaderByHash returns the block header identified by hash.
func (b *Backend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := b.GetTendermintBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		return nil, errors.Errorf("block not found for hash %s", blockHash.Hex())
	}

	blockRes, err := b.GetTendermintBlockResultByNumber(&resBlock.Block.Height)
	if err != nil {
		return nil, errors.Errorf("block result not found for height %d", resBlock.Block.Height)
	}

	bloom, err := b.BlockBloom(blockRes)
	if err != nil {
		b.logger.Debug("HeaderByHash BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", resBlock.Block.Height, "error", err)
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header, bloom, baseFee)
	return ethHeader, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (b *Backend) PendingTransactions() ([]*sdk.Tx, error) {
	res, err := b.clientCtx.Client.UnconfirmedTxs(b.ctx, nil)
	if err != nil {
		return nil, err
	}

	result := make([]*sdk.Tx, 0, len(res.Txs))
	for _, txBz := range res.Txs {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			return nil, err
		}
		result = append(result, &tx)
	}

	return result, nil
}

// GetLogsByHeight returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error) {
	// NOTE: we query the state in case the tx result logs are not persisted after an upgrade.
	blockRes, err := b.GetTendermintBlockResultByNumber(height)
	if err != nil {
		return nil, err
	}

	return GetLogsFromBlockResults(blockRes)
}

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogs(hash common.Hash) ([][]*ethtypes.Log, error) {
	resBlock, err := b.GetTendermintBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	if resBlock == nil {
		return nil, errors.Errorf("block not found for hash %s", hash)
	}

	return b.GetLogsByHeight(&resBlock.Block.Header.Height)
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (b *Backend) BloomStatus() (uint64, uint64) {
	return 4096, 0
}

// GetCoinbase is the address that staking rewards will be send to (alias for Etherbase).
func (b *Backend) GetCoinbase() (sdk.AccAddress, error) {
	node, err := b.clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	status, err := node.Status(b.ctx)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(status.ValidatorInfo.Address).String(),
	}

	res, err := b.queryClient.ValidatorAccount(b.ctx, req)
	if err != nil {
		return nil, err
	}

	address, _ := sdk.AccAddressFromBech32(res.AccountAddress)
	return address, nil
}

// GetTxByEthHash uses `/tx_query` to find transaction by ethereum tx hash
// TODO: Don't need to convert once hashing is fixed on Tendermint
// https://github.com/tendermint/tendermint/issues/6539
func (b *Backend) GetTxByEthHash(hash common.Hash) (*tmrpctypes.ResultTx, error) {
	query := fmt.Sprintf("%s.%s='%s'", evmtypes.TypeMsgEthereumTx, evmtypes.AttributeKeyEthereumTxHash, hash.Hex())
	resTxs, err := b.clientCtx.Client.TxSearch(b.ctx, query, false, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if len(resTxs.Txs) == 0 {
		return nil, errors.Errorf("ethereum tx not found for hash %s", hash.Hex())
	}
	return resTxs.Txs[0], nil
}

// GetTxByTxIndex uses `/tx_query` to find transaction by tx index of valid ethereum txs
func (b *Backend) GetTxByTxIndex(height int64, index uint) (*tmrpctypes.ResultTx, error) {
	query := fmt.Sprintf("tx.height=%d AND %s.%s=%d",
		height, evmtypes.TypeMsgEthereumTx,
		evmtypes.AttributeKeyTxIndex, index,
	)
	resTxs, err := b.clientCtx.Client.TxSearch(b.ctx, query, false, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if len(resTxs.Txs) == 0 {
		return nil, errors.Errorf("ethereum tx not found for block %d index %d", height, index)
	}
	return resTxs.Txs[0], nil
}

func (b *Backend) SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error) {
	// Look up the wallet containing the requested signer
	_, err := b.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	if err != nil {
		b.logger.Error("failed to find key in keyring", "address", args.From, "error", err.Error())
		return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	args, err = b.SetTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	msg := args.ToTransaction()
	if err := msg.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	bn, err := b.BlockNumber()
	if err != nil {
		b.logger.Debug("failed to fetch latest block number", "error", err.Error())
		return common.Hash{}, err
	}

	signer := ethtypes.MakeSigner(b.ChainConfig(), new(big.Int).SetUint64(uint64(bn)))

	// Sign transaction
	if err := msg.Sign(signer, b.clientCtx.Keyring); err != nil {
		b.logger.Debug("failed to sign tx", "error", err.Error())
		return common.Hash{}, err
	}

	// Query params to use the EVM denomination
	res, err := b.queryClient.QueryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		b.logger.Error("failed to query evm params", "error", err.Error())
		return common.Hash{}, err
	}

	// Assemble transaction from fields
	tx, err := msg.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), res.Params.EvmDenom)
	if err != nil {
		b.logger.Error("build cosmos tx failed", "error", err.Error())
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txEncoder := b.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(tx)
	if err != nil {
		b.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	ethTx := msg.AsTransaction()

	// check the local node config in case unprotected txs are disabled
	if !b.UnprotectedAllowed() && !ethTx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return common.Hash{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}

	txHash := ethTx.Hash()

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = sdkerrors.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}
	if err != nil {
		b.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	// Return transaction hash
	return txHash, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
func (b *Backend) EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error) {
	blockNr := types.EthPendingBlockNumber
	if blockNrOptional != nil {
		blockNr = *blockNrOptional
	}

	bz, err := json.Marshal(&args)
	if err != nil {
		return 0, err
	}

	req := evmtypes.EthCallRequest{
		Args:   bz,
		GasCap: b.RPCGasCap(),
	}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	res, err := b.queryClient.EstimateGas(types.ContextWithHeight(blockNr.Int64()), &req)
	if err != nil {
		return 0, err
	}
	return hexutil.Uint64(res.Gas), nil
}

// RPCGasCap is the global gas cap for eth-call variants.
func (b *Backend) RPCGasCap() uint64 {
	return b.cfg.JSONRPC.GasCap
}

// RPCEVMTimeout is the global evm timeout for eth-call variants.
func (b *Backend) RPCEVMTimeout() time.Duration {
	return b.cfg.JSONRPC.EVMTimeout
}

// RPCGasCap is the global gas cap for eth-call variants.
func (b *Backend) RPCTxFeeCap() float64 {
	return b.cfg.JSONRPC.TxFeeCap
}

// RPCFilterCap is the limit for total number of filters that can be created
func (b *Backend) RPCFilterCap() int32 {
	return b.cfg.JSONRPC.FilterCap
}

// RPCFeeHistoryCap is the limit for total number of blocks that can be fetched
func (b *Backend) RPCFeeHistoryCap() int32 {
	return b.cfg.JSONRPC.FeeHistoryCap
}

// RPCLogsCap defines the max number of results can be returned from single `eth_getLogs` query.
func (b *Backend) RPCLogsCap() int32 {
	return b.cfg.JSONRPC.LogsCap
}

// RPCBlockRangeCap defines the max block range allowed for `eth_getLogs` query.
func (b *Backend) RPCBlockRangeCap() int32 {
	return b.cfg.JSONRPC.BlockRangeCap
}

// RPCMinGasPrice returns the minimum gas price for a transaction obtained from
// the node config. If set value is 0, it will default to 20.

func (b *Backend) RPCMinGasPrice() int64 {
	evmParams, err := b.queryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		return ethermint.DefaultGasPrice
	}

	minGasPrice := b.cfg.GetMinGasPrices()
	amt := minGasPrice.AmountOf(evmParams.Params.EvmDenom).TruncateInt64()
	if amt == 0 {
		return ethermint.DefaultGasPrice
	}

	return amt
}

// ChainConfig returns the latest ethereum chain configuration
func (b *Backend) ChainConfig() *params.ChainConfig {
	params, err := b.queryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		return nil
	}

	return params.Params.ChainConfig.EthereumConfig(b.chainID)
}

// SuggestGasTipCap returns the suggested tip cap
// Although we don't support tx prioritization yet, but we return a positive value to help client to
// mitigate the base fee changes.
func (b *Backend) SuggestGasTipCap(baseFee *big.Int) (*big.Int, error) {
	if baseFee == nil {
		// london hardfork not enabled or feemarket not enabled
		return big.NewInt(0), nil
	}

	params, err := b.queryClient.FeeMarket.Params(b.ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	// calculate the maximum base fee delta in current block, assuming all block gas limit is consumed
	// ```
	// GasTarget = GasLimit / ElasticityMultiplier
	// Delta = BaseFee * (GasUsed - GasTarget) / GasTarget / Denominator
	// ```
	// The delta is at maximum when `GasUsed` is equal to `GasLimit`, which is:
	// ```
	// MaxDelta = BaseFee * (GasLimit - GasLimit / ElasticityMultiplier) / (GasLimit / ElasticityMultiplier) / Denominator
	//          = BaseFee * (ElasticityMultiplier - 1) / Denominator
	// ```
	maxDelta := baseFee.Int64() * (int64(params.Params.ElasticityMultiplier) - 1) / int64(params.Params.BaseFeeChangeDenominator)
	if maxDelta < 0 {
		// impossible if the parameter validation passed.
		maxDelta = 0
	}
	return big.NewInt(maxDelta), nil
}

// BaseFee returns the base fee tracked by the Fee Market module.
// If the base fee is not enabled globally, the query returns nil.
// If the London hard fork is not activated at the current height, the query will
// return nil.
func (b *Backend) BaseFee(blockRes *tmrpctypes.ResultBlockResults) (*big.Int, error) {
	// return BaseFee if London hard fork is activated and feemarket is enabled
	res, err := b.queryClient.BaseFee(types.ContextWithHeight(blockRes.Height), &evmtypes.QueryBaseFeeRequest{})
	if err != nil || res.BaseFee == nil {
		// we can't tell if it's london HF not enabled or the state is pruned,
		// in either case, we'll fallback to parsing from begin blocker event,
		// faster to iterate reversely
		for i := len(blockRes.BeginBlockEvents) - 1; i >= 0; i-- {
			evt := blockRes.BeginBlockEvents[i]
			if evt.Type == feemarkettypes.EventTypeFeeMarket && len(evt.Attributes) > 0 {
				baseFee, err := strconv.ParseInt(string(evt.Attributes[0].Value), 10, 64)
				if err == nil {
					return big.NewInt(baseFee), nil
				}
				break
			}
		}
		return nil, err
	}

	if res.BaseFee == nil {
		return nil, nil
	}

	return res.BaseFee.BigInt(), nil
}

// GlobalMinGasPrice returns MinGasPrice param from FeeMarket
func (b *Backend) GlobalMinGasPrice() (sdk.Dec, error) {
	res, err := b.queryClient.FeeMarket.Params(b.ctx, &feemarkettypes.QueryParamsRequest{})
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return res.Params.MinGasPrice, nil
}

// FeeHistory returns data relevant for fee estimation based on the specified range of blocks.
func (b *Backend) FeeHistory(
	userBlockCount rpc.DecimalOrHex, // number blocks to fetch, maximum is 100
	lastBlock rpc.BlockNumber, // the block to start search , to oldest
	rewardPercentiles []float64, // percentiles to fetch reward
) (*types.FeeHistoryResult, error) {
	blockEnd := int64(lastBlock)

	if blockEnd <= 0 {
		blockNumber, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}
		blockEnd = int64(blockNumber)
	}
	userBlockCountInt := int64(userBlockCount)
	maxBlockCount := int64(b.cfg.JSONRPC.FeeHistoryCap)
	if userBlockCountInt > maxBlockCount {
		return nil, fmt.Errorf("FeeHistory user block count %d higher than %d", userBlockCountInt, maxBlockCount)
	}
	blockStart := blockEnd - userBlockCountInt
	if blockStart < 0 {
		blockStart = 0
	}

	blockCount := blockEnd - blockStart

	oldestBlock := (*hexutil.Big)(big.NewInt(blockStart))

	// prepare space
	reward := make([][]*hexutil.Big, blockCount)
	rewardCount := len(rewardPercentiles)
	for i := 0; i < int(blockCount); i++ {
		reward[i] = make([]*hexutil.Big, rewardCount)
	}
	thisBaseFee := make([]*hexutil.Big, blockCount)
	thisGasUsedRatio := make([]float64, blockCount)

	// rewards should only be calculated if reward percentiles were included
	calculateRewards := rewardCount != 0

	// fetch block
	for blockID := blockStart; blockID < blockEnd; blockID++ {
		index := int32(blockID - blockStart)
		// eth block
		ethBlock, err := b.GetBlockByNumber(types.BlockNumber(blockID), true)
		if ethBlock == nil {
			return nil, err
		}

		// tendermint block
		tendermintblock, err := b.GetTendermintBlockByNumber(types.BlockNumber(blockID))
		if tendermintblock == nil {
			return nil, err
		}

		// tendermint block result
		tendermintBlockResult, err := b.GetTendermintBlockResultByNumber(&tendermintblock.Block.Height)
		if tendermintBlockResult == nil {
			b.logger.Debug("block result not found", "height", tendermintblock.Block.Height, "error", err.Error())
			return nil, err
		}

		oneFeeHistory := types.OneFeeHistory{}
		err = b.processBlock(tendermintblock, &ethBlock, rewardPercentiles, tendermintBlockResult, &oneFeeHistory)
		if err != nil {
			return nil, err
		}

		// copy
		thisBaseFee[index] = (*hexutil.Big)(oneFeeHistory.BaseFee)
		thisGasUsedRatio[index] = oneFeeHistory.GasUsedRatio
		if calculateRewards {
			for j := 0; j < rewardCount; j++ {
				reward[index][j] = (*hexutil.Big)(oneFeeHistory.Reward[j])
				if reward[index][j] == nil {
					reward[index][j] = (*hexutil.Big)(big.NewInt(0))
				}
			}
		}
	}

	feeHistory := types.FeeHistoryResult{
		OldestBlock:  oldestBlock,
		BaseFee:      thisBaseFee,
		GasUsedRatio: thisGasUsedRatio,
	}

	if calculateRewards {
		feeHistory.Reward = reward
	}

	return &feeHistory, nil
}

// GetEthereumMsgsFromTendermintBlock returns all real MsgEthereumTxs from a
// Tendermint block. It also ensures consistency over the correct txs indexes
// across RPC endpoints
func (b *Backend) GetEthereumMsgsFromTendermintBlock(
	resBlock *tmrpctypes.ResultBlock,
	blockRes *tmrpctypes.ResultBlockResults,
) []*evmtypes.MsgEthereumTx {
	var result []*evmtypes.MsgEthereumTx
	block := resBlock.Block

	txResults := blockRes.TxsResults

	for i, tx := range block.Txs {
		// Check if tx exists on EVM by cross checking with blockResults:
		//  - Include unsuccessful tx that exceeds block gas limit
		//  - Exclude unsuccessful tx with any other error but ExceedBlockGasLimit
		if !TxSuccessOrExceedsBlockGasLimit(txResults[i]) {
			b.logger.Debug("invalid tx result code", "cosmos-hash", hexutil.Encode(tx.Hash()))
			continue
		}

		tx, err := b.clientCtx.TxConfig.TxDecoder()(tx)
		if err != nil {
			b.logger.Debug("failed to decode transaction in block", "height", block.Height, "error", err.Error())
			continue
		}

		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			ethMsg.Hash = ethMsg.AsTransaction().Hash().Hex()
			result = append(result, ethMsg)
		}
	}

	return result
}

// UnprotectedAllowed returns the node configuration value for allowing
// unprotected transactions (i.e not replay-protected)
func (b Backend) UnprotectedAllowed() bool {
	return b.allowUnprotectedTxs
}

// getBlockNumber returns the BlockNumber from BlockNumberOrHash
func (b *Backend) GetBlockNumber(blockNrOrHash types.BlockNumberOrHash) (types.BlockNumber, error) {
	switch {
	case blockNrOrHash.BlockHash == nil && blockNrOrHash.BlockNumber == nil:
		return types.EthEarliestBlockNumber, fmt.Errorf("types BlockHash and BlockNumber cannot be both nil")
	case blockNrOrHash.BlockHash != nil:
		blockNumber, err := b.GetBlockNumberByHash(*blockNrOrHash.BlockHash)
		if err != nil {
			return types.EthEarliestBlockNumber, err
		}
		return types.NewBlockNumber(blockNumber), nil
	case blockNrOrHash.BlockNumber != nil:
		return *blockNrOrHash.BlockNumber, nil
	default:
		return types.EthEarliestBlockNumber, nil
	}
}

// getTransactionByBlockAndIndex is the common code shared by `GetTransactionByBlockNumberAndIndex` and `GetTransactionByBlockHashAndIndex`.
func (b *Backend) GetTransactionByBlockAndIndex(block *tmrpctypes.ResultBlock, idx hexutil.Uint) (*types.RPCTransaction, error) {
	blockRes, err := b.GetTendermintBlockResultByNumber(&block.Block.Height)
	if err != nil {
		return nil, nil
	}

	var msg *evmtypes.MsgEthereumTx
	// try /tx_search first
	res, err := b.GetTxByTxIndex(block.Block.Height, uint(idx))
	if err == nil {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(res.Tx)
		if err != nil {
			b.logger.Debug("invalid ethereum tx", "height", block.Block.Header, "index", idx)
			return nil, nil
		}

		parsedTxs, err := types.ParseTxResult(&res.TxResult)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tx events: %d, %v", idx, err)
		}

		parsedTx := parsedTxs.GetTxByTxIndex(int(idx))
		if parsedTx == nil {
			return nil, fmt.Errorf("ethereum tx not found in msgs: %d", idx)
		}

		var ok bool
		// msgIndex is inferred from tx events, should be within bound.
		msg, ok = tx.GetMsgs()[parsedTx.MsgIndex].(*evmtypes.MsgEthereumTx)
		if !ok {
			b.logger.Debug("invalid ethereum tx", "height", block.Block.Header, "index", idx)
			return nil, nil
		}
	} else {
		i := int(idx)
		ethMsgs := b.GetEthereumMsgsFromTendermintBlock(block, blockRes)
		if i >= len(ethMsgs) {
			b.logger.Debug("block txs index out of bound", "index", i)
			return nil, nil
		}

		msg = ethMsgs[i]
	}

	baseFee, err := b.BaseFee(blockRes)
	if err != nil {
		// handle the error for pruned node.
		b.logger.Error("failed to fetch Base Fee from prunned block. Check node prunning configuration", "height", block.Block.Height, "error", err)
	}

	return types.NewTransactionFromMsg(
		msg,
		common.BytesToHash(block.Block.Hash()),
		uint64(block.Block.Height),
		uint64(idx),
		baseFee,
	)
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (b *Backend) DoCall(
	args evmtypes.TransactionArgs, blockNr types.BlockNumber,
) (*evmtypes.MsgEthereumTxResponse, error) {
	bz, err := json.Marshal(&args)
	if err != nil {
		return nil, err
	}

	req := evmtypes.EthCallRequest{
		Args:   bz,
		GasCap: b.RPCGasCap(),
	}

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	ctx := types.ContextWithHeight(blockNr.Int64())
	timeout := b.RPCEVMTimeout()

	// Setup context so it may be canceled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	// Make sure the context is canceled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	res, err := b.queryClient.EthCall(ctx, &req)
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

func (b *Backend) Resend(args evmtypes.TransactionArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error) {
	if args.Nonce == nil {
		return common.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}

	args, err := b.SetTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	// The signer used should always be the 'latest' known one because we expect
	// signers to be backwards-compatible with old transactions.
	eip155ChainID, err := ethermint.ParseChainID(b.clientCtx.ChainID)
	if err != nil {
		return common.Hash{}, err
	}

	cfg := b.ChainConfig()
	if cfg == nil {
		cfg = evmtypes.DefaultChainConfig().EthereumConfig(eip155ChainID)
	}

	signer := ethtypes.LatestSigner(cfg)

	matchTx := args.ToTransaction().AsTransaction()

	// Before replacing the old transaction, ensure the _new_ transaction fee is reasonable.
	price := matchTx.GasPrice()
	if gasPrice != nil {
		price = gasPrice.ToInt()
	}
	gas := matchTx.Gas()
	if gasLimit != nil {
		gas = uint64(*gasLimit)
	}
	if err := types.CheckTxFee(price, gas, b.RPCTxFeeCap()); err != nil {
		return common.Hash{}, err
	}

	pending, err := b.PendingTransactions()
	if err != nil {
		return common.Hash{}, err
	}

	for _, tx := range pending {
		// FIXME does Resend api possible at all?  https://github.com/evmos/ethermint/issues/905
		p, err := evmtypes.UnwrapEthereumMsg(tx, common.Hash{})
		if err != nil {
			// not valid ethereum tx
			continue
		}

		pTx := p.AsTransaction()

		wantSigHash := signer.Hash(matchTx)
		pFrom, err := ethtypes.Sender(signer, pTx)
		if err != nil {
			continue
		}

		if pFrom == *args.From && signer.Hash(pTx) == wantSigHash {
			// Match. Re-sign and send the transaction.
			if gasPrice != nil && (*big.Int)(gasPrice).Sign() != 0 {
				args.GasPrice = gasPrice
			}
			if gasLimit != nil && *gasLimit != 0 {
				args.Gas = gasLimit
			}

			return b.SendTransaction(args) // TODO: this calls SetTxDefaults again, refactor to avoid calling it twice
		}
	}

	return common.Hash{}, fmt.Errorf("transaction %#x not found", matchTx.Hash())
}

func (b *Backend) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	// RLP decode raw transaction bytes
	tx := &ethtypes.Transaction{}
	if err := tx.UnmarshalBinary(data); err != nil {
		b.logger.Error("transaction decoding failed", "error", err.Error())
		return common.Hash{}, err
	}

	// check the local node config in case unprotected txs are disabled
	if !b.UnprotectedAllowed() && !tx.Protected() {
		// Ensure only eip155 signed transactions are submitted if EIP155Required is set.
		return common.Hash{}, errors.New("only replay-protected (EIP-155) transactions allowed over RPC")
	}

	ethereumTx := &evmtypes.MsgEthereumTx{}
	if err := ethereumTx.FromEthereumTx(tx); err != nil {
		b.logger.Error("transaction converting failed", "error", err.Error())
		return common.Hash{}, err
	}

	if err := ethereumTx.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return common.Hash{}, err
	}

	// Query params to use the EVM denomination
	res, err := b.queryClient.QueryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		b.logger.Error("failed to query evm params", "error", err.Error())
		return common.Hash{}, err
	}

	cosmosTx, err := ethereumTx.BuildTx(b.clientCtx.TxConfig.NewTxBuilder(), res.Params.EvmDenom)
	if err != nil {
		b.logger.Error("failed to build cosmos tx", "error", err.Error())
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txBytes, err := b.clientCtx.TxConfig.TxEncoder()(cosmosTx)
	if err != nil {
		b.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return common.Hash{}, err
	}

	txHash := ethereumTx.AsTransaction().Hash()

	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = sdkerrors.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}
	if err != nil {
		b.logger.Error("failed to broadcast tx", "error", err.Error())
		return txHash, err
	}

	return txHash, nil
}

// Accounts returns the list of accounts available to this node.
func (b *Backend) Accounts() ([]common.Address, error) {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := b.clientCtx.Keyring.List()
	if err != nil {
		return addresses, err
	}

	for _, info := range infos {
		pubKey, err := info.GetPubKey()
		if err != nil {
			return nil, err
		}
		addressBytes := pubKey.Address().Bytes()
		addresses = append(addresses, common.BytesToAddress(addressBytes))
	}

	return addresses, nil
}

// GetBalance returns the provided account's balance up to the provided block number.
func (b *Backend) GetBalance(address common.Address, blockNrOrHash types.BlockNumberOrHash) (*hexutil.Big, error) {
	blockNum, err := b.GetBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Balance(types.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return (*hexutil.Big)(val.BigInt()), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (b *Backend) GetStorageAt(address common.Address, key string, blockNrOrHash types.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.GetBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := b.queryClient.Storage(types.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetCode returns the contract code at the given address and block number.
func (b *Backend) GetCode(address common.Address, blockNrOrHash types.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.GetBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Code(types.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetProof returns an account object with proof and any storage proofs
func (b *Backend) GetProof(address common.Address, storageKeys []string, blockNrOrHash types.BlockNumberOrHash) (*types.AccountResult, error) {
	// TODO refractor this function into smaller blocks
	blockNum, err := b.GetBlockNumber(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	height := blockNum.Int64()
	ctx := types.ContextWithHeight(height)

	// if the height is equal to zero, meaning the query condition of the block is either "pending" or "latest"
	if height == 0 {
		bn, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}

		if bn > math.MaxInt64 {
			return nil, fmt.Errorf("not able to query block number greater than MaxInt64")
		}

		height = int64(bn)
	}

	clientCtx := b.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]types.StorageResult, len(storageKeys))

	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := b.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
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

	res, err := b.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := authtypes.AddressStoreKey(sdk.AccAddress(address.Bytes()))
	_, proof, err := b.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	// check for proof
	var accProofStr string
	if proof != nil {
		accProofStr = proof.String()
	}

	balance, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return &types.AccountResult{
		Address:      address,
		AccountProof: []string{accProofStr},
		Balance:      (*hexutil.Big)(balance.BigInt()),
		CodeHash:     common.HexToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Ethermint doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronize from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (b *Backend) Syncing() (interface{}, error) {
	status, err := b.clientCtx.Client.Status(b.ctx)
	if err != nil {
		return false, err
	}

	if !status.SyncInfo.CatchingUp {
		return false, nil
	}

	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(status.SyncInfo.EarliestBlockHeight),
		"currentBlock":  hexutil.Uint64(status.SyncInfo.LatestBlockHeight),
		// "highestBlock":  nil, // NA
		// "pulledStates":  nil, // NA
		// "knownStates":   nil, // NA
	}, nil
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (b *Backend) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	from := sdk.AccAddress(address.Bytes())

	_, err := b.clientCtx.Keyring.KeyByAddress(from)
	if err != nil {
		b.logger.Error("failed to find key in keyring", "address", address.String())
		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	// Sign the requested hash with the wallet
	signature, _, err := b.clientCtx.Keyring.SignByAddress(from, data)
	if err != nil {
		b.logger.Error("keyring.SignByAddress failed", "address", address.Hex())
		return nil, err
	}

	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SignTypedData signs EIP-712 conformant typed data
func (b *Backend) SignTypedData(address common.Address, typedData apitypes.TypedData) (hexutil.Bytes, error) {
	from := sdk.AccAddress(address.Bytes())

	_, err := b.clientCtx.Keyring.KeyByAddress(from)
	if err != nil {
		b.logger.Error("failed to find key in keyring", "address", address.String())
		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	sigHash, err := eip712.ComputeTypedDataHash(typedData)
	if err != nil {
		return nil, err
	}

	// Sign the requested hash with the wallet
	signature, _, err := b.clientCtx.Keyring.SignByAddress(from, sigHash)
	if err != nil {
		b.logger.Error("keyring.SignByAddress failed", "address", address.Hex())
		return nil, err
	}

	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// ChainId is the EIP-155 replay-protection chain id for the current ethereum chain config.
func (b *Backend) ChainId() (*hexutil.Big, error) {
	eip155ChainID, err := ethermint.ParseChainID(b.clientCtx.ChainID)
	if err != nil {
		panic(err)
	}
	// if current block is at or past the EIP-155 replay-protection fork block, return chainID from config
	bn, err := b.BlockNumber()
	if err != nil {
		b.logger.Debug("failed to fetch latest block number", "error", err.Error())
		return (*hexutil.Big)(eip155ChainID), nil
	}

	if config := b.ChainConfig(); config.IsEIP155(new(big.Int).SetUint64(uint64(bn))) {
		return (*hexutil.Big)(config.ChainID), nil
	}

	return nil, fmt.Errorf("chain not synced beyond EIP-155 replay-protection fork block")
}

// SetEtherbase sets the etherbase of the miner
func (b *Backend) SetEtherbase(etherbase common.Address) bool {
	delAddr, err := b.GetCoinbase()
	if err != nil {
		b.logger.Debug("failed to get coinbase address", "error", err.Error())
		return false
	}

	withdrawAddr := sdk.AccAddress(etherbase.Bytes())
	msg := distributiontypes.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

	if err := msg.ValidateBasic(); err != nil {
		b.logger.Debug("tx failed basic validation", "error", err.Error())
		return false
	}

	// Assemble transaction from fields
	builder, ok := b.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		b.logger.Debug("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
		return false
	}

	err = builder.SetMsgs(msg)
	if err != nil {
		b.logger.Error("builder.SetMsgs failed", "error", err.Error())
		return false
	}

	// Fetch minimun gas price to calculate fees using the configuration.
	minGasPrices := b.cfg.GetMinGasPrices()
	if len(minGasPrices) == 0 || minGasPrices.Empty() {
		b.logger.Debug("the minimun fee is not set")
		return false
	}
	minGasPriceValue := minGasPrices[0].Amount
	denom := minGasPrices[0].Denom

	delCommonAddr := common.BytesToAddress(delAddr.Bytes())
	nonce, err := b.GetTransactionCount(delCommonAddr, rpctypes.EthPendingBlockNumber)
	if err != nil {
		b.logger.Debug("failed to get nonce", "error", err.Error())
		return false
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(b.clientCtx.ChainID).
		WithKeybase(b.clientCtx.Keyring).
		WithTxConfig(b.clientCtx.TxConfig).
		WithSequence(uint64(*nonce)).
		WithGasAdjustment(1.25)

	_, gas, err := tx.CalculateGas(b.clientCtx, txFactory, msg)
	if err != nil {
		b.logger.Debug("failed to calculate gas", "error", err.Error())
		return false
	}

	txFactory = txFactory.WithGas(gas)

	value := new(big.Int).SetUint64(gas * minGasPriceValue.Ceil().TruncateInt().Uint64())
	fees := sdk.Coins{sdk.NewCoin(denom, sdkmath.NewIntFromBigInt(value))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(gas)

	keyInfo, err := b.clientCtx.Keyring.KeyByAddress(delAddr)
	if err != nil {
		b.logger.Debug("failed to get the wallet address using the keyring", "error", err.Error())
		return false
	}

	if err := tx.Sign(txFactory, keyInfo.Name, builder, false); err != nil {
		b.logger.Debug("failed to sign tx", "error", err.Error())
		return false
	}

	// Encode transaction by default Tx encoder
	txEncoder := b.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		b.logger.Debug("failed to encode eth tx using default encoder", "error", err.Error())
		return false
	}

	tmHash := common.BytesToHash(tmtypes.Tx(txBytes).Hash())

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := b.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if rsp != nil && rsp.Code != 0 {
		err = sdkerrors.ABCIError(rsp.Codespace, rsp.Code, rsp.RawLog)
	}
	if err != nil {
		b.logger.Debug("failed to broadcast tx", "error", err.Error())
		return false
	}

	b.logger.Debug("broadcasted tx to set miner withdraw address (etherbase)", "hash", tmHash.String())
	return true
}

// ImportRawKey armors and encrypts a given raw hex encoded ECDSA key and stores it into the key directory.
// The name of the key will have the format "personal_<length-keys>", where <length-keys> is the total number of
// keys stored on the keyring.
//
// NOTE: The key will be both armored and encrypted using the same passphrase.
func (b *Backend) ImportRawKey(privkey, password string) (common.Address, error) {
	priv, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}

	privKey := &ethsecp256k1.PrivKey{Key: crypto.FromECDSA(priv)}

	addr := sdk.AccAddress(privKey.PubKey().Address().Bytes())
	ethereumAddr := common.BytesToAddress(addr)

	// return if the key has already been imported
	if _, err := b.clientCtx.Keyring.KeyByAddress(addr); err == nil {
		return ethereumAddr, nil
	}

	// ignore error as we only care about the length of the list
	list, _ := b.clientCtx.Keyring.List()
	privKeyName := fmt.Sprintf("personal_%d", len(list))

	armor := sdkcrypto.EncryptArmorPrivKey(privKey, password, ethsecp256k1.KeyType)

	if err := b.clientCtx.Keyring.ImportPrivKey(privKeyName, armor, password); err != nil {
		return common.Address{}, err
	}

	b.logger.Info("key successfully imported", "name", privKeyName, "address", ethereumAddr.String())

	return ethereumAddr, nil
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (b *Backend) ListAccounts() ([]common.Address, error) {
	addrs := []common.Address{}

	list, err := b.clientCtx.Keyring.List()
	if err != nil {
		return nil, err
	}

	for _, info := range list {
		pubKey, err := info.GetPubKey()
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, common.BytesToAddress(pubKey.Address()))
	}

	return addrs, nil
}

// NewAccount will create a new account and returns the address for the new account.
func (b *Backend) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (*keyring.Record, error) {
	info, _, err := b.clientCtx.Keyring.NewMnemonic(uid, keyring.English, bip39Passphrase, bip39Passphrase, algo)
	if err != nil {
		return nil, err
	}
	return info, err
}

package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"

	"github.com/tharsis/ethermint/rpc/types"
	ethermint "github.com/tharsis/ethermint/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"
)

var bAttributeKeyEthereumBloom = []byte(evmtypes.AttributeKeyEthereumBloom)

// BlockNumber returns the current block number in abci app state.
// Because abci app state could lag behind from tendermint latest block, it's more stable
// for the client to use the latest block number in abci app state than tendermint rpc.
func (b *Backend) BlockNumber() (hexutil.Uint64, error) {
	// do any grpc query, ignore the response and use the returned block height
	var header metadata.MD
	_, err := b.queryClient.Params(b.ctx, &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
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
func (b *Backend) GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := b.GetTendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	// return if requested block height is greater than the current one
	if resBlock == nil || resBlock.Block == nil {
		return nil, nil
	}

	res, err := b.EthBlockFromTendermint(resBlock.Block, fullTx)
	if err != nil {
		b.logger.Debug("EthBlockFromTendermint failed", "height", blockNum, "error", err.Error())
		return nil, err
	}

	return res, nil
}

// GetBlockByHash returns the block identified by hash.
func (b *Backend) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		b.logger.Debug("BlockByHash block not found", "hash", hash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		b.logger.Debug("BlockByHash block not found", "hash", hash.Hex())
		return nil, nil
	}

	return b.EthBlockFromTendermint(resBlock.Block, fullTx)
}

// BlockByNumber returns the block identified by number.
func (b *Backend) BlockByNumber(blockNum types.BlockNumber) (*ethtypes.Block, error) {
	height := blockNum.Int64()

	switch blockNum {
	case types.EthLatestBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
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

	resBlock, err := b.clientCtx.Client.Block(b.ctx, &height)
	if err != nil {
		b.logger.Debug("HeaderByNumber failed", "height", height)
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		return nil, errors.Errorf("block not found for height %d", height)
	}

	return b.EthBlockFromTm(resBlock.Block)
}

// BlockByHash returns the block identified by hash.
func (b *Backend) BlockByHash(hash common.Hash) (*ethtypes.Block, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		b.logger.Debug("HeaderByHash failed", "hash", hash.Hex())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		return nil, errors.Errorf("block not found for hash %s", hash)
	}

	return b.EthBlockFromTm(resBlock.Block)
}

func (b *Backend) EthBlockFromTm(block *tmtypes.Block) (*ethtypes.Block, error) {
	height := block.Height
	bloom, err := b.BlockBloom(&height)
	if err != nil {
		b.logger.Debug("HeaderByNumber BlockBloom failed", "height", height)
	}

	baseFee, err := b.BaseFee(height)
	if err != nil {
		b.logger.Debug("HeaderByNumber BaseFee failed", "height", height, "error", err.Error())
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(block.Header, bloom, baseFee)

	var txs []*ethtypes.Transaction
	for _, txBz := range block.Txs {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			b.logger.Debug("failed to decode transaction in block", "height", height, "error", err.Error())
			continue
		}

		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			tx := ethMsg.AsTransaction()
			txs = append(txs, tx)
		}
	}

	// TODO: add tx receipts
	ethBlock := ethtypes.NewBlock(ethHeader, txs, nil, nil, nil)
	return ethBlock, nil
}

// GetTendermintBlockByNumber returns a Tendermint format block by block number
func (b *Backend) GetTendermintBlockByNumber(blockNum types.BlockNumber) (*tmrpctypes.ResultBlock, error) {
	height := blockNum.Int64()
	currentBlockNumber, _ := b.BlockNumber()

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

	resBlock, err := b.clientCtx.Client.Block(b.ctx, &height)
	if err != nil {
		if resBlock, err = b.clientCtx.Client.Block(b.ctx, nil); err != nil {
			b.logger.Debug("tendermint client failed to get latest block", "height", height, "error", err.Error())
			return nil, nil
		}
	}

	if resBlock.Block == nil {
		b.logger.Debug("GetBlockByNumber block not found", "height", height)
		return nil, nil
	}

	return resBlock, nil
}

// GetTendermintBlockByHash returns a Tendermint format block by block number
func (b *Backend) GetTendermintBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, blockHash.Bytes())
	if err != nil {
		b.logger.Debug("tendermint client failed to get block", "blockHash", blockHash.Hex(), "error", err.Error())
	}

	if resBlock == nil || resBlock.Block == nil {
		b.logger.Debug("GetBlockByNumber block not found", "blockHash", blockHash.Hex())
		return nil, nil
	}

	return resBlock, nil
}

// BlockBloom query block bloom filter from block results
func (b *Backend) BlockBloom(height *int64) (ethtypes.Bloom, error) {
	result, err := b.clientCtx.Client.BlockResults(b.ctx, height)
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
func (b *Backend) EthBlockFromTendermint(
	block *tmtypes.Block,
	fullTx bool,
) (map[string]interface{}, error) {
	ethRPCTxs := []interface{}{}

	ctx := types.ContextWithHeight(block.Height)

	baseFee, err := b.BaseFee(block.Height)
	if err != nil {
		return nil, err
	}

	resBlockResult, err := b.clientCtx.Client.BlockResults(ctx, &block.Height)
	if err != nil {
		return nil, err
	}

	txResults := resBlockResult.TxsResults
	txIndex := uint64(0)

	for i, txBz := range block.Txs {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			b.logger.Debug("failed to decode transaction in block", "height", block.Height, "error", err.Error())
			continue
		}

		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			tx := ethMsg.AsTransaction()

			// check tx exists on EVM by cross checking with blockResults
			if txResults[i].Code != 0 {
				b.logger.Debug("invalid tx result code", "hash", tx.Hash().Hex())
				continue
			}

			if !fullTx {
				hash := tx.Hash()
				ethRPCTxs = append(ethRPCTxs, hash)
				continue
			}

			rpcTx, err := types.NewRPCTransaction(
				tx,
				common.BytesToHash(block.Hash()),
				uint64(block.Height),
				txIndex,
				baseFee,
			)
			if err != nil {
				b.logger.Debug("NewTransactionFromData for receipt failed", "hash", tx.Hash().Hex(), "error", err.Error())
				continue
			}
			ethRPCTxs = append(ethRPCTxs, rpcTx)
			txIndex++
		}
	}

	bloom, err := b.BlockBloom(&block.Height)
	if err != nil {
		b.logger.Debug("failed to query BlockBloom", "height", block.Height, "error", err.Error())
	}

	req := &evmtypes.QueryValidatorAccountRequest{
		ConsAddress: sdk.ConsAddress(block.Header.ProposerAddress).String(),
	}

	res, err := b.queryClient.ValidatorAccount(ctx, req)
	if err != nil {
		b.logger.Debug(
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

	gasLimit, err := types.BlockMaxGasFromConsensusParams(ctx, b.clientCtx, block.Height)
	if err != nil {
		b.logger.Error("failed to query consensus params", "error", err.Error())
	}

	gasUsed := uint64(0)

	for _, txsResult := range txResults {
		// workaround for cosmos-sdk bug. https://github.com/cosmos/cosmos-sdk/issues/10832
		if txsResult.GetCode() == 11 && txsResult.GetLog() == "no block gas left to run tx: out of gas" {
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
	height := blockNum.Int64()

	switch blockNum {
	case types.EthLatestBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
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

	resBlock, err := b.clientCtx.Client.Block(b.ctx, &height)
	if err != nil {
		b.logger.Debug("HeaderByNumber failed")
		return nil, err
	}

	bloom, err := b.BlockBloom(&resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("HeaderByNumber BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := b.BaseFee(resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("HeaderByNumber BaseFee failed", "height", resBlock.Block.Height, "error", err.Error())
		return nil, err
	}

	ethHeader := types.EthHeaderFromTendermint(resBlock.Block.Header, bloom, baseFee)
	return ethHeader, nil
}

// HeaderByHash returns the block header identified by hash.
func (b *Backend) HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error) {
	resBlock, err := b.clientCtx.Client.BlockByHash(b.ctx, blockHash.Bytes())
	if err != nil {
		b.logger.Debug("HeaderByHash failed", "hash", blockHash.Hex())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		return nil, errors.Errorf("block not found for hash %s", blockHash.Hex())
	}

	bloom, err := b.BlockBloom(&resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("HeaderByHash BlockBloom failed", "height", resBlock.Block.Height)
	}

	baseFee, err := b.BaseFee(resBlock.Block.Height)
	if err != nil {
		b.logger.Debug("HeaderByHash BaseFee failed", "height", resBlock.Block.Height, "error", err.Error())
		return nil, err
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
	blockRes, err := b.clientCtx.Client.BlockResults(b.ctx, height)
	if err != nil {
		return nil, err
	}

	blockLogs := [][]*ethtypes.Log{}
	for _, txResult := range blockRes.TxsResults {
		logs, err := AllTxLogsFromEvents(txResult.Events)
		if err != nil {
			return nil, err
		}

		blockLogs = append(blockLogs, logs...)
	}

	return blockLogs, nil
}

// GetLogs returns all the logs from all the ethereum transactions in a block.
func (b *Backend) GetLogs(hash common.Hash) ([][]*ethtypes.Log, error) {
	block, err := b.clientCtx.Client.BlockByHash(b.ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}
	return b.GetLogsByHeight(&block.Block.Header.Height)
}

func (b *Backend) GetLogsByNumber(blockNum types.BlockNumber) ([][]*ethtypes.Log, error) {
	height := blockNum.Int64()

	switch blockNum {
	case types.EthLatestBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
		if currentBlockNumber > 0 {
			height = int64(currentBlockNumber)
		}
	case types.EthPendingBlockNumber:
		currentBlockNumber, _ := b.BlockNumber()
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

	return b.GetLogsByHeight(&height)
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

// GetTransactionByHash returns the Ethereum format transaction identified by Ethereum transaction hash
func (b *Backend) GetTransactionByHash(txHash common.Hash) (*types.RPCTransaction, error) {
	res, err := b.GetTxByEthHash(txHash)
	hexTx := txHash.Hex()
	if err != nil {
		// try to find tx in mempool
		txs, err := b.PendingTransactions()
		if err != nil {
			b.logger.Debug("tx not found", "hash", hexTx, "error", err.Error())
			return nil, nil
		}

		for _, tx := range txs {
			msg, err := evmtypes.UnwrapEthereumMsg(tx, txHash)
			if err != nil {
				// not ethereum tx
				continue
			}

			if msg.Hash == hexTx {
				rpctx, err := types.NewTransactionFromMsg(
					msg,
					common.Hash{},
					uint64(0),
					uint64(0),
					nil,
				)
				if err != nil {
					return nil, err
				}
				return rpctx, nil
			}
		}

		b.logger.Debug("tx not found", "hash", hexTx)
		return nil, nil
	}

	if res.TxResult.Code != 0 {
		return nil, errors.New("invalid ethereum tx")
	}

	msgIndex, attrs := types.FindTxAttributes(res.TxResult.Events, hexTx)
	if msgIndex < 0 {
		return nil, fmt.Errorf("ethereum tx not found in msgs: %s", hexTx)
	}

	tx, err := b.clientCtx.TxConfig.TxDecoder()(res.Tx)
	if err != nil {
		return nil, err
	}

	// the `msgIndex` is inferred from tx events, should be within the bound.
	msg, ok := tx.GetMsgs()[msgIndex].(*evmtypes.MsgEthereumTx)
	if !ok {
		return nil, errors.New("invalid ethereum tx")
	}

	block, err := b.clientCtx.Client.Block(b.ctx, &res.Height)
	if err != nil {
		b.logger.Debug("block not found", "height", res.Height, "error", err.Error())
		return nil, err
	}

	// Try to find txIndex from events
	found := false
	txIndex, err := types.GetUint64Attribute(attrs, evmtypes.AttributeKeyTxIndex)
	if err == nil {
		found = true
	} else {
		// Fallback to find tx index by iterating all valid eth transactions
		blockRes, err := b.clientCtx.Client.BlockResults(b.ctx, &block.Block.Height)
		if err != nil {
			return nil, nil
		}
		msgs := b.GetEthereumMsgsFromTendermintBlock(block, blockRes)
		for i := range msgs {
			if msgs[i].Hash == hexTx {
				txIndex = uint64(i)
				found = true
				break
			}
		}
	}
	if !found {
		return nil, errors.New("can't find index of ethereum tx")
	}

	baseFee, err := b.BaseFee(block.Block.Height)
	if err != nil {
		return nil, err
	}

	return types.NewTransactionFromMsg(
		msg,
		common.BytesToHash(block.BlockID.Hash.Bytes()),
		uint64(res.Height),
		txIndex,
		baseFee,
	)
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

	txHash := msg.AsTransaction().Hash()

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

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (b *Backend) GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error) {
	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := b.clientCtx.AccountRetriever

	err := accRet.EnsureExists(b.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		n := hexutil.Uint64(0)
		return &n, nil
	}

	includePending := blockNum == types.EthPendingBlockNumber
	nonce, err := b.getAccountNonce(address, includePending, blockNum.Int64(), b.logger)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
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
func (b *Backend) BaseFee(height int64) (*big.Int, error) {
	// return BaseFee if London hard fork is activated and feemarket is enabled
	res, err := b.queryClient.BaseFee(types.ContextWithHeight(height), &evmtypes.QueryBaseFeeRequest{})
	if err != nil {
		return nil, err
	}

	if res.BaseFee == nil {
		return nil, nil
	}

	return res.BaseFee.BigInt(), nil
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
		tendermintBlockResult, err := b.clientCtx.Client.BlockResults(b.ctx, &tendermintblock.Block.Height)
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

// GetEthereumMsgsFromTendermintBlock returns all real MsgEthereumTxs from a Tendermint block.
// It also ensures consistency over the correct txs indexes across RPC endpoints
func (b *Backend) GetEthereumMsgsFromTendermintBlock(block *tmrpctypes.ResultBlock, blockRes *tmrpctypes.ResultBlockResults) []*evmtypes.MsgEthereumTx {
	var result []*evmtypes.MsgEthereumTx

	txResults := blockRes.TxsResults

	for i, tx := range block.Block.Txs {
		// check tx exists on EVM by cross checking with blockResults
		if txResults[i].Code != 0 {
			b.logger.Debug("invalid tx result code", "cosmos-hash", hexutil.Encode(tx.Hash()))
			continue
		}

		tx, err := b.clientCtx.TxConfig.TxDecoder()(tx)
		if err != nil {
			b.logger.Debug("failed to decode transaction in block", "height", block.Block.Height, "error", err.Error())
			continue
		}

		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			result = append(result, ethMsg)
		}
	}

	return result
}

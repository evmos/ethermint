package types

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"
	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// RawTxToEthTx returns a evm MsgEthereum transaction from raw tx bytes.
func RawTxToEthTx(clientCtx client.Context, txBz tmtypes.Tx) ([]*evmtypes.MsgEthereumTx, error) {
	tx, err := clientCtx.TxConfig.TxDecoder()(txBz)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	ethTxs := make([]*evmtypes.MsgEthereumTx, len(tx.GetMsgs()))
	for i, msg := range tx.GetMsgs() {
		ethTx, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return nil, fmt.Errorf("invalid message type %T, expected %T", msg, &evmtypes.MsgEthereumTx{})
		}
		ethTxs[i] = ethTx
	}
	return ethTxs, nil
}

// EthHeaderFromTendermint is an util function that returns an Ethereum Header
// from a tendermint Header.
func EthHeaderFromTendermint(header tmtypes.Header, bloom ethtypes.Bloom, baseFee *big.Int) *ethtypes.Header {
	txHash := ethtypes.EmptyRootHash
	if len(header.DataHash) == 0 {
		txHash = common.BytesToHash(header.DataHash)
	}

	return &ethtypes.Header{
		ParentHash:  common.BytesToHash(header.LastBlockID.Hash.Bytes()),
		UncleHash:   ethtypes.EmptyUncleHash,
		Coinbase:    common.BytesToAddress(header.ProposerAddress),
		Root:        common.BytesToHash(header.AppHash),
		TxHash:      txHash,
		ReceiptHash: ethtypes.EmptyRootHash,
		Bloom:       bloom,
		Difficulty:  big.NewInt(0),
		Number:      big.NewInt(header.Height),
		GasLimit:    0,
		GasUsed:     0,
		Time:        uint64(header.Time.UTC().Unix()),
		Extra:       []byte{},
		MixDigest:   common.Hash{},
		Nonce:       ethtypes.BlockNonce{},
		BaseFee:     baseFee,
	}
}

// BlockMaxGasFromConsensusParams returns the gas limit for the current block from the chain consensus params.
func BlockMaxGasFromConsensusParams(goCtx context.Context, clientCtx client.Context, blockHeight int64) (int64, error) {
	resConsParams, err := clientCtx.Client.ConsensusParams(goCtx, &blockHeight)
	if err != nil {
		return int64(^uint32(0)), err
	}

	gasLimit := resConsParams.ConsensusParams.Block.MaxGas
	if gasLimit == -1 {
		// Sets gas limit to max uint32 to not error with javascript dev tooling
		// This -1 value indicating no block gas limit is set to max uint64 with geth hexutils
		// which errors certain javascript dev tooling which only supports up to 53 bits
		gasLimit = int64(^uint32(0))
	}

	return gasLimit, nil
}

// FormatBlock creates an ethereum block from a tendermint header and ethereum-formatted
// transactions.
func FormatBlock(
	header tmtypes.Header, size int, gasLimit int64,
	gasUsed *big.Int, transactions []interface{}, bloom ethtypes.Bloom,
	validatorAddr common.Address, baseFee *big.Int,
) map[string]interface{} {
	var transactionsRoot common.Hash
	if len(transactions) == 0 {
		transactionsRoot = ethtypes.EmptyRootHash
	} else {
		transactionsRoot = common.BytesToHash(header.DataHash)
	}

	result := map[string]interface{}{
		"number":           hexutil.Uint64(header.Height),
		"hash":             hexutil.Bytes(header.Hash()),
		"parentHash":       common.BytesToHash(header.LastBlockID.Hash.Bytes()),
		"nonce":            ethtypes.BlockNonce{},   // PoW specific
		"sha3Uncles":       ethtypes.EmptyUncleHash, // No uncles in Tendermint
		"logsBloom":        bloom,
		"stateRoot":        hexutil.Bytes(header.AppHash),
		"miner":            validatorAddr,
		"mixHash":          common.Hash{},
		"difficulty":       (*hexutil.Big)(big.NewInt(0)),
		"extraData":        "0x",
		"size":             hexutil.Uint64(size),
		"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
		"gasUsed":          (*hexutil.Big)(gasUsed),
		"timestamp":        hexutil.Uint64(header.Time.Unix()),
		"transactionsRoot": transactionsRoot,
		"receiptsRoot":     ethtypes.EmptyRootHash,

		"uncles":          []common.Hash{},
		"transactions":    transactions,
		"totalDifficulty": (*hexutil.Big)(big.NewInt(0)),
	}

	if baseFee != nil {
		result["baseFeePerGas"] = (*hexutil.Big)(baseFee)
	}

	return result
}

type DataError interface {
	Error() string          // returns the message
	ErrorData() interface{} // returns the error data
}

type dataError struct {
	msg  string
	data string
}

func (d *dataError) Error() string {
	return d.msg
}

func (d *dataError) ErrorData() interface{} {
	return d.data
}

type SDKTxLogs struct {
	Log string `json:"log"`
}

const LogRevertedFlag = "transaction reverted"

func ErrRevertedWith(data []byte) DataError {
	return &dataError{
		msg:  "VM execution error.",
		data: fmt.Sprintf("0x%s", hex.EncodeToString(data)),
	}
}

// NewTransactionFromMsg returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func NewTransactionFromMsg(
	msg *evmtypes.MsgEthereumTx,
	blockHash common.Hash,
	blockNumber, index uint64,
	baseFee *big.Int,
) (*RPCTransaction, error) {
	tx := msg.AsTransaction()
	return NewRPCTransaction(tx, blockHash, blockNumber, index, baseFee)
}

// NewTransactionFromData returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func NewRPCTransaction(
	tx *ethtypes.Transaction, blockHash common.Hash, blockNumber, index uint64, baseFee *big.Int,
) (*RPCTransaction, error) {
	// Determine the signer. For replay-protected transactions, use the most permissive
	// signer, because we assume that signers are backwards-compatible with old
	// transactions. For non-protected transactions, the homestead signer signer is used
	// because the return value of ChainId is zero for those transactions.
	var signer ethtypes.Signer
	if tx.Protected() {
		signer = ethtypes.LatestSignerForChainID(tx.ChainId())
	} else {
		signer = ethtypes.HomesteadSigner{}
	}
	from, _ := ethtypes.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()
	result := &RPCTransaction{
		Type:     hexutil.Uint64(tx.Type()),
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = &blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = (*hexutil.Uint64)(&index)
	}
	switch tx.Type() {
	case ethtypes.AccessListTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
	case ethtypes.DynamicFeeTxType:
		al := tx.AccessList()
		result.Accesses = &al
		result.ChainID = (*hexutil.Big)(tx.ChainId())
		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
		// if the transaction has been mined, compute the effective gas price
		if baseFee != nil && blockHash != (common.Hash{}) {
			// price = min(tip, gasFeeCap - baseFee) + baseFee
			price := math.BigMin(new(big.Int).Add(tx.GasTipCap(), baseFee), tx.GasFeeCap())
			result.GasPrice = (*hexutil.Big)(price)
		} else {
			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
		}
	}
	return result, nil
}

// BaseFeeFromEvents parses the feemarket basefee from cosmos events
func BaseFeeFromEvents(events []abci.Event) *big.Int {
	for _, event := range events {
		if event.Type != feemarkettypes.EventTypeFeeMarket {
			continue
		}

		for _, attr := range event.Attributes {
			if bytes.Equal(attr.Key, []byte(feemarkettypes.AttributeKeyBaseFee)) {
				result, success := new(big.Int).SetString(string(attr.Value), 10)
				if success {
					return result
				}

				return nil
			}
		}
	}
	return nil
}

// FindTxAttributes returns the msg index of the eth tx in cosmos tx, and the attributes,
// returns -1 and nil if not found.
func FindTxAttributes(events []abci.Event, txHash string) (int, map[string]string) {
	msgIndex := -1
	for _, event := range events {
		if event.Type != evmtypes.EventTypeEthereumTx {
			continue
		}

		msgIndex++

		value := FindAttribute(event.Attributes, []byte(evmtypes.AttributeKeyEthereumTxHash))
		if !bytes.Equal(value, []byte(txHash)) {
			continue
		}

		// found, convert attributes to map for later lookup
		attrs := make(map[string]string, len(event.Attributes))
		for _, attr := range event.Attributes {
			attrs[string(attr.Key)] = string(attr.Value)
		}
		return msgIndex, attrs
	}
	// not found
	return -1, nil
}

// FindAttribute find event attribute with specified key, if not found returns nil.
func FindAttribute(attrs []abci.EventAttribute, key []byte) []byte {
	for _, attr := range attrs {
		if !bytes.Equal(attr.Key, key) {
			continue
		}
		return attr.Value
	}
	return nil
}

// GetUint64Attribute parses the uint64 value from event attributes
func GetUint64Attribute(attrs map[string]string, key string) (uint64, error) {
	value, found := attrs[key]
	if !found {
		return 0, fmt.Errorf("tx index attribute not found: %s", key)
	}
	var result int64
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if result < 0 {
		return 0, fmt.Errorf("negative tx index: %d", result)
	}
	return uint64(result), nil
}

// AccumulativeGasUsedOfMsg accumulate the gas used by msgs before `msgIndex`.
func AccumulativeGasUsedOfMsg(events []abci.Event, msgIndex int) (gasUsed uint64) {
	for _, event := range events {
		if event.Type != evmtypes.EventTypeEthereumTx {
			continue
		}

		if msgIndex < 0 {
			break
		}
		msgIndex--

		value := FindAttribute(event.Attributes, []byte(evmtypes.AttributeKeyTxGasUsed))
		var result int64
		result, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			continue
		}
		gasUsed += uint64(result)
	}
	return
}

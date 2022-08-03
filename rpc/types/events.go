package types

import (
	"encoding/json"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// EventFormat is the format version of the events.
//
// To fix the issue of tx exceeds block gas limit, we changed the event format in a breaking way.
// But to avoid forcing clients to re-sync from scatch, we make json-rpc logic to be compatible with both formats.
type EventFormat int

const (
	eventFormatUnknown EventFormat = iota

	// Event Format 1 (the format used before PR #1062):
	// ```
	// ethereum_tx(amount, ethereumTxHash, [txIndex, txGasUsed], txHash, [receipient], ethereumTxFailed)
	// tx_log(txLog, txLog, ...)
	// ethereum_tx(amount, ethereumTxHash, [txIndex, txGasUsed], txHash, [receipient], ethereumTxFailed)
	// tx_log(txLog, txLog, ...)
	// ...
	// ```
	eventFormat1

	// Event Format 2 (the format used after PR #1062):
	// ```
	// ethereum_tx(ethereumTxHash, txIndex)
	// ethereum_tx(ethereumTxHash, txIndex)
	// ...
	// ethereum_tx(amount, ethereumTxHash, txIndex, txGasUsed, txHash, [receipient], ethereumTxFailed)
	// tx_log(txLog, txLog, ...)
	// ethereum_tx(amount, ethereumTxHash, txIndex, txGasUsed, txHash, [receipient], ethereumTxFailed)
	// tx_log(txLog, txLog, ...)
	// ...
	// ```
	// If the transaction exceeds block gas limit, it only emits the first part.
	eventFormat2
)

// ParsedTx is the tx infos parsed from events.
type ParsedTx struct {
	MsgIndex int

	// the following fields are parsed from events

	Hash common.Hash
	// -1 means uninitialized
	EthTxIndex int64
	GasUsed    uint64
	Failed     bool
	// unparsed tx log json strings
	RawLogs [][]byte
}

// NewParsedTx initialize a ParsedTx
func NewParsedTx(msgIndex int) ParsedTx {
	return ParsedTx{MsgIndex: msgIndex, EthTxIndex: -1}
}

// ParseTxLogs decode the raw logs into ethereum format.
func (p ParsedTx) ParseTxLogs() ([]*ethtypes.Log, error) {
	logs := make([]*evmtypes.Log, 0, len(p.RawLogs))
	for _, raw := range p.RawLogs {
		var log evmtypes.Log
		if err := json.Unmarshal(raw, &log); err != nil {
			return nil, err
		}

		logs = append(logs, &log)
	}
	return evmtypes.LogsToEthereum(logs), nil
}

// ParsedTxs is the tx infos parsed from eth tx events.
type ParsedTxs struct {
	// one item per message
	Txs []ParsedTx
	// map tx hash to msg index
	TxHashes map[common.Hash]int
}

// ParseTxResult parse eth tx infos from cosmos-sdk events.
// It supports two event formats, the formats are described in the comments of the format constants.
func ParseTxResult(result *abci.ResponseDeliverTx) (*ParsedTxs, error) {
	format := eventFormatUnknown
	// the index of current ethereum_tx event in format 1 or the second part of format 2
	eventIndex := -1

	p := &ParsedTxs{
		TxHashes: make(map[common.Hash]int),
	}
	for _, event := range result.Events {
		switch event.Type {
		case evmtypes.EventTypeEthereumTx:
			if format == eventFormatUnknown {
				// discover the format version by inspect the first ethereum_tx event.
				if len(event.Attributes) > 2 {
					format = eventFormat1
				} else {
					format = eventFormat2
				}
			}

			if len(event.Attributes) == 2 {
				// the first part of format 2
				if err := p.newTx(event.Attributes); err != nil {
					return nil, err
				}
			} else {
				// format 1 or second part of format 2
				eventIndex++
				if format == eventFormat1 {
					// append tx
					if err := p.newTx(event.Attributes); err != nil {
						return nil, err
					}
				} else {
					// the second part of format 2, update tx fields
					if err := p.updateTx(eventIndex, event.Attributes); err != nil {
						return nil, err
					}
				}
			}
		case evmtypes.EventTypeTxLog:
			// reuse the eventIndex set by previous ethereum_tx event
			p.Txs[eventIndex].RawLogs = parseRawLogs(event.Attributes)
		}
	}

	// some old versions miss some events, fill it with tx result
	if len(p.Txs) == 1 {
		p.Txs[0].GasUsed = uint64(result.GasUsed)
	}

	return p, nil
}

// newTx parse a new tx from events, called during parsing.
func (p *ParsedTxs) newTx(attrs []abci.EventAttribute) error {
	msgIndex := len(p.Txs)
	tx := NewParsedTx(msgIndex)
	if err := fillTxAttributes(&tx, attrs); err != nil {
		return err
	}
	p.Txs = append(p.Txs, tx)
	p.TxHashes[tx.Hash] = msgIndex
	return nil
}

// updateTx updates an exiting tx from events, called during parsing.
func (p *ParsedTxs) updateTx(eventIndex int, attrs []abci.EventAttribute) error {
	return fillTxAttributes(&p.Txs[eventIndex], attrs)
}

// GetTxByHash find ParsedTx by tx hash, returns nil if not exists.
func (p *ParsedTxs) GetTxByHash(hash common.Hash) *ParsedTx {
	if idx, ok := p.TxHashes[hash]; ok {
		return &p.Txs[idx]
	}
	return nil
}

// GetTxByMsgIndex returns ParsedTx by msg index
func (p *ParsedTxs) GetTxByMsgIndex(i int) *ParsedTx {
	if i < 0 || i >= len(p.Txs) {
		return nil
	}
	return &p.Txs[i]
}

// GetTxByTxIndex returns ParsedTx by tx index
func (p *ParsedTxs) GetTxByTxIndex(txIndex int) *ParsedTx {
	if len(p.Txs) == 0 {
		return nil
	}
	// assuming the `EthTxIndex` increase continuously,
	// convert TxIndex to MsgIndex by subtract the begin TxIndex.
	msgIndex := txIndex - int(p.Txs[0].EthTxIndex)
	// GetTxByMsgIndex will check the bound
	return p.GetTxByMsgIndex(msgIndex)
}

// AccumulativeGasUsed calculates the accumulated gas used within the batch of txs
func (p *ParsedTxs) AccumulativeGasUsed(msgIndex int) (result uint64) {
	for i := 0; i <= msgIndex; i++ {
		result += p.Txs[i].GasUsed
	}
	return result
}

// fillTxAttribute parse attributes by name, less efficient than hardcode the index, but more stable against event
// format changes.
func fillTxAttribute(tx *ParsedTx, key []byte, value []byte) error {
	switch string(key) {
	case evmtypes.AttributeKeyEthereumTxHash:
		tx.Hash = common.HexToHash(string(value))
	case evmtypes.AttributeKeyTxIndex:
		txIndex, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		tx.EthTxIndex = txIndex
	case evmtypes.AttributeKeyTxGasUsed:
		gasUsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
		tx.GasUsed = uint64(gasUsed)
	case evmtypes.AttributeKeyEthereumTxFailed:
		tx.Failed = len(value) > 0
	}
	return nil
}

func fillTxAttributes(tx *ParsedTx, attrs []abci.EventAttribute) error {
	for _, attr := range attrs {
		if err := fillTxAttribute(tx, attr.Key, attr.Value); err != nil {
			return err
		}
	}
	return nil
}

func parseRawLogs(attrs []abci.EventAttribute) (logs [][]byte) {
	for _, attr := range attrs {
		logs = append(logs, attr.Value)
	}
	return logs
}

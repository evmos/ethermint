package types

import (
	"errors"
	"fmt"

	ethermint "github.com/cosmos/ethermint/types"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// NewTransactionLogs creates a new NewTransactionLogs instance.
func NewTransactionLogs(hash ethcmn.Hash, logs []*Log) TransactionLogs { // nolint: interfacer
	return TransactionLogs{
		Hash: hash.String(),
		Logs: logs,
	}
}

// NewTransactionLogsFromEth creates a new NewTransactionLogs instance using []*ethtypes.Log.
func NewTransactionLogsFromEth(hash ethcmn.Hash, ethlogs []*ethtypes.Log) TransactionLogs { // nolint: interfacer
	logs := make([]*Log, len(ethlogs))
	for i := range ethlogs {
		logs[i] = NewLogFromEth(ethlogs[i])
	}

	return TransactionLogs{
		Hash: hash.String(),
		Logs: logs,
	}
}

// Validate performs a basic validation of a GenesisAccount fields.
func (tx TransactionLogs) Validate() error {
	if ethermint.IsEmptyHash(tx.Hash) {
		return fmt.Errorf("hash cannot be the empty %s", tx.Hash)
	}

	for i, log := range tx.Logs {
		if log == nil {
			return fmt.Errorf("log %d cannot be nil", i)
		}
		if err := log.Validate(); err != nil {
			return fmt.Errorf("invalid log %d: %w", i, err)
		}
		if log.TxHash != tx.Hash {
			return fmt.Errorf("log tx hash mismatch (%s ≠ %s)", log.TxHash, tx.Hash)
		}
	}
	return nil
}

// EthLogs returns the Ethereum type Logs from the Transaction Logs.
func (tx TransactionLogs) EthLogs() []*ethtypes.Log {
	return LogsToEthereum(tx.Logs)
}

// Validate performs a basic validation of an ethereum Log fields.
func (log *Log) Validate() error {
	if ethermint.IsZeroAddress(log.Address) {
		return fmt.Errorf("log address cannot be empty %s", log.Address)
	}
	if IsEmptyHash(log.BlockHash) {
		return fmt.Errorf("block hash cannot be the empty %s", log.BlockHash)
	}
	if log.BlockNumber == 0 {
		return errors.New("block number cannot be zero")
	}
	if ethermint.IsEmptyHash(log.TxHash) {
		return fmt.Errorf("tx hash cannot be the empty %s", log.TxHash)
	}
	return nil
}

// ToEthereum returns the Ethereum type Log from a Ethermint-proto compatible Log.
func (log *Log) ToEthereum() *ethtypes.Log {
	topics := make([]ethcmn.Hash, len(log.Topics))
	for i := range log.Topics {
		topics[i] = ethcmn.HexToHash(log.Topics[i])
	}

	return &ethtypes.Log{
		Address:     ethcmn.HexToAddress(log.Address),
		Topics:      topics,
		Data:        log.Data,
		BlockNumber: log.BlockNumber,
		TxHash:      ethcmn.HexToHash(log.TxHash),
		TxIndex:     uint(log.TxIndex),
		BlockHash:   ethcmn.HexToHash(log.BlockHash),
		Removed:     log.Removed,
	}
}

// LogsToEthereum casts the Ethermint Logs to a slice of Ethereum Logs.
func LogsToEthereum(logs []*Log) []*ethtypes.Log {
	ethLogs := make([]*ethtypes.Log, len(logs))
	for i := range logs {
		ethLogs[i] = logs[i].ToEthereum()
	}
	return ethLogs
}

// NewLogFromEth creates a new Log instance from a Ethereum type Log.
func NewLogFromEth(log *ethtypes.Log) *Log {
	topics := make([]string, len(log.Topics))
	for i := range log.Topics {
		topics[i] = log.Topics[i].String()
	}

	return &Log{
		Address:     log.Address.String(),
		Topics:      topics,
		Data:        log.Data,
		BlockNumber: log.BlockNumber,
		TxHash:      log.TxHash.String(),
		TxIndex:     uint64(log.TxIndex),
		BlockHash:   log.BlockHash.String(),
		Removed:     log.Removed,
	}
}

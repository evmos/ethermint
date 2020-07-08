package types

import (
	"bytes"
	"errors"
	"fmt"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// TransactionLogs define the logs generated from a transaction execution
// with a given hash. It it used for import/export data as transactions are not persisted
// on blockchain state after an upgrade.
type TransactionLogs struct {
	Hash ethcmn.Hash     `json:"hash"`
	Logs []*ethtypes.Log `json:"logs"`
}

// NewTransactionLogs creates a new NewTransactionLogs instance.
func NewTransactionLogs(hash ethcmn.Hash, logs []*ethtypes.Log) TransactionLogs {
	return TransactionLogs{
		Hash: hash,
		Logs: logs,
	}
}

// MarshalLogs encodes an array of logs using amino
func MarshalLogs(logs []*ethtypes.Log) ([]byte, error) {
	return ModuleCdc.MarshalBinaryLengthPrefixed(logs)
}

// UnmarshalLogs decodes an amino-encoded byte array into an array of logs
func UnmarshalLogs(in []byte) ([]*ethtypes.Log, error) {
	logs := []*ethtypes.Log{}
	err := ModuleCdc.UnmarshalBinaryLengthPrefixed(in, &logs)
	return logs, err
}

// Validate performs a basic validation of a GenesisAccount fields.
func (tx TransactionLogs) Validate() error {
	if bytes.Equal(tx.Hash.Bytes(), ethcmn.Hash{}.Bytes()) {
		return fmt.Errorf("hash cannot be the empty %s", tx.Hash.String())
	}

	for i, log := range tx.Logs {
		if err := ValidateLog(log); err != nil {
			return fmt.Errorf("invalid log %d: %w", i, err)
		}
		if !bytes.Equal(log.TxHash.Bytes(), tx.Hash.Bytes()) {
			return fmt.Errorf("log tx hash mismatch (%s ≠ %s)", log.TxHash.String(), tx.Hash.String())
		}
	}
	return nil
}

// ValidateLog performs a basic validation of an ethereum Log fields.
func ValidateLog(log *ethtypes.Log) error {
	if log == nil {
		return errors.New("log cannot be nil")
	}
	if bytes.Equal(log.Address.Bytes(), ethcmn.Address{}.Bytes()) {
		return fmt.Errorf("log address cannot be empty %s", log.Address.String())
	}
	if bytes.Equal(log.BlockHash.Bytes(), ethcmn.Hash{}.Bytes()) {
		return fmt.Errorf("block hash cannot be the empty %s", log.BlockHash.String())
	}
	if log.BlockNumber == 0 {
		return errors.New("block number cannot be zero")
	}
	if bytes.Equal(log.TxHash.Bytes(), ethcmn.Hash{}.Bytes()) {
		return fmt.Errorf("tx hash cannot be the empty %s", log.TxHash.String())
	}
	return nil
}

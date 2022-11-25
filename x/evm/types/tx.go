package types

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// GetTxPriority returns the priority of a given Ethereum tx. It relies of the
// priority reduction global variable to calculate the tx priority given the tx
// tip price:
//
//	tx_priority = tip_price / priority_reduction
func GetTxPriority(txData TxData, baseFee *big.Int) (priority int64) {
	// calculate priority based on effective gas price
	tipPrice := txData.EffectiveGasPrice(baseFee)
	// if london hardfork is not enabled, tipPrice is the gasPrice
	if baseFee != nil {
		tipPrice = new(big.Int).Sub(tipPrice, baseFee)
	}

	priority = math.MaxInt64
	priorityBig := new(big.Int).Quo(tipPrice, DefaultPriorityReduction.BigInt())

	// safety check
	if priorityBig.IsInt64() {
		priority = priorityBig.Int64()
	}

	return priority
}

// Failed returns if the contract execution failed in vm errors
func (m *MsgEthereumTxResponse) Failed() bool {
	return len(m.VmError) > 0
}

// Return is a helper function to help caller distinguish between revert reason
// and function return. Return returns the data after execution if no error occurs.
func (m *MsgEthereumTxResponse) Return() []byte {
	if m.Failed() {
		return nil
	}
	return common.CopyBytes(m.Ret)
}

// Revert returns the concrete revert reason if the execution is aborted by `REVERT`
// opcode. Note the reason can be nil if no data supplied with revert opcode.
func (m *MsgEthereumTxResponse) Revert() []byte {
	if m.VmError != vm.ErrExecutionReverted.Error() {
		return nil
	}
	return common.CopyBytes(m.Ret)
}

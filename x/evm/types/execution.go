package types

import (
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// GasInfo returns the gas limit, gas consumed and gas refunded from the EVM transition
// execution
type GasInfo struct {
	GasLimit    uint64
	GasConsumed uint64
	GasRefunded uint64
}

// ExecutionResult represents what's returned from a transition
type ExecutionResult struct {
	Logs     []*ethtypes.Log
	Bloom    *big.Int
	Response *MsgEthereumTxResponse
	GasInfo  GasInfo
}

package types

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// copied from: https://github.com/ethereum/go-ethereum/blob/v1.10.3/internal/ethapi/api.go#L754

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From       *common.Address      `json:"from"`
	To         *common.Address      `json:"to"`
	Gas        *hexutil.Uint64      `json:"gas"`
	GasPrice   *hexutil.Big         `json:"gasPrice"`
	Value      *hexutil.Big         `json:"value"`
	Data       *hexutil.Bytes       `json:"data"`
	AccessList *ethtypes.AccessList `json:"accessList"`
}

// ToMessage converts CallArgs to the Message type used by the core evm
func (args *CallArgs) ToMessage(globalGasCap uint64) ethtypes.Message {
	// Set sender address or use zero address if none specified.
	var addr common.Address
	if args.From != nil {
		addr = *args.From
	}

	// Set default gas & gas price if none were set
	gas := globalGasCap
	if gas == 0 {
		gas = uint64(math.MaxUint64 / 2)
	}
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != 0 && globalGasCap < gas {
		// log.Warn("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
		gas = globalGasCap
	}
	gasPrice := new(big.Int)
	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}
	value := new(big.Int)
	if args.Value != nil {
		value = args.Value.ToInt()
	}
	var data []byte
	if args.Data != nil {
		data = *args.Data
	}
	var accessList ethtypes.AccessList
	if args.AccessList != nil {
		accessList = *args.AccessList
	}

	msg := ethtypes.NewMessage(addr, args.To, 0, value, gas, gasPrice, data, accessList, false)
	return msg
}

// String return the struct in a string format
func (args *CallArgs) String() string {
	// Todo: There is currently a bug with hexutil.Big when the value its nil, printing would trigger an exception
	return fmt.Sprintf("CallArgs{From:%v, To:%v, Gas:%v,"+
		" Data:%v, AccessList:%v}",
		args.From,
		args.To,
		args.Gas,
		args.Data,
		args.AccessList)
}

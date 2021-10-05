package types

import (
	"errors"
	"fmt"
	math "math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// TransactionArgs represents the arguments to construct a new transaction
// or a message call using JSON-RPC.
// Duplicate struct definition since geth struct is in internal package
// Ref: https://github.com/ethereum/go-ethereum/blob/release/1.10.4/internal/ethapi/transaction_args.go#L36
type TransactionArgs struct {
	From                 *common.Address `json:"from"`
	To                   *common.Address `json:"to"`
	Gas                  *hexutil.Uint64 `json:"gas"`
	GasPrice             *hexutil.Big    `json:"gasPrice"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas"`
	Value                *hexutil.Big    `json:"value"`
	Nonce                *hexutil.Uint64 `json:"nonce"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://github.com/ethereum/go-ethereum/issues/15628
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`

	// Introduced by AccessListTxType transaction.
	AccessList *ethtypes.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big         `json:"chainId,omitempty"`
}

// String return the struct in a string format
func (args *TransactionArgs) String() string {
	// Todo: There is currently a bug with hexutil.Big when the value its nil, printing would trigger an exception
	return fmt.Sprintf("TransactionArgs{From:%v, To:%v, Gas:%v,"+
		" Nonce:%v, Data:%v, Input:%v, AccessList:%v}",
		args.From,
		args.To,
		args.Gas,
		args.Nonce,
		args.Data,
		args.Input,
		args.AccessList)
}

// ToTransaction converts the arguments to an ethereum transaction.
// This assumes that setTxDefaults has been called.
func (args *TransactionArgs) ToTransaction() *MsgEthereumTx {
	var (
		input                    []byte
		chainID, value, gasPrice *big.Int
		// maxFeePerGas, maxPriorityFeePerGas *big.Int
		addr       common.Address
		gas, nonce uint64
	)

	// Set sender address or use zero address if none specified.
	if args.From != nil {
		addr = *args.From
	}

	if args.Input != nil {
		input = *args.Input
	} else if args.Data != nil {
		input = *args.Data
	}

	if args.ChainID != nil {
		chainID = args.ChainID.ToInt()
	}

	if args.Nonce != nil {
		nonce = uint64(*args.Nonce)
	}

	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}

	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	// if args.MaxFeePerGas != nil {
	// 	maxFeePerGas = args.MaxFeePerGas.ToInt()
	// }

	// if args.MaxPriorityFeePerGas != nil {
	// 	maxPriorityFeePerGas = args.MaxPriorityFeePerGas.ToInt()
	// }

	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	if args.Value != nil {
		value = args.Value.ToInt()
	}

	tx := NewTx(chainID, nonce, args.To, value, gas, gasPrice, input, args.AccessList)
	tx.From = addr.Hex()

	return tx
}

// ToMessage converts the arguments to the Message type used by the core evm.
// This assumes that setTxDefaults has been called.
func (args *TransactionArgs) ToMessage(globalGasCap uint64) (ethtypes.Message, error) {
	var (
		input           []byte
		value, gasPrice *big.Int
		// gasFeeCap, gasTipCap *big.Int
		addr       common.Address
		gas, nonce uint64
	)

	// Reject invalid combinations of pre- and post-1559 fee styles
	if args.GasPrice != nil && (args.MaxFeePerGas != nil || args.MaxPriorityFeePerGas != nil) {
		return ethtypes.Message{}, errors.New("both gasPrice and (maxFeePerGas or maxPriorityFeePerGas) specified")
	}

	// Set sender address or use zero address if none specified.
	if args.From != nil {
		addr = *args.From
	}

	// Set default gas & gas price if none were set
	gas = globalGasCap
	if gas == 0 {
		gas = uint64(math.MaxUint64 / 2)
	}
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != 0 && globalGasCap < gas {
		gas = globalGasCap
	}

	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	if args.Value != nil {
		value = args.Value.ToInt()
	}

	// if args.MaxFeePerGas != nil {
	// 	gasFeeCap = args.MaxFeePerGas.ToInt()
	// }

	// if args.MaxPriorityFeePerGas != nil {
	// 	gasTipCap = args.MaxPriorityFeePerGas.ToInt()
	// }

	if args.Data != nil {
		input = *args.Data
	}
	var accessList ethtypes.AccessList
	if args.AccessList != nil {
		accessList = *args.AccessList
	}

	msg := ethtypes.NewMessage(addr, args.To, nonce, value, gas, gasPrice, input, accessList, false)
	return msg, nil
}

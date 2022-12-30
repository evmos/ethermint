// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"

	evm "github.com/evmos/ethermint/x/evm/vm"
)

var _ evm.Logger = &JSONLogger{}

type JSONLogger struct {
	encoder *json.Encoder
	cfg     *logger.Config
	env     evm.EVM
}

// NewJSONLogger creates a new EVM tracer that prints execution steps as JSON objects
// into the provided stream.
func NewJSONLogger(cfg *logger.Config, writer io.Writer) *JSONLogger {
	l := &JSONLogger{encoder: json.NewEncoder(writer), cfg: cfg}
	if l.cfg == nil {
		l.cfg = &logger.Config{}
	}
	return l
}

func (l *JSONLogger) CaptureStart(env evm.EVM, from, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	l.env = env
}

func (l *JSONLogger) CaptureFault(pc uint64, op vm.OpCode, gas uint64, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	// TODO: Add rData to this interface as well
	l.CaptureState(pc, op, gas, cost, scope, nil, depth, err)
}

// CaptureState outputs state information on the logger.
func (l *JSONLogger) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	memory := scope.Memory
	stack := scope.Stack

	log := logger.StructLog{
		Pc:            pc,
		Op:            op,
		Gas:           gas,
		GasCost:       cost,
		MemorySize:    memory.Len(),
		Depth:         depth,
		RefundCounter: l.env.GetStateDB().GetRefund(),
		Err:           err,
	}
	if l.cfg.EnableMemory {
		log.Memory = memory.Data()
	}
	if !l.cfg.DisableStack {
		log.Stack = stack.Data()
	}
	if l.cfg.EnableReturnData {
		log.ReturnData = rData
	}
	l.encoder.Encode(log)
}

// CaptureEnd is triggered at end of execution.
func (l *JSONLogger) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	type endLog struct {
		Output  string              `json:"output"`
		GasUsed math.HexOrDecimal64 `json:"gasUsed"`
		Time    time.Duration       `json:"time"`
		Err     string              `json:"error,omitempty"`
	}
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	l.encoder.Encode(endLog{common.Bytes2Hex(output), math.HexOrDecimal64(gasUsed), t, errMsg})
}

func (l *JSONLogger) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

func (l *JSONLogger) CaptureExit(output []byte, gasUsed uint64, err error) {}

func (l *JSONLogger) CaptureTxStart(gasLimit uint64) {}

func (l *JSONLogger) CaptureTxEnd(restGas uint64) {}

// formatLogs formats EVM returned structured logs for json output
func formatLogs(logs []logger.StructLog) []logger.StructLogRes {
	formatted := make([]logger.StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = logger.StructLogRes{
			Pc:            trace.Pc,
			Op:            trace.Op.String(),
			Gas:           trace.Gas,
			GasCost:       trace.GasCost,
			Depth:         trace.Depth,
			Error:         trace.ErrorString(),
			RefundCounter: trace.RefundCounter,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = stackValue.Hex()
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

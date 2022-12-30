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
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/holiman/uint256"
)

var _ evm.Logger = &StructLogger{}

// StructLogger is an EVM state logger and implements EVMLogger.
//
// StructLogger can capture state based on the given Log configuration and also keeps
// a track record of modified storage which is used in reporting snapshots of the
// contract their storage.
type StructLogger struct {
	cfg logger.Config
	env evm.EVM

	storage  map[common.Address]logger.Storage
	logs     []logger.StructLog
	output   []byte
	err      error
	gasLimit uint64
	usedGas  uint64

	interrupt uint32 // Atomic flag to signal execution interruption
	reason    error  // Textual reason for the interruption
}

// NewStructLogger returns a new logger
func NewStructLogger(cfg *logger.Config) *StructLogger {
	logger := &StructLogger{
		storage: make(map[common.Address]logger.Storage),
	}
	if cfg != nil {
		logger.cfg = *cfg
	}
	return logger
}

// Reset clears the data held by the logger.
func (l *StructLogger) Reset() {
	l.storage = make(map[common.Address]logger.Storage)
	l.output = make([]byte, 0)
	l.logs = l.logs[:0]
	l.err = nil
}

// CaptureStart implements the EVMLogger interface to initialize the tracing operation.
func (l *StructLogger) CaptureStart(env evm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	l.env = env
}

// CaptureState logs a new structured log message and pushes it out to the environment
//
// CaptureState also tracks SLOAD/SSTORE ops to track storage change.
func (l *StructLogger) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	// If tracing was interrupted, set the error and stop
	if atomic.LoadUint32(&l.interrupt) > 0 {
		l.env.Cancel()
		return
	}
	// check if already accumulated the specified number of logs
	if l.cfg.Limit != 0 && l.cfg.Limit <= len(l.logs) {
		return
	}

	memory := scope.Memory
	stack := scope.Stack
	contract := scope.Contract
	// Copy a snapshot of the current memory state to a new buffer
	var mem []byte
	if l.cfg.EnableMemory {
		mem = make([]byte, len(memory.Data()))
		copy(mem, memory.Data())
	}
	// Copy a snapshot of the current stack state to a new buffer
	var stck []uint256.Int
	if !l.cfg.DisableStack {
		stck = make([]uint256.Int, len(stack.Data()))
		for i, item := range stack.Data() {
			stck[i] = item
		}
	}
	stackData := stack.Data()
	stackLen := len(stackData)
	// Copy a snapshot of the current storage to a new container
	var storage logger.Storage
	if !l.cfg.DisableStorage && (op == vm.SLOAD || op == vm.SSTORE) {
		// initialise new changed values storage container for this contract
		// if not present.
		if l.storage[contract.Address()] == nil {
			l.storage[contract.Address()] = make(logger.Storage)
		}
		// capture SLOAD opcodes and record the read entry in the local storage
		if op == vm.SLOAD && stackLen >= 1 {
			var (
				address = common.Hash(stackData[stackLen-1].Bytes32())
				value   = l.env.GetStateDB().GetState(contract.Address(), address)
			)
			l.storage[contract.Address()][address] = value
			storage = l.storage[contract.Address()].Copy()
		} else if op == vm.SSTORE && stackLen >= 2 {
			// capture SSTORE opcodes and record the written entry in the local storage.
			var (
				value   = common.Hash(stackData[stackLen-2].Bytes32())
				address = common.Hash(stackData[stackLen-1].Bytes32())
			)
			l.storage[contract.Address()][address] = value
			storage = l.storage[contract.Address()].Copy()
		}
	}
	var rdata []byte
	if l.cfg.EnableReturnData {
		rdata = make([]byte, len(rData))
		copy(rdata, rData)
	}
	// create a new snapshot of the EVM.
	log := logger.StructLog{pc, op, gas, cost, mem, memory.Len(), stck, rdata, storage, depth, l.env.GetStateDB().GetRefund(), err}
	l.logs = append(l.logs, log)
}

// CaptureFault implements the EVMLogger interface to trace an execution fault
// while running an opcode.
func (l *StructLogger) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
}

// CaptureEnd is called after the call finishes to finalize the tracing.
func (l *StructLogger) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	l.output = output
	l.err = err
	if l.cfg.Debug {
		fmt.Printf("%#x\n", output)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
		}
	}
}

func (l *StructLogger) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

func (l *StructLogger) CaptureExit(output []byte, gasUsed uint64, err error) {
}

func (l *StructLogger) GetResult() (json.RawMessage, error) {
	// Tracing aborted
	if l.reason != nil {
		return nil, l.reason
	}
	failed := l.err != nil
	returnData := common.CopyBytes(l.output)
	// Return data when successful and revert reason when reverted, otherwise empty.
	returnVal := fmt.Sprintf("%x", returnData)
	if failed && l.err != vm.ErrExecutionReverted {
		returnVal = ""
	}
	return json.Marshal(&logger.ExecutionResult{
		Gas:         l.usedGas,
		Failed:      failed,
		ReturnValue: returnVal,
		StructLogs:  formatLogs(l.StructLogs()),
	})
}

// Stop terminates execution of the tracer at the first opportune moment.
func (l *StructLogger) Stop(err error) {
	l.reason = err
	atomic.StoreUint32(&l.interrupt, 1)
}

func (l *StructLogger) CaptureTxStart(gasLimit uint64) {
	l.gasLimit = gasLimit
}

func (l *StructLogger) CaptureTxEnd(restGas uint64) {
	l.usedGas = l.gasLimit - restGas
}

// StructLogs returns the captured log entries.
func (l *StructLogger) StructLogs() []logger.StructLog { return l.logs }

// Error returns the VM error captured by the trace.
func (l *StructLogger) Error() error { return l.err }

// Output returns the VM return value captured by the trace.
func (l *StructLogger) Output() []byte { return l.output }

// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE

package ethermint

import (
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/vm"
	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/rs/zerolog/log"
)

// keccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type keccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// EVMInterpreter represents an EVM interpreter
type Interpreter struct {
	evm EVM
	cfg config

	hasher    keccakState // Keccak256 hasher instance shared across opcodes
	hasherBuf common.Hash // Keccak256 hasher result array shared aross opcodes

	readOnly   bool   // Whether to throw on stateful modifications
	returnData []byte // Last CALL's return data for subsequent reuse
}

// NewEVMInterpreter returns a new instance of the Interpreter.
func NewEVMInterpreter(evm EVM, cfg config) *evm.Interpreter {
	jumpTable := cfg.JumpTable()

	// If jump table was not initialised we set the default one.
	if cfg.JumpTable() == nil {
		rules := evm.chainRules

		switch {
		case rules.IsMerge:
			jumpTable = &mergeInstructionSet
		case rules.IsLondon:
			jumpTable = &londonInstructionSet
		case rules.IsBerlin:
			jumpTable = &berlinInstructionSet
		case rules.IsIstanbul:
			jumpTable = &istanbulInstructionSet
		case rules.IsConstantinople:
			jumpTable = &constantinopleInstructionSet
		case rules.IsByzantium:
			jumpTable = &byzantiumInstructionSet
		case rules.IsEIP158:
			jumpTable = &spuriousDragonInstructionSet
		case rules.IsEIP150:
			jumpTable = &tangerineWhistleInstructionSet
		case rules.IsHomestead:
			jumpTable = &homesteadInstructionSet
		default:
			jumpTable = &frontierInstructionSet
		}
		extraEips := cfg.ExtraEips()

		for i, eip := range extraEips {
			copy := *jumpTable
			if err := vm.EnableEIP(eip, &copy); err != nil {
				// Disable it, so caller can check if it's activated or not
				extraEips = append(cfg.ExtraEips()[:i], extraEips[i+1:]...)
				log.Error("EIP activation failed", "eip", eip, "error", err)
			}
			cfg.JumpTable = &copy
		}
	}

	return &Interpreter{
		evm: evm,
		cfg: cfg,
	}
}

// Run loops and evaluates the contract's code with the given input data and returns
// the return byte-slice and an error if one occurred.
//
// It's important to note that any errors returned by the interpreter should be
// considered a revert-and-consume-all-gas operation except for
// ErrExecutionReverted which means revert-and-keep-gas-left.
func (in *Interpreter) Run(contract *vm.Contract, input []byte, readOnly bool) (ret []byte, err error) {
	// Increment the call depth which is restricted to 1024
	in.evm.depth++
	defer func() { in.evm.depth-- }()

	// Make sure the readOnly is only set if we aren't in readOnly yet.
	// This also makes sure that the readOnly flag isn't removed for child calls.
	if readOnly && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	// Reset the previous call's return data. It's unimportant to preserve the old buffer
	// as every returning call will return new data anyway.
	in.returnData = nil

	// Don't bother with the execution if there's no code.
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		op   vm.OpCode // current opcode
		cost uint64

		// For optimisation reason we're using uint64 as the program counter.
		// It's theoretically possible to go above 2^64. The YP defines the PC
		// to be uint256. Practically much less so feasible.
		pc uint64 // program counter

		// copies used by tracer
		pcCopy  uint64 // needed for the deferred EVMLogger
		gasCopy uint64 // for EVMLogger to log gas remaining before execution
		logged  bool   // deferred EVMLogger should ignore already logged steps
		res     []byte // result of the opcode execution function
	)

	mem := vm.NewMemory()   // bound memory
	stack := evm.NewStack() // local stack
	callContext := &evm.ScopeContext{
		Memory:   mem,
		Stack:    stack,
		Contract: contract,
	}

	// Don't move this deferred function, it's placed before the capturestate-deferred method,
	// so that it get's executed _after_: the capturestate needs the stacks before
	// they are returned to the pools
	defer func() {
		stack.Return()
	}()
	contract.Input = input

	if in.cfg.debug {
		defer func() {
			if err != nil {
				if !logged {
					in.cfg.Tracer().CaptureState(pcCopy, op, gasCopy, cost, callContext, in.returnData, in.evm.depth, err)
				} else {
					in.cfg.Tracer().CaptureFault(pcCopy, op, gasCopy, cost, callContext, in.evm.depth, err)
				}
			}
		}()
	}
	// The Interpreter main run loop (contextual). This loop runs until either an
	// explicit STOP, RETURN or SELFDESTRUCT is executed, an error occurred during
	// the execution of one of the operations or until the done flag is set by the
	// parent context.
	for {
		if in.cfg.Debug() {
			// Capture pre-execution values for tracing.
			logged, pcCopy, gasCopy = false, pc, contract.Gas
		}
		// Get the operation from the jump table and validate the stack to ensure there are
		// enough stack items available to perform the operation.
		op = contract.GetOp(pc)
		operation := in.cfg.jumpTable[op]
		cost = operation.constantGas // For tracing
		// Validate stack
		if sLen := stack.len(); sLen < operation.minStack {
			return nil, &vm.ErrStackUnderflow{stackLen: sLen, required: operation.minStack}
		} else if sLen > operation.maxStack {
			return nil, &vm.ErrStackOverflow{stackLen: sLen, limit: operation.maxStack}
		}
		if !contract.UseGas(cost) {
			return nil, vm.ErrOutOfGas
		}
		if operation.dynamicGas != nil {
			// All ops with a dynamic memory usage also has a dynamic gas cost.
			var memorySize uint64
			// calculate the new memory size and expand the memory to fit
			// the operation
			// Memory check needs to be done prior to evaluating the dynamic gas portion,
			// to detect calculation overflows
			if operation.memorySize != nil {
				memSize, overflow := operation.memorySize(stack)
				if overflow {
					return nil, vm.ErrGasUintOverflow
				}
				// memory is expanded in words of 32 bytes. Gas
				// is also calculated in words.
				if memorySize, overflow = math.SafeMul(toWordSize(memSize), 32); overflow {
					return nil, vm.ErrGasUintOverflow
				}
			}
			// Consume the gas and return an error if not enough gas is available.
			// cost is explicitly set so that the capture state defer method can get the proper cost
			var dynamicCost uint64
			dynamicCost, err = operation.dynamicGas(in.evm, contract, stack, mem, memorySize)
			cost += dynamicCost // for tracing
			if err != nil || !contract.UseGas(dynamicCost) {
				return nil, vm.ErrOutOfGas
			}

			// Do tracing before memory expansion
			if in.cfg.debug {
				in.cfg.Tracer.CaptureState(pc, op, gasCopy, cost, callContext, in.returnData, in.evm.depth, err)
				logged = true
			}

			if memorySize > 0 {
				mem.Resize(memorySize)
			}
		} else if in.cfg.debug {
			in.cfg.Tracer.CaptureState(pc, op, gasCopy, cost, callContext, in.returnData, in.evm.depth, err)
			logged = true
		}
		// execute the operation
		res, err = operation.execute(&pc, in, callContext)
		if err != nil {
			break
		}
		pc++
	}

	if err == errStopToken {
		err = nil // clear stop token error
	}

	return res, err
}

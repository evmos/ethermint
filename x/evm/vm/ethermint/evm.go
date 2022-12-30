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
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/exp/maps"

	"github.com/evmos/ethermint/x/evm/statedb"
	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/evmos/ethermint/x/evm/vm/geth"
)

var (
	_ evm.EVM         = (*EVM)(nil)
	_ evm.Constructor = NewEVM
)

// EVM is the wrapper for the go-ethereum EVM.
type EVM struct {
	// Context provides auxiliary blockchain related information
	Context   vm.BlockContext
	TxContext vm.TxContext
	// StateDB gives access to the underlying state
	StateDB statedb.ExtStateDB

	// chainConfig contains information about the current chain
	chainConfig *params.ChainConfig
	// chain rules contains the chain rules for the current epoch
	chainRules params.Rules
	// virtual machine configuration options used to initialise the
	// evm.
	Config evm.Config
	// global (to this context) ethereum virtual machine
	// used throughout the execution of the tx.
	interpreter evm.Interpreter
	// abort is used to abort the EVM calling operations
	// NOTE: must be set atomically
	abort int32
	// callGasTemp holds the gas available for the current call. This is needed because the
	// available gas is calculated in gasCall* according to the 63/64 rule and later
	// applied in opCall*.
	callGasTemp uint64

	precompiledContracts evm.PrecompiledContracts
	// Depth is the current call stack
	depth int
}

// NewEVM defines the constructor function for the go-ethereum EVM. It uses the
// default precompiled contracts and the EVM from go-ethereum.
func NewEVM(
	blockCtx vm.BlockContext,
	txCtx vm.TxContext,
	stateDB statedb.ExtStateDB,
	chainConfig *params.ChainConfig,
	config evm.Config,
	pc evm.PrecompiledContracts,
) evm.EVM {
	evm := &EVM{
		Context:              blockCtx,
		TxContext:            txCtx,
		StateDB:              stateDB,
		Config:               config,
		chainConfig:          chainConfig,
		chainRules:           chainConfig.Rules(blockCtx.BlockNumber, blockCtx.Random != nil),
		precompiledContracts: pc,
	}

	return evm
}

// Context returns the EVM's Block Context
func (e EVM) GetContext() vm.BlockContext {
	return e.Context
}

// TxContext returns the EVM's Tx Context
func (e EVM) GetTxContext() vm.TxContext {
	return e.TxContext
}

// Config returns the configuration options for the EVM.
func (e EVM) GetConfig() evm.Config {
	return e.Config
}

// StateDB returns the State Database
func (e EVM) GetStateDB() vm.StateDB {
	return e.StateDB
}

// Reset resets the EVM with a new transaction context.Reset
// This is not threadsafe and should only be done very cautiously.
func (evm *EVM) Reset(txCtx vm.TxContext, statedb statedb.ExtStateDB) {
	evm.TxContext = txCtx
	evm.StateDB = statedb
}

// Cancel cancels any running EVM operation. This may be called concurrently and
// it's safe to be called multiple times.
func (evm *EVM) Cancel() {
	atomic.StoreInt32(&evm.abort, 1)
}

// Cancelled returns true if Cancel has been called
func (evm *EVM) Cancelled() bool {
	return atomic.LoadInt32(&evm.abort) == 1
}

// Interpreter returns the EVM Interpreter instance
func (e EVM) Interpreter() evm.Interpreter {
	return e.interpreter
}

// ChainConfig returns the chain config instance
func (e EVM) ChainConfig() *params.ChainConfig {
	return e.chainConfig
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (e *EVM) Call(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if e.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}

	// Fail if we're trying to transfer more than the available balance
	if value.Sign() != 0 && !e.GetContext().CanTransfer(e.GetStateDB(), caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}

	snapshot := e.GetStateDB().Snapshot()
	p, isPrecompile := e.Precompile(addr)

	if !e.GetStateDB().Exist(addr) {
		if !isPrecompile && e.ChainConfig().Rules(e.Context.BlockNumber, false).IsEIP158 && value.Sign() == 0 {
			// Calling a non existing account, don't do anything, but ping the tracer
			if e.GetConfig().Debug() {
				if e.depth == 0 {
					e.GetConfig().Tracer().CaptureStart(e, caller.Address(), addr, false, input, gas, value)
					e.GetConfig().Tracer().CaptureEnd(ret, 0, 0, nil)
				} else {
					e.GetConfig().Tracer().CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)
					e.GetConfig().Tracer().CaptureExit(ret, 0, nil)
				}
			}
			return nil, gas, nil
		}
		e.GetStateDB().CreateAccount(addr)
	}
	e.GetContext().Transfer(e.GetStateDB(), caller.Address(), addr, value)

	// Capture the tracer start/end events in debug mode
	if e.GetConfig().Debug() {
		if e.depth == 0 {
			e.GetConfig().Tracer().CaptureStart(e, caller.Address(), addr, false, input, gas, value)

			defer func(startGas uint64, startTime time.Time) { // Lazy evaluation of the parameters
				e.GetConfig().Tracer().CaptureEnd(ret, startGas-gas, time.Since(startTime), err)
			}(gas, time.Now())
		} else {
			// Handle tracer events for entering and exiting a call frame
			e.GetConfig().Tracer().CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)

			defer func(startGas uint64) {
				e.GetConfig().Tracer().CaptureExit(ret, startGas-gas, err)
			}(gas)
		}
	}

	if isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, addr, input, gas, value)
	} else {
		// Initialize a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		code := e.GetStateDB().GetCode(addr)
		if len(code) == 0 {
			ret, err = nil, nil // gas is unchanged
		} else {
			addrCopy := addr
			// If the account has no code, we can abort here
			// The depth-check is already done, and precompiles handled above
			contract := vm.NewContract(caller, vm.AccountRef(addrCopy), value, gas)
			contract.SetCallCode(&addrCopy, e.GetStateDB().GetCodeHash(addrCopy), code)
			ret, err = e.Interpreter().Run(contract, input, false)
			gas = contract.Gas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		e.GetStateDB().RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			gas = 0
		}

		// TODO: consider clearing up unused snapshots:
		// } else {
		//	evm.StateDB().DiscardSnapshot(snapshot)
	}
	return ret, gas, err
}

// CallCode executes the contract associated with the addr with the given input
// as parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
//
// CallCode differs from Call in the sense that it executes the given address'
// code with the caller as context.
func (e *EVM) CallCode(
	caller vm.ContractRef,
	addr common.Address,
	input []byte,
	gas uint64,
	value *big.Int,
) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if e.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	// Note although it's noop to transfer X ether to caller itself. But
	// if caller doesn't have enough balance, it would be an error to allow
	// over-charging itself. So the check here is necessary.
	if !e.GetContext().CanTransfer(e.GetStateDB(), caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}

	snapshot := e.GetStateDB().Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if e.GetConfig().Debug() {
		e.GetConfig().Tracer().CaptureEnter(vm.CALLCODE, caller.Address(), addr, input, gas, value)

		defer func(startGas uint64) {
			e.GetConfig().Tracer().CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := e.Precompile(addr); isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, addr, input, gas, value)
	} else {
		addrCopy := addr
		// Initialize a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		contract := vm.NewContract(caller, vm.AccountRef(caller.Address()), value, gas)
		contract.SetCallCode(&addrCopy, e.GetStateDB().GetCodeHash(addrCopy), e.GetStateDB().GetCode(addrCopy))
		ret, err = e.Interpreter().Run(contract, input, false)
		gas = contract.Gas
	}

	if err != nil {
		e.GetStateDB().RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			gas = 0
		}
	}

	return ret, gas, err
}

// Precompile returns the precompiled contract associated with the given address
// and the current chain configuration. If the contract cannot be found it returns
// nil.
func (e EVM) Precompile(addr common.Address) (p vm.PrecompiledContract, found bool) {
	// get the original precompiles from ethereum
	precompiles := geth.GetPrecompiles(e.ChainConfig(), e.GetContext().BlockNumber)
	// copy the custom precompiles entries to the ethereum precompiles map
	maps.Copy(precompiles, e.precompiledContracts)

	p, found = precompiles[addr]
	return p, found
}

// ActivePrecompiles returns a list of all the active precompiled contract addresses
// for the current chain configuration.
func (e EVM) ActivePrecompiles(rules params.Rules) []common.Address {
	precompileAddresses := vm.ActivePrecompiles(rules)

	// add the custom precompiled addresses

	for address := range e.precompiledContracts {
		precompileAddresses = append(precompileAddresses, address)
	}

	return precompileAddresses
}

// RunPrecompileContract runs a stateless precompiled contract and ignores the address and
// value arguments. It uses the RunPrecompiledContract function from the vm package.
func (e *EVM) RunPrecompiledContract(
	pc vm.PrecompiledContract,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	value *big.Int,
) (ret []byte, remainingGas uint64, err error) {
	gasCost := pc.RequiredGas(input)
	if suppliedGas < gasCost {
		return nil, 0, vm.ErrOutOfGas
	}

	sp, isStateful := pc.(evm.StatefulPrecompiledContract)
	if isStateful {
		ret, err = sp.RunStateful(e, addr, input, value)
	} else {
		ret, err = pc.Run(input)
	}

	suppliedGas -= gasCost
	return ret, suppliedGas, err
}

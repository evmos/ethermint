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

package vm

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"

	"github.com/evmos/ethermint/x/evm/statedb"
)

// PrecompiledContracts defines a map of address -> precompiled contract
type PrecompiledContracts map[common.Address]vm.PrecompiledContract

type StatefulPrecompiledContract interface {
	vm.PrecompiledContract
	RunStateful(evm EVM, addr common.Address, input []byte, value *big.Int) (ret []byte, err error)
}

// EVM defines the interface for the Ethereum Virtual Machine used by the EVM module.
type EVM interface {
	GetConfig() Config
	GetContext() vm.BlockContext
	GetTxContext() vm.TxContext
	GetStateDB() vm.StateDB

	Reset(txCtx vm.TxContext, statedb vm.StateDB)
	Cancel()
	Cancelled() bool //nolint
	GetInterpreter() Interpreter
	Call(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error)
	CallCode(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error)
	DelegateCall(caller vm.ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error)
	StaticCall(caller vm.ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error)
	Create(caller vm.ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error)
	Create2(
		caller vm.ContractRef,
		code []byte,
		gas uint64,
		endowment *big.Int,
		salt *uint256.Int) (
		ret []byte, contractAddr common.Address, leftOverGas uint64, err error,
	)
	ChainConfig() *params.ChainConfig

	ActivePrecompiles(rules params.Rules) []common.Address
	Precompile(addr common.Address) (vm.PrecompiledContract, bool)
	RunPrecompiledContract(
		pc vm.PrecompiledContract,
		addr common.Address,
		input []byte,
		suppliedGas uint64,
		value *big.Int) (
		ret []byte, remainingGas uint64, err error,
	)
}

// Interpreter represents an EVM interpreter to run contracts
type Interpreter interface {
	Run(contract *vm.Contract, input []byte, readOnly bool) (ret []byte, err error)
}

// EVMLogger is used to collect execution traces from an EVM transaction
// execution. CaptureState is called for each step of the VM with the
// current VM state.
// Note that reference types are actual VM data structures; make copies
// if you need to retain them beyond the current call.
type Logger interface {
	// Transaction level
	CaptureTxStart(gasLimit uint64)
	CaptureTxEnd(restGas uint64)
	// Top call frame
	CaptureStart(env EVM, from, to common.Address, create bool, input []byte, gas uint64, value *big.Int)
	CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error)
	// Rest of call frames
	CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int)
	CaptureExit(output []byte, gasUsed uint64, err error)
	// Opcode level
	CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *ScopeContext, rData []byte, depth int, err error)
	CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *ScopeContext, depth int, err error)
}

type Config interface {
	// Enables debugging
	Debug() bool
	// Opcode logger
	Tracer() Logger
	// Forces the EIP-1559 baseFee to 0 (needed for 0 price calls)
	NoBaseFee() bool
	// Enables recording of SHA3/keccak preimages
	EnablePreimageRecording() bool
	// EVM instruction table, automatically populated if unset
	JumpTable() *vm.JumpTable
	// Additional EIPS that are to be enabled
	ExtraEips() []int
}

// Constructor defines the function used to instantiate the EVM on
// each state transition.
type Constructor func(
	blockCtx vm.BlockContext,
	txCtx vm.TxContext,
	stateDB statedb.ExtStateDB,
	chainConfig *params.ChainConfig,
	config Config,
	customPrecompiles PrecompiledContracts,
) EVM

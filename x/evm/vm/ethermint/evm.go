package ethermint

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"golang.org/x/exp/maps"

	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/evmos/ethermint/x/evm/vm/geth"
)

var (
	_ evm.EVM         = (*EVM)(nil)
	_ evm.Constructor = NewEVM
)

// EVM is the wrapper for the go-ethereum EVM.
type EVM struct {
	*vm.EVM

	precompiledContracts evm.PrecompiledContracts
	// Depth is the current call stack
	depth int
}

// NewEVM defines the constructor function for the go-ethereum EVM. It uses the
// default precompiled contracts and the EVM from go-ethereum.
func NewEVM(
	blockCtx vm.BlockContext,
	txCtx vm.TxContext,
	stateDB vm.StateDB,
	chainConfig *params.ChainConfig,
	config vm.Config,
	pc evm.PrecompiledContracts,
) evm.EVM {
	return &EVM{
		EVM:                  vm.NewEVM(blockCtx, txCtx, stateDB, chainConfig, config),
		precompiledContracts: pc,
	}
}

// Context returns the EVM's Block Context
func (e EVM) Context() vm.BlockContext {
	return e.EVM.Context
}

// TxContext returns the EVM's Tx Context
func (e EVM) TxContext() vm.TxContext {
	return e.EVM.TxContext
}

// Config returns the configuration options for the EVM.
func (e EVM) Config() vm.Config {
	return e.EVM.Config
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
	if value.Sign() != 0 && !e.Context().CanTransfer(e.StateDB, caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}
	snapshot := e.StateDB.Snapshot()
	p, isPrecompile := e.Precompile(addr)

	if !e.StateDB.Exist(addr) {
		if !isPrecompile && e.ChainConfig().Rules(e.EVM.Context.BlockNumber, false).IsEIP158 && value.Sign() == 0 {
			// Calling a non existing account, don't do anything, but ping the tracer
			if e.Config().Debug {
				if e.depth == 0 {
					e.Config().Tracer.CaptureStart(e.EVM, caller.Address(), addr, false, input, gas, value)
					e.Config().Tracer.CaptureEnd(ret, 0, 0, nil)
				} else {
					e.Config().Tracer.CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)
					e.Config().Tracer.CaptureExit(ret, 0, nil)
				}
			}
			return nil, gas, nil
		}
		e.StateDB.CreateAccount(addr)
	}
	e.Context().Transfer(e.StateDB, caller.Address(), addr, value)

	// Capture the tracer start/end events in debug mode
	if e.Config().Debug {
		if e.depth == 0 {
			e.Config().Tracer.CaptureStart(e.EVM, caller.Address(), addr, false, input, gas, value)
			defer func(startGas uint64, startTime time.Time) { // Lazy evaluation of the parameters
				e.Config().Tracer.CaptureEnd(ret, startGas-gas, time.Since(startTime), err)
			}(gas, time.Now())
		} else {
			// Handle tracer events for entering and exiting a call frame
			e.Config().Tracer.CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)
			defer func(startGas uint64) {
				e.Config().Tracer.CaptureExit(ret, startGas-gas, err)
			}(gas)
		}
	}

	if isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, addr, input, gas, value)
	} else {
		// Initialise a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		code := e.StateDB.GetCode(addr)
		if len(code) == 0 {
			ret, err = nil, nil // gas is unchanged
		} else {
			addrCopy := addr
			// If the account has no code, we can abort here
			// The depth-check is already done, and precompiles handled above
			contract := vm.NewContract(caller, vm.AccountRef(addrCopy), value, gas)
			contract.SetCallCode(&addrCopy, e.StateDB.GetCodeHash(addrCopy), code)
			ret, err = e.Interpreter().Run(contract, input, false)
			gas = contract.Gas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		e.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			gas = 0
		}
		// TODO: consider clearing up unused snapshots:
		//} else {
		//	evm.StateDB.DiscardSnapshot(snapshot)
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
	if !e.Context().CanTransfer(e.StateDB, caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}
	snapshot := e.StateDB.Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if e.Config().Debug {
		e.Config().Tracer.CaptureEnter(vm.CALLCODE, caller.Address(), addr, input, gas, value)
		defer func(startGas uint64) {
			e.Config().Tracer.CaptureExit(ret, startGas-gas, err)
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
		contract.SetCallCode(&addrCopy, e.StateDB.GetCodeHash(addrCopy), e.StateDB.GetCode(addrCopy))
		ret, err = e.Interpreter().Run(contract, input, false)
		gas = contract.Gas
	}

	if err != nil {
		e.StateDB.RevertToSnapshot(snapshot)
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
	precompiles := geth.GetPrecompiles(e.ChainConfig(), e.Context().BlockNumber)
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

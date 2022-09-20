package ethermint

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	eevm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/evmos/ethermint/x/evm/vm/geth"
)

var (
	_ eevm.EVM         = (*EVM)(nil)
	_ eevm.Constructor = NewEVM
)

// EVM is the wrapper for the go-ethereum EVM.
type EVM struct {
	*vm.EVM

	precompiledContracts eevm.PrecompiledContracts
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
	pc eevm.PrecompiledContracts,
) eevm.EVM {
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
func (evm *EVM) Call(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	if value.Sign() != 0 && !evm.Context().CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}
	snapshot := evm.StateDB.Snapshot()
	p, isPrecompile := evm.Precompile(addr)

	if !evm.StateDB.Exist(addr) {
		if !isPrecompile && evm.ChainConfig().Rules(evm.EVM.Context.BlockNumber, false).IsEIP158 && value.Sign() == 0 {
			// Calling a non existing account, don't do anything, but ping the tracer
			if evm.Config().Debug {
				if evm.depth == 0 {
					evm.Config().Tracer.CaptureStart(evm.EVM, caller.Address(), addr, false, input, gas, value)
					evm.Config().Tracer.CaptureEnd(ret, 0, 0, nil)
				} else {
					evm.Config().Tracer.CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)
					evm.Config().Tracer.CaptureExit(ret, 0, nil)
				}
			}
			return nil, gas, nil
		}
		evm.StateDB.CreateAccount(addr)
	}
	evm.Context().Transfer(evm.StateDB, caller.Address(), addr, value)

	// Capture the tracer start/end events in debug mode
	if evm.Config().Debug {
		if evm.depth == 0 {
			evm.Config().Tracer.CaptureStart(evm.EVM, caller.Address(), addr, false, input, gas, value)
			defer func(startGas uint64, startTime time.Time) { // Lazy evaluation of the parameters
				evm.Config().Tracer.CaptureEnd(ret, startGas-gas, time.Since(startTime), err)
			}(gas, time.Now())
		} else {
			// Handle tracer events for entering and exiting a call frame
			evm.Config().Tracer.CaptureEnter(vm.CALL, caller.Address(), addr, input, gas, value)
			defer func(startGas uint64) {
				evm.Config().Tracer.CaptureExit(ret, startGas-gas, err)
			}(gas)
		}
	}

	if isPrecompile {
		ret, gas, err = evm.RunPrecompiledContract(p, addr, input, gas, value)
	} else {
		// Initialise a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		code := evm.StateDB.GetCode(addr)
		if len(code) == 0 {
			ret, err = nil, nil // gas is unchanged
		} else {
			addrCopy := addr
			// If the account has no code, we can abort here
			// The depth-check is already done, and precompiles handled above
			contract := vm.NewContract(caller, vm.AccountRef(addrCopy), value, gas)
			contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), code)
			ret, err = evm.Interpreter().Run(contract, input, false)
			gas = contract.Gas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
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
func (evm *EVM) CallCode(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	// Note although it's noop to transfer X ether to caller itself. But
	// if caller doesn't have enough balance, it would be an error to allow
	// over-charging itself. So the check here is necessary.
	if !evm.Context().CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, gas, vm.ErrInsufficientBalance
	}
	snapshot := evm.StateDB.Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if evm.Config().Debug {
		evm.Config().Tracer.CaptureEnter(vm.CALLCODE, caller.Address(), addr, input, gas, value)
		defer func(startGas uint64) {
			evm.Config().Tracer.CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := evm.Precompile(addr); isPrecompile {
		ret, gas, err = evm.RunPrecompiledContract(p, addr, input, gas, value)
	} else {
		addrCopy := addr
		// Initialize a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		contract := vm.NewContract(caller, vm.AccountRef(caller.Address()), value, gas)
		contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), evm.StateDB.GetCode(addrCopy))
		ret, err = evm.Interpreter().Run(contract, input, false)
		gas = contract.Gas
	}

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			gas = 0
		}
	}

	return ret, gas, err
}

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (evm *EVM) DelegateCall(caller vm.ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}

	snapshot := evm.StateDB.Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if evm.Config().Debug {
		evm.Config().Tracer.CaptureEnter(vm.DELEGATECALL, caller.Address(), addr, input, gas, nil)
		defer func(startGas uint64) {
			evm.Config().Tracer.CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	// It is allowed to call precompiles, even via delegate call
	p, isPrecompile := evm.Precompile(addr)

	if isPrecompile && p.(vm.StatefulPrecompiledContract) {
		ret, gas, err = evm.RunPrecompiledContract(p, addr, input, gas, nil) // TODO: should value be
	} else {
		addrCopy := addr
		// Initialize a new contract and make initialize the delegate values
		contract := vm.NewContract(caller, vm.AccountRef(caller.Address()), nil, gas).AsDelegate()
		contract.SetCallCode(&addrCopy, evm.StateDB.GetCodeHash(addrCopy), evm.StateDB.GetCode(addrCopy))
		ret, err = evm.Interpreter().Run(contract, input, false)
		gas = contract.Gas
	}

	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
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
	precompiles := geth.GetPrecompiles(e.ChainConfig(), e.Context().BlockNumber)
	p, found = precompiles[addr]
	return p, found
}

// ActivePrecompiles returns a list of all the active precompiled contract addresses
// for the current chain configuration.
func (EVM) ActivePrecompiles(rules params.Rules) []common.Address {
	// TODO: add new precompiles
	return vm.ActivePrecompiles(rules)
}

// RunPrecompileContract runs a stateless precompiled contract and ignores the address and
// value arguments. It uses the RunPrecompiledContract function from the vm package.
func (e *EVM) RunPrecompiledContract(pc vm.PrecompiledContract, addr common.Address, input []byte, suppliedGas uint64, value *big.Int) (ret []byte, remainingGas uint64, err error) {
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

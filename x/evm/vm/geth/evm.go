package geth

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

var (
	// emptyCodeHash is used by create to ensure deployment is disallowed to already
	// deployed contract addresses (relevant after the account abstraction).
	emptyCodeHash = crypto.Keccak256Hash(nil)

	big0 = big.NewInt(0)
)

type codeAndHash struct {
	code []byte
	hash common.Hash
}

func (c *codeAndHash) Hash() common.Hash {
	if c.hash == (common.Hash{}) {
		c.hash = crypto.Keccak256Hash(c.code)
	}
	return c.hash
}

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (e EVM) Call(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
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
		chainRules := e.ChainConfig().Rules(e.Context().BlockNumber, e.Context().Random != nil)
		if !isPrecompile && chainRules.IsEIP158 && value.Sign() == 0 {
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
		ret, gas, err = e.RunPrecompiledContract(p, caller.Address(), input, gas, value)
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
			ret, err = e.RunInterpreter(contract, input, false)
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
		//	e.StateDB.DiscardSnapshot(snapshot)
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
func (e EVM) CallCode(caller vm.ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
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
	var snapshot = e.StateDB.Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if e.Config().Debug {
		e.Config().Tracer.CaptureEnter(vm.CALLCODE, caller.Address(), addr, input, gas, value)
		defer func(startGas uint64) {
			e.Config().Tracer.CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := e.Precompile(addr); isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, caller.Address(), input, gas, value)
	} else {
		addrCopy := addr
		// Initialise a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		contract := vm.NewContract(caller, vm.AccountRef(caller.Address()), value, gas)
		contract.SetCallCode(&addrCopy, e.StateDB.GetCodeHash(addrCopy), e.StateDB.GetCode(addrCopy))
		ret, err = e.RunInterpreter(contract, input, false)
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

// DelegateCall executes the contract associated with the addr with the given input
// as parameters. It reverses the state in case of an execution error.
//
// DelegateCall differs from CallCode in the sense that it executes the given address'
// code with the caller as context and the caller is set to the caller of the caller.
func (e EVM) DelegateCall(caller vm.ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if e.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}
	var snapshot = e.StateDB.Snapshot()

	// Invoke tracer hooks that signal entering/exiting a call frame
	if e.Config().Debug {
		e.Config().Tracer.CaptureEnter(vm.DELEGATECALL, caller.Address(), addr, input, gas, nil)
		defer func(startGas uint64) {
			e.Config().Tracer.CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	// It is allowed to call precompiles, even via delegatecall
	if p, isPrecompile := e.Precompile(addr); isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, caller.Address(), input, gas, big0)
	} else {
		addrCopy := addr
		// Initialise a new contract and make initialise the delegate values
		contract := vm.NewContract(caller, vm.AccountRef(caller.Address()), nil, gas).AsDelegate()
		contract.SetCallCode(&addrCopy, e.StateDB.GetCodeHash(addrCopy), e.StateDB.GetCode(addrCopy))
		ret, err = e.RunInterpreter(contract, input, false)
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

// StaticCall executes the contract associated with the addr with the given input
// as parameters while disallowing any modifications to the state during the call.
// Opcodes that attempt to perform such modifications will result in exceptions
// instead of performing the modifications.
func (e EVM) StaticCall(caller vm.ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to execute above the call depth limit
	if e.depth > int(params.CallCreateDepth) {
		return nil, gas, vm.ErrDepth
	}
	// We take a snapshot here. This is a bit counter-intuitive, and could probably be skipped.
	// However, even a staticcall is considered a 'touch'. On mainnet, static calls were introduced
	// after all empty accounts were deleted, so this is not required. However, if we omit this,
	// then certain tests start failing; stRevertTest/RevertPrecompiledTouchExactOOG.json.
	// We could change this, but for now it's left for legacy reasons
	var snapshot = e.StateDB.Snapshot()

	// We do an AddBalance of zero here, just in order to trigger a touch.
	// This doesn't matter on Mainnet, where all empties are gone at the time of Byzantium,
	// but is the correct thing to do and matters on other networks, in tests, and potential
	// future scenarios
	e.StateDB.AddBalance(addr, big0)

	// Invoke tracer hooks that signal entering/exiting a call frame
	if e.Config().Debug {
		e.Config().Tracer.CaptureEnter(vm.STATICCALL, caller.Address(), addr, input, gas, nil)
		defer func(startGas uint64) {
			e.Config().Tracer.CaptureExit(ret, startGas-gas, err)
		}(gas)
	}

	if p, isPrecompile := e.Precompile(addr); isPrecompile {
		ret, gas, err = e.RunPrecompiledContract(p, caller.Address(), input, gas, big0)
	} else {
		// At this point, we use a copy of address. If we don't, the go compiler will
		// leak the 'contract' to the outer scope, and make allocation for 'contract'
		// even if the actual execution ends on RunPrecompiled above.
		addrCopy := addr
		// Initialise a new contract and set the code that is to be used by the EVM.
		// The contract is a scoped environment for this execution context only.
		contract := vm.NewContract(caller, vm.AccountRef(addrCopy), new(big.Int), gas)
		contract.SetCallCode(&addrCopy, e.StateDB.GetCodeHash(addrCopy), e.StateDB.GetCode(addrCopy))
		// When an error was returned by the EVM or when setting the creation code
		// above we revert to the snapshot and consume any gas remaining. Additionally
		// when we're in Homestead this also counts for code storage gas errors.
		ret, err = e.RunInterpreter(contract, input, true)
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

// create creates a new contract using code as deployment code.
func (e EVM) create(caller vm.ContractRef, codeAndHash *codeAndHash, gas uint64, value *big.Int, address common.Address, typ vm.OpCode) ([]byte, common.Address, uint64, error) {
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if e.depth > int(params.CallCreateDepth) {
		return nil, common.Address{}, gas, vm.ErrDepth
	}
	if !e.Context().CanTransfer(e.StateDB, caller.Address(), value) {
		return nil, common.Address{}, gas, vm.ErrInsufficientBalance
	}
	nonce := e.StateDB.GetNonce(caller.Address())
	if nonce+1 < nonce {
		return nil, common.Address{}, gas, vm.ErrNonceUintOverflow
	}
	e.StateDB.SetNonce(caller.Address(), nonce+1)
	// We add this to the access list _before_ taking a snapshot. Even if the creation fails,
	// the access-list change should not be rolled back
	chainRules := e.ChainConfig().Rules(e.Context().BlockNumber, e.Context().Random != nil)
	if chainRules.IsBerlin {
		e.StateDB.AddAddressToAccessList(address)
	}
	// Ensure there's no existing contract already at the designated address
	contractHash := e.StateDB.GetCodeHash(address)
	if e.StateDB.GetNonce(address) != 0 || (contractHash != (common.Hash{}) && contractHash != emptyCodeHash) {
		return nil, common.Address{}, 0, vm.ErrContractAddressCollision
	}
	// Create a new account on the state
	snapshot := e.StateDB.Snapshot()
	e.StateDB.CreateAccount(address)
	if chainRules.IsEIP158 {
		e.StateDB.SetNonce(address, 1)
	}
	e.Context().Transfer(e.StateDB, caller.Address(), address, value)

	// Initialise a new contract and set the code that is to be used by the EVM.
	// The contract is a scoped environment for this execution context only.
	contract := vm.NewContract(caller, vm.AccountRef(address), value, gas)
	contract.Code = codeAndHash.code
	contract.CodeHash = codeAndHash.hash
	contract.CodeAddr = &address

	if e.Config().Debug {
		if e.depth == 0 {
			e.Config().Tracer.CaptureStart(e.EVM, caller.Address(), address, true, codeAndHash.code, gas, value)
		} else {
			e.Config().Tracer.CaptureEnter(typ, caller.Address(), address, codeAndHash.code, gas, value)
		}
	}

	start := time.Now()

	ret, err := e.RunInterpreter(contract, nil, false)

	// Check whether the max code size has been exceeded, assign err if the case.
	if err == nil && chainRules.IsEIP158 && len(ret) > params.MaxCodeSize {
		err = vm.ErrMaxCodeSizeExceeded
	}

	// Reject code starting with 0xEF if EIP-3541 is enabled.
	if err == nil && len(ret) >= 1 && ret[0] == 0xEF && chainRules.IsLondon {
		err = vm.ErrInvalidCode
	}

	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil {
		createDataGas := uint64(len(ret)) * params.CreateDataGas
		if contract.UseGas(createDataGas) {
			e.StateDB.SetCode(address, ret)
		} else {
			err = vm.ErrCodeStoreOutOfGas
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil && (chainRules.IsHomestead || err != vm.ErrCodeStoreOutOfGas) {
		e.StateDB.RevertToSnapshot(snapshot)
		if err != vm.ErrExecutionReverted {
			contract.UseGas(contract.Gas)
		}
	}

	if e.Config().Debug {
		if e.depth == 0 {
			e.Config().Tracer.CaptureEnd(ret, gas-contract.Gas, time.Since(start), err)
		} else {
			e.Config().Tracer.CaptureExit(ret, gas-contract.Gas, err)
		}
	}
	return ret, address, contract.Gas, err
}

// Create creates a new contract using code as deployment code.
func (e EVM) Create(caller vm.ContractRef, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	contractAddr = crypto.CreateAddress(caller.Address(), e.StateDB.GetNonce(caller.Address()))
	return e.create(caller, &codeAndHash{code: code}, gas, value, contractAddr, vm.CREATE)
}

// Create2 creates a new contract using code as deployment code.
//
// The different between Create2 with Create is Create2 uses keccak256(0xff ++ msg.sender ++ salt ++ keccak256(init_code))[12:]
// instead of the usual sender-and-nonce-hash as the address where the contract is initialized at.
func (e EVM) Create2(caller vm.ContractRef, code []byte, gas uint64, endowment *big.Int, salt *uint256.Int) (ret []byte, contractAddr common.Address, leftOverGas uint64, err error) {
	codeAndHash := &codeAndHash{code: code}
	contractAddr = crypto.CreateAddress2(caller.Address(), salt.Bytes32(), codeAndHash.Hash().Bytes())
	return e.create(caller, codeAndHash, gas, endowment, contractAddr, vm.CREATE2)
}

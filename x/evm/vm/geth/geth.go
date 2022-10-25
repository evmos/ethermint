package geth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	evm "github.com/evmos/ethermint/x/evm/vm"
)

var (
	_ evm.EVM         = (*EVM)(nil)
	_ evm.Constructor = NewEVM
)

// EVM is the wrapper for the go-ethereum EVM.
type EVM struct {
	*vm.EVM

	// call depth for Ethermint EVM interpreter
	depth int

	// custom precompiles
	precompiles *evm.PrecompiledContracts
}

// NewEVM defines the constructor function for the go-ethereum (geth) EVM. It uses
// the default precompiled contracts and the EVM concrete implementation from
// geth.
func NewEVM(
	blockCtx vm.BlockContext,
	txCtx vm.TxContext,
	stateDB vm.StateDB,
	chainConfig *params.ChainConfig,
	config vm.Config,
	precompiles *evm.PrecompiledContracts,
) evm.EVM {
	return &EVM{
		EVM:         vm.NewEVM(blockCtx, txCtx, stateDB, chainConfig, config),
		precompiles: precompiles,
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

// RunInterpreter runs the EVM interpreter with manual tracking of call depth
func (e EVM) RunInterpreter(contract *vm.Contract, input []byte, readOnly bool) (ret []byte, err error) {
	e.depth++
	defer func() { e.depth-- }()

	return e.Interpreter().Run(contract, input, readOnly)
}

// Precompile returns the precompiled contract associated with the given address
// and the current chain configuration. If the contract cannot be found it returns
// nil.
func (e EVM) Precompile(addr common.Address) (p vm.PrecompiledContract, found bool) {
	precompiles := GetPrecompiles(e.ChainConfig(), e.EVM.Context.BlockNumber)
	p, found = precompiles[addr]
	if !found {
		p, found = (*e.precompiles)[addr]
	}
	return
}

// ActivePrecompiles returns a list of all the active precompiled contract addresses
// for the current chain configuration.
func (e EVM) ActivePrecompiles(rules params.Rules) []common.Address {
	addrs := make([]common.Address, len(*e.precompiles))
	for addr := range *e.precompiles {
		addrs = append(addrs, addr)
	}
	return append(vm.ActivePrecompiles(rules), addrs...)
}

// RunPrecompiledContract runs a precompiled contract. It uses the RunPrecompiledContract
// function from the geth vm package for stateless precompiles and InitStateful for stateful.
func (e EVM) RunPrecompiledContract(
	p vm.PrecompiledContract,
	addr common.Address,
	input []byte,
	suppliedGas uint64,
	value *big.Int,
) (ret []byte, remainingGas uint64, err error) {
	if precompile, isStateful := p.(evm.StatefulPrecompiledContract); isStateful {
		gasCost := precompile.RequiredGas(input)
		if suppliedGas < gasCost {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= gasCost
		output, err := e.InitStateful(precompile, input, addr, value)
		return output, suppliedGas, err
	}
	return vm.RunPrecompiledContract(p, input, suppliedGas)
}

func (e EVM) InitStateful(precompile evm.StatefulPrecompiledContract, input []byte, caller common.Address, value *big.Int) ([]byte, error) {
	output, err := precompile.RunStateful(e, caller, input, value)
	if err != nil {
		// revert this contract's state changes
		return nil, err
	}

	// precompile executed successully
	return output, nil
}

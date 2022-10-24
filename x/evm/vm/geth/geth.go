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
	precompiles evm.PrecompiledContracts
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
	precompiles evm.PrecompiledContracts,
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

// Precompile returns the precompiled contract associated with the given address
// and the current chain configuration. If the contract cannot be found it returns
// nil.
func (e EVM) Precompile(addr common.Address) (p vm.PrecompiledContract, found bool) {
	precompiles := GetPrecompiles(e.ChainConfig(), e.EVM.Context.BlockNumber)
	p, found = precompiles[addr]
	if !found {
		p, found = e.precompiles[addr]
	}
	return
}

// ActivePrecompiles returns a list of all the active precompiled contract addresses
// for the current chain configuration.
func (EVM) ActivePrecompiles(rules params.Rules) []common.Address {
	return vm.ActivePrecompiles(rules)
}

// RunPrecompiledContract runs a stateless precompiled contract and ignores the address and
// value arguments. It uses the RunPrecompiledContract function from the geth vm package.
func (e EVM) RunPrecompiledContract(
	p vm.PrecompiledContract,
	addr common.Address, // address arg is unused
	input []byte,
	suppliedGas uint64,
	value *big.Int, // 	value arg is unused
) (ret []byte, remainingGas uint64, err error) {
	precompile, isStateful := p.(evm.StatefulPrecompiledContract)
	if isStateful {
		gasCost := precompile.RequiredGas(input)
		if suppliedGas < gasCost {
			return nil, 0, vm.ErrOutOfGas
		}
		suppliedGas -= gasCost
		output, err := precompile.RunStateful(e, addr, input, value)
		return output, suppliedGas, err
	}
	return vm.RunPrecompiledContract(p, input, suppliedGas)
}

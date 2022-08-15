package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

var (
	_ EVM         = (*Geth)(nil)
	_ Constructor = NewGeth
)

// Geth is the wrapper for the go-ethereum EVM.
type Geth struct {
	*vm.EVM
}

// NewGeth defines the constructor function for the go-ethereum EVM.
func NewGeth(
	blockCtx vm.BlockContext,
	txCtx vm.TxContext,
	stateDB vm.StateDB,
	chainConfig *params.ChainConfig,
	config vm.Config,
	_ PrecompiledContracts,
) EVM {
	return &Geth{
		EVM: vm.NewEVM(blockCtx, txCtx, stateDB, chainConfig, config),
	}
}

// Context returns the EVM's Block Context
func (g Geth) Context() vm.BlockContext {
	return g.EVM.Context
}

// Context returns the EVM's Tx Context
func (g Geth) TxContext() vm.TxContext {
	return g.EVM.TxContext
}

// Config returns the Config for the EVM.
func (g Geth) Config() vm.Config {
	return g.EVM.Config
}

// Precompile returns the precompiled contract associated with the given address
// and the current chain configuration. If the contract cannot be found it returns
// nil.
func (g Geth) Precompile(addr common.Address) (p vm.PrecompiledContract, found bool) {
	precompiles := getPrecompiles(g.ChainConfig(), g.Context().BlockNumber)
	p, found = precompiles[addr]
	return p, found
}

// ActivePrecompiles returns a list of all the active precompiled contract addresses
// for the current chain configuration.
func (g Geth) ActivePrecompiles(rules params.Rules) []common.Address {
	return vm.ActivePrecompiles(rules)
}

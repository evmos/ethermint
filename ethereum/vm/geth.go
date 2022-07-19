package vm

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

var _ EVM = (*GethEVM)(nil)

// GethEVM is the wrapper for the Geth EVM.
type GethEVM struct {
	*vm.EVM
}

func NewGethEVM(blockCtx vm.BlockContext, txCtx vm.TxContext, stateDB vm.StateDB, chainConfig *params.ChainConfig, config vm.Config) *GethEVM {
	return &GethEVM{
		EVM: vm.NewEVM(blockCtx, txCtx, stateDB, chainConfig, config),
	}
}

func (evm GethEVM) Context() vm.BlockContext {
	return evm.EVM.Context
}

func (evm GethEVM) TxContext() vm.TxContext {
	return evm.EVM.TxContext
}

func (evm GethEVM) Config() vm.Config {
	return evm.EVM.Config
}

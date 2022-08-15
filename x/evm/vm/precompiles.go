package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

type PrecompiledContracts map[common.Address]vm.PrecompiledContract

// getPrecompiles returns all the precompiled contracts defined given the
// current chain configuration and block height.
func getPrecompiles(cfg *params.ChainConfig, blockNumber *big.Int) PrecompiledContracts {
	var precompiles PrecompiledContracts
	switch {
	case cfg.IsBerlin(blockNumber):
		precompiles = vm.PrecompiledContractsBerlin
	case cfg.IsIstanbul(blockNumber):
		precompiles = vm.PrecompiledContractsIstanbul
	case cfg.IsByzantium(blockNumber):
		precompiles = vm.PrecompiledContractsByzantium
	default:
		precompiles = vm.PrecompiledContractsHomestead
	}
	return precompiles
}

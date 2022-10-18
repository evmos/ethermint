package geth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	evm "github.com/evmos/ethermint/x/evm/vm"
)

// GetPrecompiles returns all the precompiled contracts defined given the
// current chain configuration and block height.
func GetPrecompiles(cfg *params.ChainConfig, blockNumber *big.Int) evm.PrecompiledContracts {
	var precompiles evm.PrecompiledContracts
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

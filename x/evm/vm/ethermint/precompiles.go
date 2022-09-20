package ethermint

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	evm "github.com/evmos/ethermint/x/evm/vm"
	"github.com/evmos/ethermint/x/evm/vm/geth"
	"golang.org/x/exp/maps"
)

// ActivePrecompiles returns the precompiles enabled with the current configuration.
// TODO: add the precompiles
func ActivePrecompiles(rules params.Rules) []common.Address {
	switch {
	case rules.IsBerlin:
		return vm.PrecompiledAddressesBerlin
	case rules.IsIstanbul:
		return vm.PrecompiledAddressesIstanbul
	case rules.IsByzantium:
		return vm.PrecompiledAddressesByzantium
	default:
		return vm.PrecompiledAddressesHomestead
	}
}

// GetPrecompiles returns all the precompiled contracts defined given the
// current chain configuration and block height.
func GetPrecompiles(cfg *params.ChainConfig, blockNumber *big.Int, customPrecompiles evm.PrecompiledContracts) evm.PrecompiledContracts {
	// get the original precompiles from ethereum
	precompiles := geth.GetPrecompiles(cfg, blockNumber)
	// copy the custom precompiles entries to the ethereum precompiles map
	maps.Copy(precompiles, customPrecompiles)

	return precompiles
}

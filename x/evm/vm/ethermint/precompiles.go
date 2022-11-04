package ethermint

import (
	"bytes"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"golang.org/x/exp/maps"
)

var (
	// StakingPrecompileAddress is 0x0000000000000000000000000000000000000100
	StakingPrecompileAddress common.Address = common.BytesToAddress([]byte{100})
	// DistributionPrecompileAddress is 0x0000000000000000000000000000000000000101
	DistributionPrecompileAddress common.Address = common.BytesToAddress([]byte{101})
	// ICS20PrecompileAddress is 0x0000000000000000000000000000000000000102
	ICS20PrecompileAddress common.Address = common.BytesToAddress([]byte{102})
)

var (
	// PrecompiledAddressesEthermint defines the precompiled contracts used by Ethermint
	PrecompiledAddressesEthermint []common.Address

	// PrecompiledContractsEthermint contains the default set of pre-compiled Ethermint
	// contracts used in production.
	PrecompiledContractsEthermint map[common.Address]vm.PrecompiledContract
)

func NewPrecompiles() {
	// gosec: nosec
	for k := range PrecompiledContractsEthermint {
		PrecompiledAddressesEthermint = append(PrecompiledAddressesEthermint, k)
	}

	// PrecompiledContractsEthermint[StakingPrecompileAddress] = precompiles.NewStakingPrecompile()
	// PrecompiledContractsEthermint[DistributionPrecompileAddress] = precompiles.NewDistributionPrecompile()
	// PrecompiledContractsEthermint[ICS20PrecompileAddress] = precompiles.NewICS20Precompile()

	// sort the precompiled addresses
	sort.Slice(PrecompiledAddressesEthermint, func(i, j int) bool {
		return bytes.Compare(
			PrecompiledAddressesEthermint[i].Bytes(),
			PrecompiledAddressesEthermint[j].Bytes(),
		) == -1
	})
}

func init() {
	// copy latest Ethereum precompiles to Ethermint precompile map
	maps.Copy(PrecompiledContractsEthermint, vm.PrecompiledContractsBerlin)
}

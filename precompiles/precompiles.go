// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE

package ethermint

import (
	"bytes"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"golang.org/x/exp/maps"

	"github.com/evmos/ethermint/precompiles/ibc/transfer"
	"github.com/evmos/ethermint/precompiles/staking"
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

	PrecompiledContractsEthermint[StakingPrecompileAddress] = staking.NewStakingPrecompile()
	// PrecompiledContractsEthermint[DistributionPrecompileAddress] = precompiles.NewDistributionPrecompile()
	PrecompiledContractsEthermint[ICS20PrecompileAddress] = transfer.NewICS20Precompile()

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

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
	"github.com/ethereum/go-ethereum/core/vm"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

var _ evm.Config = &config{}

// config defines the configuration options for the Interpreter
type config struct {
	debug                   bool       // Enables debugging
	tracer                  evm.Logger // Opcode logger
	noBaseFee               bool       // Forces the EIP-1559 baseFee to 0 (needed for 0 price calls)
	enablePreimageRecording bool       // Enables recording of SHA3/keccak preimages

	jumpTable *vm.JumpTable // EVM instruction table, automatically populated if unset

	extraEips []int // Additional EIPS that are to be enabled
}

func NewConfig(
	debug bool,
	tracer evm.Logger,
	noBaseFee,
	enablePreimageRecording bool,
	jumpTable *vm.JumpTable,
	extraEips ...int,
) evm.Config {
	return &config{
		debug:                   debug,
		tracer:                  tracer,
		noBaseFee:               noBaseFee,
		enablePreimageRecording: enablePreimageRecording,
		jumpTable:               jumpTable,
		extraEips:               extraEips,
	}
}

func (c *config) Debug() bool { return c.debug }

func (c *config) Tracer() evm.Logger { return c.tracer }

func (c *config) NoBaseFee() bool { return c.noBaseFee }

func (c *config) EnablePreimageRecording() bool { return c.enablePreimageRecording }

func (c *config) JumpTable() *vm.JumpTable { return c.jumpTable }

func (c *config) ExtraEips() []int { return c.extraEips }

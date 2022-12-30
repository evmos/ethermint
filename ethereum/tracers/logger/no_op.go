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

package logger

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

var _ evm.Logger = &NoOpTracer{}

// NoOpTracer is an empty implementation of vm.Tracer interface
type NoOpTracer struct{}

// NewNoOpTracer creates a no-op vm.Tracer
func NewNoOpTracer() *NoOpTracer {
	return &NoOpTracer{}
}

// CaptureStart implements vm.Tracer interface
func (dt NoOpTracer) CaptureStart(
	env evm.EVM,
	from,
	to common.Address,
	create bool,
	input []byte,
	gas uint64,
	value *big.Int,
) {
}

// CaptureState implements vm.Tracer interface
func (dt NoOpTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
}

// CaptureFault implements vm.Tracer interface
func (dt NoOpTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
}

// CaptureEnd implements vm.Tracer interface
func (dt NoOpTracer) CaptureEnd(output []byte, gasUsed uint64, tm time.Duration, err error) {}

// CaptureEnter implements vm.Tracer interface
func (dt NoOpTracer) CaptureEnter(typ vm.OpCode, from, to common.Address, input []byte, gas uint64, value *big.Int) {
}

// CaptureExit implements vm.Tracer interface
func (dt NoOpTracer) CaptureExit(output []byte, gasUsed uint64, err error) {}

// CaptureTxStart implements vm.Tracer interface
func (dt NoOpTracer) CaptureTxStart(gasLimit uint64) {}

// CaptureTxEnd implements vm.Tracer interface
func (dt NoOpTracer) CaptureTxEnd(restGas uint64) {}

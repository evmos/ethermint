package types

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// This file was created to maintain a public type of NoopTracer based on
// Go Ethereum v 1.10.12 tracer's implementation. Having this public type will prevent
// big changes throughout this codebase

// NoopTracer is a go implementation of the Tracer interface which
// performs no action.
type NoopTracer struct{}

func (t NoopTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
}

func (t NoopTracer) CaptureEnd(output []byte, gasUsed uint64, _ time.Duration, err error) {
}

func (t NoopTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
}

func (t NoopTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, _ *vm.ScopeContext, depth int, err error) {
}

func (t NoopTracer) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

func (t NoopTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
}

func (t NoopTracer) GetResult() (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

func (t NoopTracer) Stop(err error) {
}

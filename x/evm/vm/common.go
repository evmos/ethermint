package vm

import (
	"github.com/ethereum/go-ethereum/core/vm"
)

// ScopeContext contains the things that are per-call, such as stack and memory,
// but not transients like pc and gas
type ScopeContext struct {
	Memory   *vm.Memory
	Stack    *Stack
	Contract *vm.Contract
}

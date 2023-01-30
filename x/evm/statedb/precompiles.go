package statedb

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// StatefulPrecompiledContract is a stateful precompiled contract in evm
type StatefulPrecompiledContract interface {
	vm.PrecompiledContract
}

type PrecompiledContractCreator func(sdk.Context) StatefulPrecompiledContract

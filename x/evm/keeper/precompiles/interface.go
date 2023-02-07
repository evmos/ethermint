package precompiles

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// PrecompiledContracts defines list of precompiled contract
type PrecompiledContracts []vm.PrecompiledContract

type StatefulPrecompiledContract interface {
	vm.PrecompiledContract
}

// ExtStateDB defines extra methods of statedb to support stateful precompiled contracts
type ExtStateDB interface {
	ExecuteNativeAction(action func(ctx sdk.Context) error) error
}

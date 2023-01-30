package precompiles

import sdk "github.com/cosmos/cosmos-sdk/types"

// ExtStateDB defines extra methods of statedb to support stateful precompiled contracts
type ExtStateDB interface {
	ExecuteNativeAction(action func(ctx sdk.Context) error) error
}

package v2

import (
	"github.com/Entangle-Protocol/entangle-blockchain/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// MigrateStore sets the default AllowUnprotectedTxs parameter.
func MigrateStore(ctx sdk.Context, paramstore *paramtypes.Subspace) error {
	if !paramstore.HasKeyTable() {
		ps := paramstore.WithKeyTable(types.ParamKeyTable())
		paramstore = &ps
	}

	// add RejectUnprotected
	paramstore.Set(ctx, types.ParamStoreKeyAllowUnprotectedTxs, types.DefaultAllowUnprotectedTxs)
	return nil
}

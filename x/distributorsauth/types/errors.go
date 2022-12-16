package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/distributorsauth module sentinel errors
var (
	ErrNoDistributorInStoreForAddress = sdkerrors.Register(ModuleName, 1, "No distributor in store with address")
	ErrWrongDistributorAddress        = sdkerrors.Register(ModuleName, 2, "Not correct distributor address")
	ErrNoAdminInStoreForAddress       = sdkerrors.Register(ModuleName, 3, "No admin in store with address")
	ErrWrongAdminAddress              = sdkerrors.Register(ModuleName, 4, "Not correct admin address")
	ErrSenderIsNotAnAdmin             = sdkerrors.Register(ModuleName, 5, "Sender is not an admin")
	ErrWrongAdminEditOption           = sdkerrors.Register(ModuleName, 6, "Admin have no edit option")
)

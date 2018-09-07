package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Application multi-store keys
var (
	StoreKeyAccount     = sdk.NewKVStoreKey("acc")
	StoreKeyStorage     = sdk.NewKVStoreKey("contract_storage")
	StoreKeyMain        = sdk.NewKVStoreKey("main")
	StoreKeyStake       = sdk.NewKVStoreKey("stake")
	StoreKeySlashing    = sdk.NewKVStoreKey("slashing")
	StoreKeyGov         = sdk.NewKVStoreKey("gov")
	StoreKeyFeeColl     = sdk.NewKVStoreKey("fee")
	StoreKeyParams      = sdk.NewKVStoreKey("params")
	StoreKeyTransParams = sdk.NewTransientStoreKey("transient_params")
)

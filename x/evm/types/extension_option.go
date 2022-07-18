package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// CheckDynamicFeeExtensionOption accepts the `ExtensionOptionDynamicFeeTx` extension option.
func CheckDynamicFeeExtensionOption(any *codectypes.Any) bool {
	_, ok := any.GetCachedValue().(*ExtensionOptionDynamicFeeTx)
	return ok
}

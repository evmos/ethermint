package types

const (
	// ModuleName string name of module
	ModuleName = "feemarket"

	// StoreKey key for base fee and block gas used.
	// The Fee Market module should use a prefix store.
	StoreKey = ModuleName

	// RouterKey uses module name for routing
	RouterKey = ModuleName
)

// prefix bytes for the feemarket persistent store
const (
	prefixBlockGasUsed = iota + 1
)

// KVStore key prefixes
var (
	KeyPrefixBlockGasUsed = []byte{prefixBlockGasUsed}
)

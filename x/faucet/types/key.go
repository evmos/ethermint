package types

const (
	// ModuleName is the name of the module
	ModuleName = "faucet"

	// StoreKey to be used when creating the KVStore
	StoreKey = ModuleName

	// RouterKey uses module name for tx routing
	RouterKey = ModuleName

	// QuerierRoute uses module name for query routing
	QuerierRoute = ModuleName
)

var (
	EnableFaucetKey  = []byte{0x01}
	TimeoutKey       = []byte{0x02}
	CapKey           = []byte{0x03}
	MaxPerRequestKey = []byte{0x04}
	FundedKey        = []byte{0x05}
)

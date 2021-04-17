package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

const (
	// ModuleName string name of module
	ModuleName = "evm"

	// StoreKey key for ethereum storage data, account code (StateDB) or block
	// related data for Web3.
	// The EVM module should use a prefix store.
	StoreKey = ModuleName

	// RouterKey uses module name for routing
	RouterKey = ModuleName
)

// KVStore key prefixes
var (
	KeyPrefixBloom       = []byte{0x01}
	KeyPrefixLogs        = []byte{0x02}
	KeyPrefixCode        = []byte{0x03}
	KeyPrefixStorage     = []byte{0x04}
	KeyPrefixChainConfig = []byte{0x05}
	KeyPrefixHeightHash  = []byte{0x06}
)

// HeightHashKey returns the key for the given chain epoch and height.
// The key will be composed in the following order:
//   key = prefix + bytes(height)
// This ordering facilitates the iteration by height for the EVM GetHashFn
// queries.
func HeightHashKey(height uint64) []byte {
	return sdk.Uint64ToBigEndian(height)
}

// BloomKey defines the store key for a block Bloom
func BloomKey(height int64) []byte {
	return sdk.Uint64ToBigEndian(uint64(height))
}

// AddressStoragePrefix returns a prefix to iterate over a given account storage.
func AddressStoragePrefix(address ethcmn.Address) []byte {
	return append(KeyPrefixStorage, address.Bytes()...)
}

// StateKey defines the full key under which an account state is stored.
func StateKey(address ethcmn.Address, key []byte) []byte {
	return append(AddressStoragePrefix(address), key...)
}

// TODO: fix Logs key and append block hash

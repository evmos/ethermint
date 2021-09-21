package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// ModuleName string name of module
	ModuleName = "evm"

	// StoreKey key for ethereum storage data, account code (StateDB) or block
	// related data for Web3.
	// The EVM module should use a prefix store.
	StoreKey = ModuleName

	// TransientKey is the key to access the EVM transient store, that is reset
	// during the Commit phase.
	TransientKey = "transient_" + ModuleName

	// RouterKey uses module name for routing
	RouterKey = ModuleName
)

// prefix bytes for the EVM persistent store
const (
	prefixCode = iota + 1
	prefixStorage
)

// prefix bytes for the EVM transient store
const (
	prefixTransientSuicided = iota + 1
	prefixTransientBloom
	prefixTransientTxIndex
	prefixTransientRefund
	prefixTransientAccessListAddress
	prefixTransientAccessListSlot
	prefixTransientTxHash
	prefixTransientLogSize
	prefixTransientTxLogs
)

// KVStore key prefixes
var (
	KeyPrefixCode    = []byte{prefixCode}
	KeyPrefixStorage = []byte{prefixStorage}
)

// Transient Store key prefixes
var (
	KeyPrefixTransientSuicided          = []byte{prefixTransientSuicided}
	KeyPrefixTransientBloom             = []byte{prefixTransientBloom}
	KeyPrefixTransientTxIndex           = []byte{prefixTransientTxIndex}
	KeyPrefixTransientRefund            = []byte{prefixTransientRefund}
	KeyPrefixTransientAccessListAddress = []byte{prefixTransientAccessListAddress}
	KeyPrefixTransientAccessListSlot    = []byte{prefixTransientAccessListSlot}
	KeyPrefixTransientTxHash            = []byte{prefixTransientTxHash}
	KeyPrefixTransientLogSize           = []byte{prefixTransientLogSize}
	KeyPrefixTransientTxLogs            = []byte{prefixTransientTxLogs}
)

// AddressStoragePrefix returns a prefix to iterate over a given account storage.
func AddressStoragePrefix(address common.Address) []byte {
	return append(KeyPrefixStorage, address.Bytes()...)
}

// StateKey defines the full key under which an account state is stored.
func StateKey(address common.Address, key []byte) []byte {
	return append(AddressStoragePrefix(address), key...)
}

// KeyAddressStorage returns the key hash to access a given account state. The composite key
// (address + hash) is hashed using Keccak256.
func KeyAddressStorage(address common.Address, hash common.Hash) common.Hash {
	prefix := address.Bytes()
	key := hash.Bytes()

	compositeKey := make([]byte, len(prefix)+len(key))

	copy(compositeKey, prefix)
	copy(compositeKey[len(prefix):], key)

	return crypto.Keccak256Hash(compositeKey)
}

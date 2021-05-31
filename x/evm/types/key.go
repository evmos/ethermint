package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

const (
	// ModuleName string name of module
	ModuleName = "evm"

	// StoreKey key for ethereum storage data, account code (StateDB) or block
	// related data for Web3.
	// The EVM module should use a prefix store.
	StoreKey = ModuleName

	// Transient Key is the key to access the EVM transient store, that is reset
	// during the Commit phase.
	TransientKey = "transient_" + ModuleName

	// RouterKey uses module name for routing
	RouterKey = ModuleName
)

const (
	prefixHeightToHeaderHash = iota + 1
	prefixBloom
	prefixLogs
	prefixCode
	prefixStorage
	prefixChainConfig
	prefixHashTxReceipt
	prefixBlockHeightTxs
)

const (
	prefixTransientSuicided = iota + 1
	prefixTransientBloom
	prefixTransientTxIndex
	prefixTransientRefund
)

// KVStore key prefixes
var (
	KeyPrefixHeightToHeaderHash = []byte{prefixHeightToHeaderHash}
	KeyPrefixBloom              = []byte{prefixBloom}
	KeyPrefixLogs               = []byte{prefixLogs}
	KeyPrefixCode               = []byte{prefixCode}
	KeyPrefixStorage            = []byte{prefixStorage}
	KeyPrefixChainConfig        = []byte{prefixChainConfig}
	KeyPrefixHashTxReceipt      = []byte{prefixHashTxReceipt}
	KeyPrefixBlockHeightTxs     = []byte{prefixBlockHeightTxs}
)

var (
	KeyPrefixTransientSuicided = []byte{prefixTransientSuicided}
	KeyPrefixTransientBloom    = []byte{prefixTransientBloom}
	KeyPrefixTransientTxIndex  = []byte{prefixTransientTxIndex}
	KeyPrefixTransientRefund   = []byte{prefixTransientRefund}
)

// BloomKey defines the store key for a block Bloom
func BloomKey(height int64) []byte {
	heightBytes := sdk.Uint64ToBigEndian(uint64(height))
	return append(KeyPrefixBloom, heightBytes...)
}

// AddressStoragePrefix returns a prefix to iterate over a given account storage.
func AddressStoragePrefix(address ethcmn.Address) []byte {
	return append(KeyPrefixStorage, address.Bytes()...)
}

// StateKey defines the full key under which an account state is stored.
func StateKey(address ethcmn.Address, key []byte) []byte {
	return append(AddressStoragePrefix(address), key...)
}

// KeyHashTxReceipt returns a key for accessing tx receipt data by hash.
func KeyHashTxReceipt(hash ethcmn.Hash) []byte {
	return append(KeyPrefixHashTxReceipt, hash.Bytes()...)
}

// KeyBlockHeightTxs returns a key for accessing tx hash list by block height.
func KeyBlockHeightTxs(height uint64) []byte {
	heightBytes := sdk.Uint64ToBigEndian(height)
	return append(KeyPrefixBlockHeightTxs, heightBytes...)
}

// KeyAddressStorage returns the key hash to access a given account state. The composite key
// (address + hash) is hashed using Keccak256.
func KeyAddressStorage(address ethcmn.Address, hash ethcmn.Hash) ethcmn.Hash {
	prefix := address.Bytes()
	key := hash.Bytes()

	compositeKey := make([]byte, len(prefix)+len(key))

	copy(compositeKey, prefix)
	copy(compositeKey[len(prefix):], key)

	return ethcrypto.Keccak256Hash(compositeKey)
}

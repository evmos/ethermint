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
	KeyPrefixBlockHash       = []byte{0x01}
	KeyPrefixBloom           = []byte{0x02}
	KeyPrefixLogs            = []byte{0x03}
	KeyPrefixCode            = []byte{0x04}
	KeyPrefixStorage         = []byte{0x05}
	KeyPrefixChainConfig     = []byte{0x06}
	KeyPrefixBlockHeightHash = []byte{0x07}
	KeyPrefixHashTxReceipt   = []byte{0x08}
	KeyPrefixBlockHeightTxs  = []byte{0x09}
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

// KeyBlockHash returns a key for accessing block hash data.
func KeyBlockHash(hash ethcmn.Hash) []byte {
	return append(KeyPrefixBlockHash, hash.Bytes()...)
}

// KeyBlockHash returns a key for accessing block hash data.
func KeyBlockHeightHash(height uint64) []byte {
	heightBytes := sdk.Uint64ToBigEndian(height)
	return append(KeyPrefixBlockHeightHash, heightBytes...)
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

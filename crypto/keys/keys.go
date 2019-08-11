package keys

import (
	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
)

// SigningAlgo defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type SigningAlgo string

const (
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = cosmosKeys.SigningAlgo("emintsecp256k1")
)

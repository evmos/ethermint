package web3

import (
	"github.com/tharsis/ethermint/version"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// PublicApi is the web3_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicApi struct{}

// NewPublicApi creates an instance of the Web3 API.
func NewPublicApi() *PublicApi {
	return &PublicApi{}
}

// ClientVersion returns the client version in the Web3 user agent format.
func (a *PublicApi) ClientVersion() string {
	return version.Version()
}

// Sha3 returns the keccak-256 hash of the passed-in input.
func (a *PublicApi) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}

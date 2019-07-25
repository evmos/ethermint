package rpc

import (
	"github.com/cosmos/ethermint/version"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// PublicWeb3API is the web3_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicWeb3API struct{}

// NewPublicWeb3API creates an instance of the Web3 API.
func NewPublicWeb3API() *PublicWeb3API {
	return &PublicWeb3API{}
}

// ClientVersion returns the client version in the Web3 user agent format.
func (a *PublicWeb3API) ClientVersion() string {
	return version.ClientVersion()
}

// Sha3 returns the keccak-256 hash of the passed-in input.
func (a *PublicWeb3API) Sha3(input hexutil.Bytes) hexutil.Bytes {
	return crypto.Keccak256(input)
}

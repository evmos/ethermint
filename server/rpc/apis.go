// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/cosmos/ethermint/version"
)

// returns the master list of public APIs for use with StartHTTPEndpoint
func GetRPCAPIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: "web3",
			Version:   "1.0",
			Service:   NewPublicWeb3API(),
		},
	}
}

// PublicWeb3API is the web3_ prefixed set of APIs in the WEB3 JSON-RPC spec.
type PublicWeb3API struct {
}

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

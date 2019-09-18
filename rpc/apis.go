// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	emintcrypto "github.com/cosmos/ethermint/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(cliCtx context.CLIContext, key emintcrypto.PrivKeySecp256k1) []rpc.API {
	return []rpc.API{
		{
			Namespace: "web3",
			Version:   "1.0",
			Service:   NewPublicWeb3API(),
			Public:    true,
		},
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthAPI(cliCtx, key),
			Public:    true,
		},
		{
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPersonalEthAPI(cliCtx),
			Public:    false,
		},
	}
}

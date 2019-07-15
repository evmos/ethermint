// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/ethereum/go-ethereum/rpc"
)

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(cliCtx context.CLIContext) []rpc.API {
	return []rpc.API{
		{
			Namespace: "web3",
			Version:   "1.0",
			Service:   NewPublicWeb3API(),
		},
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthAPI(cliCtx),
		},
	}
}

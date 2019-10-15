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
	nonceLock := new(AddrLocker)
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
			Service:   NewPublicEthAPI(cliCtx, nonceLock, key),
			Public:    true,
		},
		{
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPersonalEthAPI(cliCtx, nonceLock),
			Public:    false,
		},
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicFilterAPI(cliCtx),
			Public:    true,
		},
		{
			Namespace: "net",
			Version:   "1.0",
			Service:   NewPublicNetAPI(cliCtx),
			Public:    true,
		},
	}
}

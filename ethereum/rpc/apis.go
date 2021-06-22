// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tharsis/ethermint/ethereum/rpc/types"

	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// RPC namespaces and API version
const (
	Web3Namespace     = "web3"
	EthNamespace      = "eth"
	PersonalNamespace = "personal"
	NetNamespace      = "net"
	TxPoolNamespace   = "txpool"

	apiVersion = "1.0"
)

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(clientCtx client.Context, tmWSClient *rpcclient.WSClient) []rpc.API {
	nonceLock := new(types.AddrLocker)
	backend := NewEVMBackend(clientCtx)
	ethAPI := NewPublicEthAPI(clientCtx, backend, nonceLock)

	return []rpc.API{
		{
			Namespace: Web3Namespace,
			Version:   apiVersion,
			Service:   NewPublicWeb3API(),
			Public:    true,
		},
		{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   ethAPI,
			Public:    true,
		},
		{
			Namespace: EthNamespace,
			Version:   apiVersion,
			Service:   NewPublicFilterAPI(tmWSClient, backend),
			Public:    true,
		},
		{
			Namespace: NetNamespace,
			Version:   apiVersion,
			Service:   NewPublicNetAPI(clientCtx),
			Public:    true,
		},
		{
			Namespace: PersonalNamespace,
			Version:   apiVersion,
			Service:   NewPersonalAPI(ethAPI),
			Public:    true,
		},
		{
			Namespace: TxPoolNamespace,
			Version:   apiVersion,
			Service:   NewPublicTxPoolAPI(),
			Public:    true,
		},
	}
}

// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth/filters"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/net"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/personal"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/txpool"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/web3"
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
	backend := backend.NewEVMBackend(clientCtx)
	ethAPI := eth.NewPublicAPI(clientCtx, backend, nonceLock)

	return []rpc.API{
		{
			Namespace: Web3Namespace,
			Version:   apiVersion,
			Service:   web3.NewPublicAPI(),
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
			Service:   filters.NewPublicAPI(tmWSClient, backend),
			Public:    true,
		},
		{
			Namespace: NetNamespace,
			Version:   apiVersion,
			Service:   net.NewPublicAPI(clientCtx),
			Public:    true,
		},
		{
			Namespace: PersonalNamespace,
			Version:   apiVersion,
			Service:   personal.NewAPI(ethAPI),
			Public:    true,
		},
		{
			Namespace: TxPoolNamespace,
			Version:   apiVersion,
			Service:   txpool.NewPublicAPI(),
			Public:    true,
		},
	}
}

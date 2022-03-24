// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tharsis/ethermint/rpc/ethereum/backend"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/debug"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/eth"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/eth/filters"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/miner"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/net"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/personal"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/txpool"
	"github.com/tharsis/ethermint/rpc/ethereum/namespaces/web3"
	"github.com/tharsis/ethermint/rpc/ethereum/types"

	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// RPC namespaces and API version
const (
	Web3Namespace     = "web3"
	EthNamespace      = "eth"
	PersonalNamespace = "personal"
	NetNamespace      = "net"
	TxPoolNamespace   = "txpool"
	DebugNamespace    = "debug"
	MinerNamespace    = "miner"

	apiVersion = "1.0"
)

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(ctx *server.Context, clientCtx client.Context, tmWSClient *rpcclient.WSClient, selectedAPIs []string) []rpc.API {
	nonceLock := new(types.AddrLocker)
	evmBackend := backend.NewEVMBackend(ctx, ctx.Logger, clientCtx)

	var apis []rpc.API
	// remove duplicates
	selectedAPIs = unique(selectedAPIs)

	for index := range selectedAPIs {
		switch selectedAPIs[index] {
		case EthNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   eth.NewPublicAPI(ctx.Logger, clientCtx, evmBackend, nonceLock),
					Public:    true,
				},
				rpc.API{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   filters.NewPublicAPI(ctx.Logger, clientCtx, tmWSClient, evmBackend),
					Public:    true,
				},
			)
		case Web3Namespace:
			apis = append(apis,
				rpc.API{
					Namespace: Web3Namespace,
					Version:   apiVersion,
					Service:   web3.NewPublicAPI(),
					Public:    true,
				},
			)
		case NetNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: NetNamespace,
					Version:   apiVersion,
					Service:   net.NewPublicAPI(clientCtx),
					Public:    true,
				},
			)
		case PersonalNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: PersonalNamespace,
					Version:   apiVersion,
					Service:   personal.NewAPI(ctx.Logger, clientCtx, evmBackend),
					Public:    false,
				},
			)
		case TxPoolNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: TxPoolNamespace,
					Version:   apiVersion,
					Service:   txpool.NewPublicAPI(ctx.Logger),
					Public:    true,
				},
			)
		case DebugNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: DebugNamespace,
					Version:   apiVersion,
					Service:   debug.NewAPI(ctx, evmBackend, clientCtx),
					Public:    true,
				},
			)
		case MinerNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: MinerNamespace,
					Version:   apiVersion,
					Service:   miner.NewPrivateAPI(ctx, clientCtx, evmBackend),
					Public:    false,
				},
			)
		default:
			ctx.Logger.Error("invalid namespace value", "namespace", selectedAPIs[index])
		}
	}

	return apis
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

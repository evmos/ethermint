// Package rpc contains RPC handler methods and utilities to start
// Ethermint's Web3-compatibly JSON-RPC server.
package rpc

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"

	"github.com/evmos/ethermint/rpc/backend"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/debug"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/eth"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/eth/filters"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/miner"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/net"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/personal"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/txpool"
	"github.com/evmos/ethermint/rpc/namespaces/ethereum/web3"
	ethermint "github.com/evmos/ethermint/types"

	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// RPC namespaces and API version
const (
	// Cosmos namespaces

	CosmosNamespace = "cosmos"

	// Ethereum namespaces

	Web3Namespace     = "web3"
	EthNamespace      = "eth"
	PersonalNamespace = "personal"
	NetNamespace      = "net"
	TxPoolNamespace   = "txpool"
	DebugNamespace    = "debug"
	MinerNamespace    = "miner"

	apiVersion = "1.0"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace     string      // namespace under which the rpc methods of Service are exposed
	Version       string      // deprecated - this field is no longer used, but retained for compatibility
	Service       interface{} // receiver instance which holds the methods
	Public        bool        // deprecated - this field is no longer used, but retained for compatibility
	Authenticated bool        // whether the api should only be available behind authentication.
}

// APICreator creates the JSON-RPC API implementations.
type APICreator = func(
	ctx *server.Context,
	clientCtx client.Context,
	tendermintWebsocketClient *rpcclient.WSClient,
	allowUnprotectedTxs bool,
	indexer ethermint.EVMTxIndexer,
) []API

// apiCreators defines the JSON-RPC API namespaces.
var apiCreators map[string]APICreator

func init() {
	apiCreators = map[string]APICreator{
		EthNamespace: func(ctx *server.Context,
			clientCtx client.Context,
			tmWSClient *rpcclient.WSClient,
			allowUnprotectedTxs bool,
			indexer ethermint.EVMTxIndexer,
		) []API {
			evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
			return []API{
				{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   eth.NewPublicAPI(ctx.Logger, evmBackend),
					Public:    true,
				},
				{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   filters.NewPublicAPI(ctx.Logger, clientCtx, tmWSClient, evmBackend),
					Public:    true,
				},
			}
		},
		Web3Namespace: func(*server.Context, client.Context, *rpcclient.WSClient, bool, ethermint.EVMTxIndexer) []API {
			return []API{
				{
					Namespace: Web3Namespace,
					Version:   apiVersion,
					Service:   web3.NewPublicAPI(),
					Public:    true,
				},
			}
		},
		NetNamespace: func(_ *server.Context, clientCtx client.Context, _ *rpcclient.WSClient, _ bool, _ ethermint.EVMTxIndexer) []API {
			return []API{
				{
					Namespace: NetNamespace,
					Version:   apiVersion,
					Service:   net.NewPublicAPI(clientCtx),
					Public:    true,
				},
			}
		},
		PersonalNamespace: func(ctx *server.Context,
			clientCtx client.Context,
			_ *rpcclient.WSClient,
			allowUnprotectedTxs bool,
			indexer ethermint.EVMTxIndexer,
		) []API {
			evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
			return []API{
				{
					Namespace: PersonalNamespace,
					Version:   apiVersion,
					Service:   personal.NewAPI(ctx.Logger, evmBackend),
					Public:    false,
				},
			}
		},
		TxPoolNamespace: func(ctx *server.Context, _ client.Context, _ *rpcclient.WSClient, _ bool, _ ethermint.EVMTxIndexer) []API {
			return []API{
				{
					Namespace: TxPoolNamespace,
					Version:   apiVersion,
					Service:   txpool.NewPublicAPI(ctx.Logger),
					Public:    true,
				},
			}
		},
		DebugNamespace: func(ctx *server.Context,
			clientCtx client.Context,
			_ *rpcclient.WSClient,
			allowUnprotectedTxs bool,
			indexer ethermint.EVMTxIndexer,
		) []API {
			evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
			return []API{
				{
					Namespace: DebugNamespace,
					Version:   apiVersion,
					Service:   debug.NewAPI(ctx, evmBackend),
					Public:    true,
				},
			}
		},
		MinerNamespace: func(ctx *server.Context,
			clientCtx client.Context,
			_ *rpcclient.WSClient,
			allowUnprotectedTxs bool,
			indexer ethermint.EVMTxIndexer,
		) []API {
			evmBackend := backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs, indexer)
			return []API{
				{
					Namespace: MinerNamespace,
					Version:   apiVersion,
					Service:   miner.NewPrivateAPI(ctx, evmBackend),
					Public:    false,
				},
			}
		},
	}
}

// GetRPCAPIs returns the list of all APIs
func GetRPCAPIs(ctx *server.Context,
	clientCtx client.Context,
	tmWSClient *rpcclient.WSClient,
	allowUnprotectedTxs bool,
	indexer ethermint.EVMTxIndexer,
	selectedAPIs []string,
) []API {
	var apis []API

	for _, ns := range selectedAPIs {
		if creator, ok := apiCreators[ns]; ok {
			apis = append(apis, creator(ctx, clientCtx, tmWSClient, allowUnprotectedTxs, indexer)...)
		} else {
			ctx.Logger.Error("invalid namespace value", "namespace", ns)
		}
	}

	return apis
}

// RegisterAPINamespace registers a new API namespace with the API creator.
// This function fails if the namespace is already registered.
func RegisterAPINamespace(ns string, creator APICreator) error {
	if _, ok := apiCreators[ns]; ok {
		return fmt.Errorf("duplicated api namespace %s", ns)
	}
	apiCreators[ns] = creator
	return nil
}

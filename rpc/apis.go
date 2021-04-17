package rpc

import (
	"github.com/cosmos/ethermint/rpc/namespaces/eth/filters"
	"github.com/cosmos/ethermint/rpc/namespaces/net"
	"github.com/cosmos/ethermint/rpc/namespaces/personal"
	"github.com/cosmos/ethermint/rpc/namespaces/web3"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"

	"github.com/cosmos/ethermint/rpc/backend"
	"github.com/cosmos/ethermint/rpc/namespaces/eth"
	rpctypes "github.com/cosmos/ethermint/rpc/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// RPC namespaces and API version
const (
	Web3Namespace     = "web3"
	EthNamespace      = "eth"
	PersonalNamespace = "personal"
	NetNamespace      = "net"
	flagRPCAPI        = "rpc-api"

	apiVersion = "1.0"
	flagRPCAPI = "rpc-api"
)

// GetAPIs returns the list of all APIs from the Ethereum namespaces
func GetAPIs(clientCtx client.Context, selectedApis []string, keys ...ethsecp256k1.PrivKey) []rpc.API {
	nonceLock := new(rpctypes.AddrLocker)
	backend := backend.New(clientCtx)
	ethAPI := eth.NewAPI(clientCtx, backend, nonceLock, keys...)

	var apis []rpc.API

	for _, api := range selectedApis {
		switch api {
		case Web3Namespace:
			apis = append(apis,
				rpc.API{
					Namespace: Web3Namespace,
					Version:   apiVersion,
					Service:   web3.NewAPI(),
					Public:    true,
				},
			)
		case EthNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   ethAPI,
					Public:    true,
				},
				rpc.API{
					Namespace: EthNamespace,
					Version:   apiVersion,
					Service:   filters.NewAPI(clientCtx, backend),
					Public:    true,
				},
			)
		case PersonalNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: PersonalNamespace,
					Version:   apiVersion,
					Service:   personal.NewAPI(ethAPI),
					Public:    false,
				},
			)
		case NetNamespace:
			apis = append(apis,
				rpc.API{
					Namespace: NetNamespace,
					Version:   apiVersion,
					Service:   net.NewAPI(clientCtx),
					Public:    true,
				},
			)
		}
	}

	return apis
}

package rpc

import (
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/ethereum/go-ethereum/rpc"
)

// RegisterEthereum creates a new ethereum JSON-RPC server and recreates a CLI command to start Cosmos REST server with web3 RPC API and
// Cosmos rest-server endpoints
func RegisterEthereum(clientCtx client.Context, r *mux.Router) {
	server := rpc.NewServer()
	r.HandleFunc("/", server.ServeHTTP).Methods("POST", "OPTIONS")

	rpcapi := viper.GetString(flagRPCAPI)
	rpcapi = strings.ReplaceAll(rpcapi, " ", "")
	rpcapiArr := strings.Split(rpcapi, ",")

	apis := GetAPIs(clientCtx, rpcapiArr)

	// Register all the APIs exposed by the namespace services
	// TODO: handle allowlist and private APIs
	for _, api := range apis {
		if err := server.RegisterName(api.Namespace, api.Service); err != nil {
			panic(err)
		}
	}
}

package rpc

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	)

// StartHTTPEndpoint starts the Tendermint Web3-compatible RPC layer. Consumes a Context for cancellation, a config
// struct, and a list of rpc.API interfaces that will be automatically wired into a JSON-RPC webserver.
func StartHTTPEndpoint(ctx context.Context, config *Config, apis []rpc.API) (*rpc.Server, error) {
	uniqModules := make(map[string]string)
	for _, api := range apis {
		uniqModules[api.Namespace] = api.Namespace
	}
	modules := make([]string, len(uniqModules))
	i := 0
	for k := range uniqModules {
		modules[i] = k
		i++
	}

	endpoint := fmt.Sprintf("%s:%d", config.RPCAddr, config.RPCPort)
	_, server, err := rpc.StartHTTPEndpoint(endpoint, apis, modules, config.RPCCORSDomains, config.RPCVHosts)

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down server.")
		server.Stop()
	}()

	return server, err
}

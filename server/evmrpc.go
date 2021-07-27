package server

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/tharsis/ethermint/cmd/ethermintd/config"
	"github.com/tharsis/ethermint/ethereum/rpc"
)

// StartEVMRPC start evm rpc server
func StartEVMRPC(ctx *server.Context, clientCtx client.Context, tmRPCAddr string, tmEndpoint string, config config.Config) (*http.Server, chan struct{}, error) {
	tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint)

	rpcServer := ethrpc.NewServer()

	rpcAPIArr := config.EVMRPC.API
	apis := rpc.GetRPCAPIs(ctx, clientCtx, tmWsClient, rpcAPIArr)

	for _, api := range apis {
		if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			ctx.Logger.Error(
				"failed to register service in EVM RPC namespace",
				"namespace", api.Namespace,
				"service", api.Service,
			)
			return nil, nil, err
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/", rpcServer.ServeHTTP).Methods("POST")

	handlerWithCors := cors.Default()
	if config.EVMRPC.EnableUnsafeCORS {
		handlerWithCors = cors.AllowAll()
	}

	httpSrv := &http.Server{
		Addr:    config.EVMRPC.RPCAddress,
		Handler: handlerWithCors.Handler(r),
	}
	httpSrvDone := make(chan struct{}, 1)

	errCh := make(chan error)
	go func() {
		ctx.Logger.Info("Starting EVM RPC server", "address", config.EVMRPC.RPCAddress)
		if err := httpSrv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			ctx.Logger.Error("failed to start EVM RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		ctx.Logger.Error("failed to boot EVM RPC server", "error", err.Error())
		return nil, nil, err
	case <-time.After(types.ServerStartTime): // assume EVM RPC server started successfully
	}

	ctx.Logger.Info("Starting EVM WebSocket server", "address", config.EVMRPC.WsAddress)
	_, port, _ := net.SplitHostPort(config.EVMRPC.RPCAddress)

	// allocate separate WS connection to Tendermint
	tmWsClient = ConnectTmWS(tmRPCAddr, tmEndpoint)
	wsSrv := rpc.NewWebsocketsServer(ctx.Logger, tmWsClient, "localhost:"+port, config.EVMRPC.WsAddress)
	wsSrv.Start()
	return httpSrv, httpSrvDone, nil
}

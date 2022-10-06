package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/metrics/prometheus"
	"github.com/evmos/ethermint/rpc"
	rpcServer "github.com/evmos/ethermint/rpc/server"

	"github.com/evmos/ethermint/server/config"
	ethermint "github.com/evmos/ethermint/types"
)

// StartJSONRPC starts the JSON-RPC server
func StartJSONRPC(ctx *server.Context,
	clientCtx client.Context,
	tmRPCAddr,
	tmEndpoint string,
	config *config.Config,
	indexer ethermint.EVMTxIndexer,
) (*http.Server, chan struct{}, error) {
	tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint, ctx.Logger)

	logger := ctx.Logger.With("module", "geth")
	ethlog.Root().SetHandler(ethlog.FuncHandler(func(r *ethlog.Record) error {
		switch r.Lvl {
		case ethlog.LvlTrace, ethlog.LvlDebug:
			logger.Debug(r.Msg, r.Ctx...)
		case ethlog.LvlInfo, ethlog.LvlWarn:
			logger.Info(r.Msg, r.Ctx...)
		case ethlog.LvlError, ethlog.LvlCrit:
			logger.Error(r.Msg, r.Ctx...)
		}
		return nil
	}))

	rpcServer := rpcServer.NewServer()

	if config.JSONRPC.EnableMetrics {
		registry := metrics.NewRegistry()
		rpcServer.WithMetrics(registry)
		metricsErrChan := setupMetricsServer(ctx, config, registry)

		select {
		case err := <-metricsErrChan:
			ctx.Logger.Error("failed to boot JSON-RPC metrics server", "error", err.Error())
			return nil, nil, err
		case <-time.After(types.ServerStartTime): // assume JSON RPC server started successfully
		}
	}

	allowUnprotectedTxs := config.JSONRPC.AllowUnprotectedTxs
	rpcAPIArr := config.JSONRPC.API

	apis := rpc.GetRPCAPIs(ctx, clientCtx, tmWsClient, allowUnprotectedTxs, indexer, rpcAPIArr)

	for _, api := range apis {
		if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			ctx.Logger.Error(
				"failed to register service in JSON RPC namespace",
				"namespace", api.Namespace,
				"service", api.Service,
			)
			return nil, nil, err
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/", rpcServer.ServeHTTP).Methods("POST")

	handlerWithCors := cors.Default()
	if config.API.EnableUnsafeCORS {
		handlerWithCors = cors.AllowAll()
	}

	httpSrv := &http.Server{
		Addr:              config.JSONRPC.Address,
		Handler:           handlerWithCors.Handler(r),
		ReadHeaderTimeout: config.JSONRPC.HTTPTimeout,
		ReadTimeout:       config.JSONRPC.HTTPTimeout,
		WriteTimeout:      config.JSONRPC.HTTPTimeout,
		IdleTimeout:       config.JSONRPC.HTTPIdleTimeout,
	}
	httpSrvDone := make(chan struct{}, 1)

	ln, err := Listen(httpSrv.Addr, config)
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error)
	go func() {
		ctx.Logger.Info("Starting JSON-RPC server", "address", config.JSONRPC.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			ctx.Logger.Error("failed to start JSON-RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		ctx.Logger.Error("failed to boot JSON-RPC server", "error", err.Error())
		return nil, nil, err
	case <-time.After(types.ServerStartTime): // assume JSON RPC server started successfully
	}

	ctx.Logger.Info("Starting JSON WebSocket server", "address", config.JSONRPC.WsAddress)

	// allocate separate WS connection to Tendermint
	tmWsClient = ConnectTmWS(tmRPCAddr, tmEndpoint, ctx.Logger)
	wsSrv := rpc.NewWebsocketsServer(clientCtx, ctx.Logger, tmWsClient, config)
	wsSrv.Start()

	return httpSrv, httpSrvDone, nil
}

func setupMetricsServer(ctx *server.Context, config *config.Config, registry metrics.Registry) chan error {
	address := config.JSONRPC.MetricsAddress

	m := http.NewServeMux()
	m.Handle("/metrics", prometheus.Handler(registry))

	metricsErrCh := make(chan error)

	go func() {
		ctx.Logger.Info("Starting JSON-RPC metrics server", "address", address)

		server := &http.Server{
			Addr:              address,
			Handler:           m,
			ReadHeaderTimeout: config.JSONRPC.HTTPTimeout,
			ReadTimeout:       config.JSONRPC.HTTPTimeout,
			WriteTimeout:      config.JSONRPC.HTTPTimeout,
			IdleTimeout:       config.JSONRPC.HTTPIdleTimeout,
		}

		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				return
			}

			ctx.Logger.Error("failed to start JSON-RPC metrics server", "error", err.Error())
			metricsErrCh <- err
		}
	}()

	return metricsErrCh
}

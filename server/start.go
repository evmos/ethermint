package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"time"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/xlab/closer"
	log "github.com/xlab/suplog"
	"google.golang.org/grpc"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/client/local"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	"github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	ethlog "github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/tharsis/ethermint/cmd/ethermintd/config"
	"github.com/tharsis/ethermint/ethereum/rpc"
)

// Tendermint full-node start flags
const (
	flagAddress            = "address"
	flagTransport          = "transport"
	flagTraceStore         = "trace-store"
	flagCPUProfile         = "cpu-profile"
	FlagMinGasPrices       = "minimum-gas-prices"
	FlagHaltHeight         = "halt-height"
	FlagHaltTime           = "halt-time"
	FlagInterBlockCache    = "inter-block-cache"
	FlagUnsafeSkipUpgrades = "unsafe-skip-upgrades"
	FlagTrace              = "trace"
	FlagInvCheckPeriod     = "inv-check-period"

	FlagPruning           = "pruning"
	FlagPruningKeepRecent = "pruning-keep-recent"
	FlagPruningKeepEvery  = "pruning-keep-every"
	FlagPruningInterval   = "pruning-interval"
	FlagMinRetainBlocks   = "min-retain-blocks"
)

// GRPC-related flags.
const (
	flagGRPCEnable    = "grpc.enable"
	flagGRPCAddress   = "grpc.address"
	flagEVMRPCEnable  = "evm-rpc.enable"
	flagEVMRPCAddress = "evm-rpc.address"
	flagEVMWSAddress  = "evm-rpc.ws-address"
)

// State sync-related flags.
const (
	FlagStateSyncSnapshotInterval   = "state-sync.snapshot-interval"
	FlagStateSyncSnapshotKeepRecent = "state-sync.snapshot-keep-recent"
)

// StartCmd runs the service passed in, either stand-alone or in-process with
// Tendermint.
func StartCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with Tendermint in or out of process. By
default, the application will run with Tendermint in process.

Pruning options can be provided via the '--pruning' flag or alternatively with '--pruning-keep-recent',
'pruning-keep-every', and 'pruning-interval' together.

For '--pruning' the options are as follows:

default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'. During
the ABCI Commit phase, the node will check if the current block height is greater than or equal to
the halt-height or if the current block time is greater than or equal to the halt-time. If so, the
node will attempt to gracefully shutdown and the block will not be committed. In addition, the node
will not be able to commit subsequent blocks.

For profiling and benchmarking purposes, CPU profiling can be enabled via the '--cpu-profile' flag
which accepts a path for the resulting pprof file.
`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)

			// Bind flags to the Context's Viper so the app construction can set
			// options accordingly.
			err := serverCtx.Viper.BindPFlags(cmd.Flags())
			if err != nil {
				return err
			}

			_, err = server.GetPruningOptionsFromFlags(serverCtx.Viper)
			return err
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx.Logger.Info("starting ABCI with Tendermint")

			// amino is needed here for backwards compatibility of REST routes
			err := startInProcess(serverCtx, clientCtx, appCreator)
			return err
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flagAddress, "tcp://0.0.0.0:26658", "Listen address")
	cmd.Flags().String(flagTransport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(FlagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photino;0.0001stake)")
	cmd.Flags().IntSlice(FlagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")
	cmd.Flags().Uint64(FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Uint64(FlagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain and shutdown the node")
	cmd.Flags().Bool(FlagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to the provided file")
	cmd.Flags().Bool(FlagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().String(FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningKeepEvery, 0, "Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(FlagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint(FlagInvCheckPeriod, 0, "Assert registered invariants every N blocks")
	cmd.Flags().Uint64(FlagMinRetainBlocks, 0, "Minimum block height offset during ABCI commit to prune Tendermint blocks")

	cmd.Flags().Bool(flagGRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(flagGRPCAddress, config.DefaultGRPCAddress, "the gRPC server address to listen on")

	cmd.Flags().Bool(flagEVMRPCEnable, true, "Define if the gRPC server should be enabled")
	cmd.Flags().String(flagEVMRPCAddress, config.DefaultEVMAddress, "the EVM RPC server address to listen on")
	cmd.Flags().String(flagEVMWSAddress, config.DefaultEVMWSAddress, "the EVM WS server address to listen on")

	cmd.Flags().Uint64(FlagStateSyncSnapshotInterval, 0, "State sync snapshot interval")
	cmd.Flags().Uint32(FlagStateSyncSnapshotKeepRecent, 2, "State sync snapshot to keep")

	// add support for all Tendermint-specific command line options
	tcmd.AddNodeFlags(cmd)
	return cmd
}

func startInProcess(ctx *server.Context, clientCtx client.Context, appCreator types.AppCreator) error {
	cfg := ctx.Config
	home := cfg.RootDir

	traceWriterFile := ctx.Viper.GetString(flagTraceStore)
	db, err := openDB(home)
	if err != nil {
		log.WithError(err).Errorln("failed to open DB")
		return err
	}

	traceWriter, err := openTraceWriter(traceWriterFile)
	if err != nil {
		log.WithError(err).Errorln("failed to open trace writer")
		return err
	}

	app := appCreator(ctx.Logger, db, traceWriter, ctx.Viper)

	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		log.WithError(err).Errorln("failed load or gen node key")
		return err
	}

	genDocProvider := node.DefaultGenesisDocProviderFunc(cfg)
	tmNode, err := node.NewNode(
		cfg,
		pvm.LoadOrGenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(cfg.Instrumentation),
		ctx.Logger.With("module", "node"),
	)
	if err != nil {
		log.WithError(err).Errorln("failed init node")
		return err
	}

	if err := tmNode.Start(); err != nil {
		log.WithError(err).Errorln("failed start tendermint server")
		return err
	}

	genDoc, err := genDocProvider()
	if err != nil {
		return err
	}

	clientCtx = clientCtx.
		WithHomeDir(home).
		WithChainID(genDoc.ChainID).
		WithClient(local.New(tmNode))

	var apiSrv *api.Server
	config := config.GetConfig(ctx.Viper)

	var grpcSrv *grpc.Server
	if config.GRPC.Enable {
		grpcSrv, err = servergrpc.StartGRPCServer(clientCtx, app, config.GRPC.Address)
		if err != nil {
			log.WithError(err).Errorln("failed to boot GRPC server")
			return err
		}
	}

	var httpSrv *http.Server
	var httpSrvDone = make(chan struct{}, 1)
	var wsSrv rpc.WebsocketsServer

	ethlog.Root().SetHandler(ethlog.StdoutHandler)
	if config.EVMRPC.Enable {
		tmEndpoint := "/websocket"
		tmRPCAddr := cfg.RPC.ListenAddress
		log.Infoln("EVM RPC Connecting to Tendermint WebSocket at", tmRPCAddr+tmEndpoint)
		tmWsClient := ConnectTmWS(tmRPCAddr, tmEndpoint)

		rpcServer := ethrpc.NewServer()
		apis := rpc.GetRPCAPIs(clientCtx, tmWsClient)

		for _, api := range apis {
			if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
				log.WithFields(log.Fields{
					"namespace": api.Namespace,
					"service":   api.Service,
				}).WithError(err).Fatalln("failed to register service in EVM RPC namespace")
				return err
			}
		}

		r := mux.NewRouter()
		r.HandleFunc("/", rpcServer.ServeHTTP).Methods("POST")
		if grpcSrv != nil {
			grpcWeb := grpcweb.WrapServer(grpcSrv)
			MountGRPCWebServices(r, grpcWeb, grpcweb.ListGRPCResources(grpcSrv))
		}

		handlerWithCors := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:     []string{"*"},
			AllowCredentials:   false,
			OptionsPassthrough: false,
		})

		httpSrv = &http.Server{
			Addr:    config.EVMRPC.RPCAddress,
			Handler: handlerWithCors.Handler(r),
		}

		errCh := make(chan error)
		go func() {
			log.Infoln("Starting EVM RPC server on", config.EVMRPC.RPCAddress)
			if err := httpSrv.ListenAndServe(); err != nil {
				if err == http.ErrServerClosed {
					close(httpSrvDone)
					return
				}

				log.WithError(err).Errorln("failed to start EVM RPC server")
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			log.WithError(err).Errorln("failed to boot EVM RPC server")
			return err
		case <-time.After(1 * time.Second): // assume EVM RPC server started successfully
		}

		log.Infoln("Starting EVM WebSocket server on", config.EVMRPC.WsAddress)
		_, port, _ := net.SplitHostPort(config.EVMRPC.RPCAddress)

		// allocate separate WS connection to Tendermint
		tmWsClient = ConnectTmWS(tmRPCAddr, tmEndpoint)
		wsSrv = rpc.NewWebsocketsServer(tmWsClient, "localhost:"+port, config.EVMRPC.WsAddress)
		go wsSrv.Start()
	}

	sdkcfg := sdkconfig.GetConfig(ctx.Viper)
	sdkcfg.API = config.API
	if sdkcfg.API.Enable {
		apiSrv = api.New(clientCtx, ctx.Logger.With("module", "api-server"))
		app.RegisterAPIRoutes(apiSrv, sdkcfg.API)
		errCh := make(chan error)

		go func() {
			if err := apiSrv.Start(sdkcfg); err != nil {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			log.WithError(err).Errorln("failed to boot API server")
			return err
		case <-time.After(5 * time.Second): // assume server started successfully
		}
	}

	var cpuProfileCleanup func()

	if cpuProfile := ctx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.WithError(err).Errorln("failed to create CP profile")
			return err
		}

		log.WithField("profile", cpuProfile).Infoln("starting CPU profiler")
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}

		cpuProfileCleanup = func() {
			log.WithField("profile", cpuProfile).Infoln("stopping CPU profiler")
			pprof.StopCPUProfile()
			f.Close()
		}
	}

	closer.Bind(func() {
		if tmNode.IsRunning() {
			_ = tmNode.Stop()
		}

		if cpuProfileCleanup != nil {
			cpuProfileCleanup()
		}

		if httpSrv != nil {
			shutdownCtx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelFn()

			if err := httpSrv.Shutdown(shutdownCtx); err != nil {
				log.WithError(err).Warningln("HTTP server shutdown produced a warning")
			} else {
				log.Infoln("HTTP server shut down, waiting 5 sec")
				select {
				case <-time.Tick(5 * time.Second):
				case <-httpSrvDone:
				}
			}
		}

		if grpcSrv != nil {
			grpcSrv.Stop()
		}

		log.Infoln("Bye!")
	})

	closer.Hold()

	return nil
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile == "" {
		return
	}
	return os.OpenFile(
		traceWriterFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
}

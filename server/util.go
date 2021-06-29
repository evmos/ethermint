package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"
	log "github.com/xlab/suplog"

	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"

	"github.com/tharsis/ethermint/app"
)

// add server commands
func AddCommands(rootCmd *cobra.Command, defaultNodeHome string, appCreator types.AppCreator, appExport types.AppExporter, addStartFlags types.ModuleInitFlags) {
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		sdkserver.ShowNodeIDCmd(),
		sdkserver.ShowValidatorCmd(),
		sdkserver.ShowAddressCmd(),
		sdkserver.VersionCmd(),
	)

	startCmd := StartCmd(appCreator, defaultNodeHome)
	addStartFlags(startCmd)

	rootCmd.AddCommand(
		startCmd,
		sdkserver.UnsafeResetAllCmd(),
		tendermintCmd,
		sdkserver.ExportCmd(appExport, app.DefaultNodeHome),
		version.NewVersionCommand(),
	)
}

func ConnectTmWS(tmRPCAddr, tmEndpoint string) *rpcclient.WSClient {
	tmWsClient, err := rpcclient.NewWS(tmRPCAddr, tmEndpoint,
		rpcclient.MaxReconnectAttempts(256),
		rpcclient.ReadWait(120*time.Second),
		rpcclient.WriteWait(120*time.Second),
		rpcclient.PingPeriod(50*time.Second),
		rpcclient.OnReconnect(func() {
			log.WithField("tendermint_rpc", tmRPCAddr+tmEndpoint).
				Debugln("EVM RPC reconnects to Tendermint WS")
		}),
	)

	if err != nil {
		log.WithError(err).Fatalln("Tendermint WS client could not be created for ", tmRPCAddr+tmEndpoint)
	} else if err := tmWsClient.OnStart(); err != nil {
		log.WithError(err).Fatalln("Tendermint WS client could not start for ", tmRPCAddr+tmEndpoint)
	}

	return tmWsClient
}

func MountGRPCWebServices(
	router *mux.Router,
	grpcWeb *grpcweb.WrappedGrpcServer,
	grpcResources []string,
) {
	for _, res := range grpcResources {
		log.Printf("[GRPC Web] HTTP POST mounted on %s", res)

		s := router.Methods("POST").Subrouter()
		s.HandleFunc(res, func(resp http.ResponseWriter, req *http.Request) {
			if grpcWeb.IsGrpcWebSocketRequest(req) {
				grpcWeb.HandleGrpcWebsocketRequest(resp, req)
				return
			}

			if grpcWeb.IsGrpcWebRequest(req) {
				grpcWeb.HandleGrpcWebRequest(resp, req)
				return
			}
		})
	}
}

package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"
	mintlog "github.com/tharsis/ethermint/log"

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
			(*mintlog.EthermintLoggerInstance.TendermintLogger).Debug(fmt.Sprintf("EVM RPC reconnects to Tendermint WS %s", tmRPCAddr+tmEndpoint))
		}),
	)

	if err != nil {
		(*mintlog.EthermintLoggerInstance.TendermintLogger).Error(fmt.Sprintf("Tendermint WS client could not be created for %s, Error %v", tmRPCAddr+tmEndpoint, err))
	} else if err := tmWsClient.OnStart(); err != nil {
		(*mintlog.EthermintLoggerInstance.TendermintLogger).Error(fmt.Sprintf("Tendermint WS client could not start for %s, Error %v", tmRPCAddr+tmEndpoint, err))
	}

	return tmWsClient
}

func MountGRPCWebServices(
	router *mux.Router,
	grpcWeb *grpcweb.WrappedGrpcServer,
	grpcResources []string,
) {
	for _, res := range grpcResources {

		(*mintlog.EthermintLoggerInstance.TendermintLogger).Info(fmt.Sprintf("[GRPC Web] HTTP POST mounted on %s", res))

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

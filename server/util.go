package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"

	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/version"

	tmcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
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
		tmcmd.ResetAllCmd,
		tmcmd.ResetStateCmd,
	)

	startCmd := StartCmd(appCreator, defaultNodeHome)
	addStartFlags(startCmd)

	rootCmd.AddCommand(
		startCmd,
		tendermintCmd,
		sdkserver.ExportCmd(appExport, defaultNodeHome),
		version.NewVersionCommand(),
		sdkserver.NewRollbackCmd(appCreator, defaultNodeHome),
	)
}

func ConnectTmWS(tmRPCAddr, tmEndpoint string, logger tmlog.Logger) *rpcclient.WSClient {
	tmWsClient, err := rpcclient.NewWS(tmRPCAddr, tmEndpoint,
		rpcclient.MaxReconnectAttempts(256),
		rpcclient.ReadWait(120*time.Second),
		rpcclient.WriteWait(120*time.Second),
		rpcclient.PingPeriod(50*time.Second),
		rpcclient.OnReconnect(func() {
			logger.Debug("EVM RPC reconnects to Tendermint WS", "address", tmRPCAddr+tmEndpoint)
		}),
	)

	if err != nil {
		logger.Error(
			"Tendermint WS client could not be created",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	} else if err := tmWsClient.OnStart(); err != nil {
		logger.Error(
			"Tendermint WS client could not start",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	}

	return tmWsClient
}

func MountGRPCWebServices(
	router *mux.Router,
	grpcWeb *grpcweb.WrappedGrpcServer,
	grpcResources []string,
	logger tmlog.Logger,
) {
	for _, res := range grpcResources {

		logger.Info("[GRPC Web] HTTP POST mounted", "resource", res)

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

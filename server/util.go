package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	log "github.com/xlab/suplog"
)

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

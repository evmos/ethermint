package websockets

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/ethereum/go-ethereum/rpc"

	context "github.com/cosmos/cosmos-sdk/client/context"
)

// Server defines a server that handles Ethereum websockets.
type Server struct {
	rpcAddr string // listen address of rest-server
	wsAddr  string // listen address of ws server
	api     *PubSubAPI
	logger  log.Logger
}

// NewServer creates a new websocket server instance.
func NewServer(clientCtx context.CLIContext, wsAddr string) *Server {
	return &Server{
		rpcAddr: viper.GetString("laddr"),
		wsAddr:  wsAddr,
		api:     NewAPI(clientCtx),
		logger:  log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "websocket-server"),
	}
}

// Start runs the websocket server
func (s *Server) Start() {
	ws := mux.NewRouter()
	ws.Handle("/", s)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", s.wsAddr), ws)
		if err != nil {
			s.logger.Error("http error:", err)
		}
	}()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("websocket upgrade failed; error:", err)
		return
	}

	s.readLoop(wsConn)
}

func (s *Server) sendErrResponse(conn *websocket.Conn, msg string) {
	res := &ErrorResponseJSON{
		Jsonrpc: "2.0",
		Error: &ErrorMessageJSON{
			Code:    big.NewInt(-32600),
			Message: msg,
		},
		ID: nil,
	}
	err := conn.WriteJSON(res)
	if err != nil {
		s.logger.Error("websocket failed write message", "error", err)
	}
}

func (s *Server) readLoop(wsConn *websocket.Conn) {
	for {
		_, mb, err := wsConn.ReadMessage()
		if err != nil {
			_ = wsConn.Close()
			s.logger.Error("failed to read message; error", err)
			return
		}

		var msg map[string]interface{}
		err = json.Unmarshal(mb, &msg)
		if err != nil {
			s.sendErrResponse(wsConn, "invalid request")
			continue
		}

		// check if method == eth_subscribe or eth_unsubscribe
		method := msg["method"]
		if method.(string) == "eth_subscribe" {
			params := msg["params"].([]interface{})
			if len(params) == 0 {
				s.sendErrResponse(wsConn, "invalid parameters")
				continue
			}

			id, err := s.api.subscribe(wsConn, params)
			if err != nil {
				s.sendErrResponse(wsConn, err.Error())
				continue
			}

			res := &SubscriptionResponseJSON{
				Jsonrpc: "2.0",
				ID:      1,
				Result:  id,
			}

			err = wsConn.WriteJSON(res)
			if err != nil {
				s.logger.Error("failed to write json response", err)
				continue
			}

			continue
		} else if method.(string) == "eth_unsubscribe" {
			ids, ok := msg["params"].([]interface{})
			if _, idok := ids[0].(string); !ok || !idok {
				s.sendErrResponse(wsConn, "invalid parameters")
				continue
			}

			ok = s.api.unsubscribe(rpc.ID(ids[0].(string)))
			res := &SubscriptionResponseJSON{
				Jsonrpc: "2.0",
				ID:      1,
				Result:  ok,
			}

			err = wsConn.WriteJSON(res)
			if err != nil {
				s.logger.Error("failed to write json response", err)
				continue
			}

			continue
		}

		// otherwise, call the usual rpc server to respond
		err = s.tcpGetAndSendResponse(wsConn, mb)
		if err != nil {
			s.sendErrResponse(wsConn, err.Error())
		}
	}
}

// tcpGetAndSendResponse connects to the rest-server over tcp, posts a JSON-RPC request, and sends the response
// to the client over websockets
func (s *Server) tcpGetAndSendResponse(conn *websocket.Conn, mb []byte) error {
	addr := strings.Split(s.rpcAddr, "tcp://")
	if len(addr) != 2 {
		return fmt.Errorf("invalid laddr %s", s.rpcAddr)
	}

	tcpConn, err := net.Dial("tcp", addr[1])
	if err != nil {
		return fmt.Errorf("cannot connect to %s; %s", s.rpcAddr, err)
	}

	buf := &bytes.Buffer{}
	_, err = buf.Write(mb)
	if err != nil {
		return fmt.Errorf("failed to write message; %s", err)
	}

	req, err := http.NewRequest("POST", s.rpcAddr, buf)
	if err != nil {
		return fmt.Errorf("failed to request; %s", err)
	}

	req.Header.Set("Content-Type", "application/json;")
	err = req.Write(tcpConn)
	if err != nil {
		return fmt.Errorf("failed to write to rest-server; %s", err)
	}

	respBytes, err := ioutil.ReadAll(tcpConn)
	if err != nil {
		return fmt.Errorf("error reading response from rest-server; %s", err)
	}

	respbuf := &bytes.Buffer{}
	respbuf.Write(respBytes)
	resp, err := http.ReadResponse(bufio.NewReader(respbuf), req)
	if err != nil {
		return fmt.Errorf("could not read response; %s", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read body from response; %s", err)
	}

	var wsSend interface{}
	err = json.Unmarshal(body, &wsSend)
	if err != nil {
		return fmt.Errorf("failed to unmarshal rest-server response; %s", err)
	}

	return conn.WriteJSON(wsSend)
}

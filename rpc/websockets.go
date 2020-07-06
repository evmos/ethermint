package rpc

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
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/log"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"

	context "github.com/cosmos/cosmos-sdk/client/context"
)

type SubscriptionResponseJSON struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	ID      float64     `json:"id"`
}

type SubscriptionNotification struct {
	Jsonrpc string              `json:"jsonrpc"`
	Method  string              `json:"method"`
	Params  *SubscriptionResult `json:"params"`
}

type SubscriptionResult struct {
	Subscription rpc.ID      `json:"subscription"`
	Result       interface{} `json:"result"`
}

type ErrorResponseJSON struct {
	Jsonrpc string            `json:"jsonrpc"`
	Error   *ErrorMessageJSON `json:"error"`
	ID      *big.Int          `json:"id"`
}

type ErrorMessageJSON struct {
	Code    *big.Int `json:"code"`
	Message string   `json:"message"`
}

type websocketsServer struct {
	rpcAddr string // listen address of rest-server
	wsAddr  string // listen address of ws server
	api     *pubSubAPI
	logger  log.Logger
}

func newWebsocketsServer(cliCtx context.CLIContext, wsAddr string) *websocketsServer {
	return &websocketsServer{
		rpcAddr: viper.GetString("laddr"),
		wsAddr:  wsAddr,
		api:     newPubSubAPI(cliCtx),
		logger:  log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "websocket-server"),
	}
}

func (s *websocketsServer) start() {
	ws := mux.NewRouter()
	ws.Handle("/", s)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", s.wsAddr), ws)
		if err != nil {
			s.logger.Error("http error:", err)
		}
	}()
}

func (s *websocketsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (s *websocketsServer) sendErrResponse(conn *websocket.Conn, msg string) {
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

func (s *websocketsServer) readLoop(wsConn *websocket.Conn) {
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
func (s *websocketsServer) tcpGetAndSendResponse(conn *websocket.Conn, mb []byte) error {
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

type wsSubscription struct {
	sub          *Subscription
	unsubscribed chan struct{} // closed when unsubscribing
	conn         *websocket.Conn
}

// pubSubAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec
type pubSubAPI struct {
	cliCtx    context.CLIContext
	events    *EventSystem
	filtersMu sync.Mutex
	filters   map[rpc.ID]*wsSubscription
	logger    log.Logger
}

// newPubSubAPI creates an instance of the ethereum PubSub API.
func newPubSubAPI(cliCtx context.CLIContext) *pubSubAPI {
	return &pubSubAPI{
		cliCtx:  cliCtx,
		events:  NewEventSystem(cliCtx.Client),
		filters: make(map[rpc.ID]*wsSubscription),
		logger:  log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "websocket-client"),
	}
}

func (api *pubSubAPI) subscribe(conn *websocket.Conn, params []interface{}) (rpc.ID, error) {
	method, ok := params[0].(string)
	if !ok {
		return "0", fmt.Errorf("invalid parameters")
	}

	switch method {
	case "newHeads":
		// TODO: handle extra params
		return api.subscribeNewHeads(conn)
	case "logs":
		if len(params) > 1 {
			return api.subscribeLogs(conn, params[1])
		}

		return api.subscribeLogs(conn, nil)
	case "newPendingTransactions":
		return api.subscribePendingTransactions(conn)
	case "syncing":
		return api.subscribeSyncing(conn)
	default:
		return "0", fmt.Errorf("unsupported method %s", method)
	}
}

func (api *pubSubAPI) unsubscribe(id rpc.ID) bool {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	if api.filters[id] == nil {
		return false
	}

	close(api.filters[id].unsubscribed)
	delete(api.filters, id)
	return true
}

func (api *pubSubAPI) subscribeNewHeads(conn *websocket.Conn) (rpc.ID, error) {
	sub, _, err := api.events.SubscribeNewHeads()
	if err != nil {
		return "", fmt.Errorf("error creating block filter: %s", err.Error())
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[sub.ID()] = &wsSubscription{
		sub:          sub,
		conn:         conn,
		unsubscribed: unsubscribed,
	}
	api.filtersMu.Unlock()

	go func(headersCh <-chan coretypes.ResultEvent, errCh <-chan error) {
		for {
			select {
			case event := <-headersCh:
				data, _ := event.Data.(tmtypes.EventDataNewBlockHeader)
				header := EthHeaderFromTendermint(data.Header)

				api.filtersMu.Lock()
				if f, found := api.filters[sub.ID()]; found {
					// write to ws conn
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: sub.ID(),
							Result:       header,
						},
					}

					err = f.conn.WriteJSON(res)
					if err != nil {
						api.logger.Error("error writing header")
					}
				}
				api.filtersMu.Unlock()
			case <-errCh:
				api.filtersMu.Lock()
				delete(api.filters, sub.ID())
				api.filtersMu.Unlock()
				return
			case <-unsubscribed:
				return
			}
		}
	}(sub.eventCh, sub.Err())

	return sub.ID(), nil
}

func (api *pubSubAPI) subscribeLogs(conn *websocket.Conn, extra interface{}) (rpc.ID, error) {
	crit := filters.FilterCriteria{}

	if extra != nil {
		params, ok := extra.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("invalid criteria")
		}

		if params["address"] != nil {
			address, ok := params["address"].(string)
			addresses, sok := params["address"].([]interface{})
			if !ok && !sok {
				return "", fmt.Errorf("invalid address; must be address or array of addresses")
			}

			if ok {
				crit.Addresses = []common.Address{common.HexToAddress(address)}
			}

			if sok {
				crit.Addresses = []common.Address{}
				for _, addr := range addresses {
					address, ok := addr.(string)
					if !ok {
						return "", fmt.Errorf("invalid address")
					}

					crit.Addresses = append(crit.Addresses, common.HexToAddress(address))
				}
			}
		}

		if params["topics"] != nil {
			topics, ok := params["topics"].([]interface{})
			if !ok {
				return "", fmt.Errorf("invalid topics")
			}

			crit.Topics = [][]common.Hash{}
			for _, topic := range topics {
				tstr, ok := topic.(string)
				if !ok {
					return "", fmt.Errorf("invalid topics")
				}

				h := common.HexToHash(tstr)
				crit.Topics = append(crit.Topics, []common.Hash{h})
			}
		}
	}

	sub, _, err := api.events.SubscribeLogs(crit)
	if err != nil {
		return rpc.ID(""), err
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[sub.ID()] = &wsSubscription{
		sub:          sub,
		conn:         conn,
		unsubscribed: unsubscribed,
	}
	api.filtersMu.Unlock()

	go func(ch <-chan coretypes.ResultEvent, errCh <-chan error) {
		for {
			select {
			case event := <-ch:
				dataTx, ok := event.Data.(tmtypes.EventDataTx)
				if !ok {
					err = fmt.Errorf("invalid event data %T, expected EventDataTx", event.Data)
					return
				}

				var resultData evmtypes.ResultData
				resultData, err = evmtypes.DecodeResultData(dataTx.TxResult.Result.Data)
				if err != nil {
					return
				}

				logs := filterLogs(resultData.Logs, crit.FromBlock, crit.ToBlock, crit.Addresses, crit.Topics)

				api.filtersMu.Lock()
				if f, found := api.filters[sub.ID()]; found {
					// write to ws conn
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: sub.ID(),
							Result:       logs,
						},
					}

					err = f.conn.WriteJSON(res)
				}
				api.filtersMu.Unlock()

				if err != nil {
					err = fmt.Errorf("failed to write header: %w", err)
					return
				}
			case <-errCh:
				api.filtersMu.Lock()
				delete(api.filters, sub.ID())
				api.filtersMu.Unlock()
				return
			case <-unsubscribed:
				return
			}
		}
	}(sub.eventCh, sub.Err())

	return sub.ID(), nil
}

func (api *pubSubAPI) subscribePendingTransactions(conn *websocket.Conn) (rpc.ID, error) {
	sub, _, err := api.events.SubscribePendingTxs()
	if err != nil {
		return "", fmt.Errorf("error creating block filter: %s", err.Error())
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[sub.ID()] = &wsSubscription{
		sub:          sub,
		conn:         conn,
		unsubscribed: unsubscribed,
	}
	api.filtersMu.Unlock()

	go func(txsCh <-chan coretypes.ResultEvent, errCh <-chan error) {
		for {
			select {
			case ev := <-txsCh:
				data, _ := ev.Data.(tmtypes.EventDataTx)
				txHash := common.BytesToHash(data.Tx.Hash())

				api.filtersMu.Lock()
				if f, found := api.filters[sub.ID()]; found {
					// write to ws conn
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: sub.ID(),
							Result:       txHash,
						},
					}

					err = f.conn.WriteJSON(res)
				}
				api.filtersMu.Unlock()

				if err != nil {
					err = fmt.Errorf("failed to write header: %w", err)
					return
				}
			case <-errCh:
				api.filtersMu.Lock()
				delete(api.filters, sub.ID())
				api.filtersMu.Unlock()
			}
		}
	}(sub.eventCh, sub.Err())

	return sub.ID(), nil
}

func (api *pubSubAPI) subscribeSyncing(conn *websocket.Conn) (rpc.ID, error) {
	return "", nil
}

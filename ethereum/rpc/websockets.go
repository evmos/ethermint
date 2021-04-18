package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"

	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

type WebsocketsServer interface {
	Start()
}

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

func NewWebsocketsServer(tmWSClient *rpcclient.WSClient, rpcAddr, wsAddr string) *websocketsServer {
	return &websocketsServer{
		rpcAddr: rpcAddr,
		wsAddr:  wsAddr,
		api:     newPubSubAPI(tmWSClient),
		logger:  log.WithField("module", "websocket-server"),
	}
}

func (s *websocketsServer) Start() {
	ws := mux.NewRouter()
	ws.Handle("/", s)

	go func() {
		err := http.ListenAndServe(s.wsAddr, ws)
		if err != nil {
			if err == http.ErrServerClosed {
				return
			}

			s.logger.WithError(err).Errorln("failed to start HTTP server for WS")
		}
	}()
}

func (s *websocketsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Warningln("websocket upgrade failed")
		return
	}

	s.readLoop(&wsConn{
		mux:  new(sync.Mutex),
		conn: conn,
	})
}

func (s *websocketsServer) sendErrResponse(wsConn *wsConn, msg string) {
	res := &ErrorResponseJSON{
		Jsonrpc: "2.0",
		Error: &ErrorMessageJSON{
			Code:    big.NewInt(-32600),
			Message: msg,
		},
		ID: nil,
	}

	_ = wsConn.WriteJSON(res)
}

type wsConn struct {
	conn *websocket.Conn
	mux  *sync.Mutex
}

func (w *wsConn) WriteJSON(v interface{}) error {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.conn.WriteJSON(v)
}

func (w *wsConn) Close() error {
	w.mux.Lock()
	defer w.mux.Unlock()

	return w.conn.Close()
}

func (w *wsConn) ReadMessage() (messageType int, p []byte, err error) {
	// not protected by write mutex

	return w.conn.ReadMessage()
}

func (s *websocketsServer) readLoop(wsConn *wsConn) {
	for {
		_, mb, err := wsConn.ReadMessage()
		if err != nil {
			_ = wsConn.Close()
			return
		}

		var msg map[string]interface{}
		err = json.Unmarshal(mb, &msg)
		if err != nil {
			s.sendErrResponse(wsConn, "invalid request")
			continue
		}

		// check if method == eth_subscribe or eth_unsubscribe
		method, ok := msg["method"].(string)
		if !ok {
			// otherwise, call the usual rpc server to respond
			err = s.tcpGetAndSendResponse(wsConn, mb)
			if err != nil {
				s.sendErrResponse(wsConn, err.Error())
			}

			continue
		}

		connId := msg["id"].(float64)
		if method == "eth_subscribe" {
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
				ID:      connId,
				Result:  id,
			}

			err = wsConn.WriteJSON(res)
			if err != nil {
				continue
			}

			continue
		} else if method == "eth_unsubscribe" {
			ids, ok := msg["params"].([]interface{})
			if _, idok := ids[0].(string); !ok || !idok {
				s.sendErrResponse(wsConn, "invalid parameters")
				continue
			}

			ok = s.api.unsubscribe(rpc.ID(ids[0].(string)))
			res := &SubscriptionResponseJSON{
				Jsonrpc: "2.0",
				ID:      connId,
				Result:  ok,
			}

			err = wsConn.WriteJSON(res)
			if err != nil {
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
func (s *websocketsServer) tcpGetAndSendResponse(wsConn *wsConn, mb []byte) error {
	req, err := http.NewRequest("POST", "http://"+s.rpcAddr, bytes.NewBuffer(mb))
	if err != nil {
		return errors.Wrap(err, "Could not build request")
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Could not perform request")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "could not read body from response")
	}

	var wsSend interface{}
	err = json.Unmarshal(body, &wsSend)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal rest-server response")
	}

	return wsConn.WriteJSON(wsSend)
}

type wsSubscription struct {
	sub          *Subscription
	unsubscribed chan struct{} // closed when unsubscribing
	wsConn       *wsConn
	query        string
}

// pubSubAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec
type pubSubAPI struct {
	events    *EventSystem
	filtersMu *sync.RWMutex
	filters   map[rpc.ID]*wsSubscription
	logger    log.Logger
}

// newPubSubAPI creates an instance of the ethereum PubSub API.
func newPubSubAPI(tmWSClient *rpcclient.WSClient) *pubSubAPI {
	return &pubSubAPI{
		events:    NewEventSystem(tmWSClient),
		filtersMu: new(sync.RWMutex),
		filters:   make(map[rpc.ID]*wsSubscription),
		logger:    log.WithField("module", "websocket-client"),
	}
}

func (api *pubSubAPI) subscribe(wsConn *wsConn, params []interface{}) (rpc.ID, error) {
	method, ok := params[0].(string)
	if !ok {
		return "0", errors.New("invalid parameters")
	}

	switch method {
	case "newHeads":
		// TODO: handle extra params
		return api.subscribeNewHeads(wsConn)
	case "logs":
		if len(params) > 1 {
			return api.subscribeLogs(wsConn, params[1])
		}
		return api.subscribeLogs(wsConn, nil)
	case "newPendingTransactions":
		return api.subscribePendingTransactions(wsConn)
	case "syncing":
		return api.subscribeSyncing(wsConn)
	default:
		return "0", errors.Errorf("unsupported method %s", method)
	}
}

func (api *pubSubAPI) unsubscribe(id rpc.ID) bool {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	wsSub, ok := api.filters[id]
	if !ok {
		return false
	}

	wsSub.sub.Unsubscribe(api.events)
	close(api.filters[id].unsubscribed)
	delete(api.filters, id)
	return true
}

func (api *pubSubAPI) getSubscriptionByQuery(query string) *Subscription {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	for _, wsSub := range api.filters {
		if wsSub.query == query {
			return wsSub.sub
		}
	}
	return nil
}

func (api *pubSubAPI) subscribeNewHeads(wsConn *wsConn) (rpc.ID, error) {
	var query = "subscribeNewHeads"
	var subID = rpc.NewID()

	sub, _, err := api.events.SubscribeNewHeads()
	if err != nil {
		return "", errors.Wrap(err, "error creating block filter")
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[subID] = &wsSubscription{
		sub:          sub,
		wsConn:       wsConn,
		unsubscribed: unsubscribed,
		query:        query,
	}
	api.filtersMu.Unlock()

	go func(headersCh <-chan coretypes.ResultEvent, errCh <-chan error) {
		for {
			select {
			case event, ok := <-headersCh:
				if !ok {
					api.unsubscribe(subID)
					return
				}

				data, ok := event.Data.(tmtypes.EventDataNewBlockHeader)
				if !ok {
					log.WithFields(log.Fields{
						"expected": "tmtypes.EventDataNewBlockHeader",
						"actual":   fmt.Sprintf("%T", event.Data),
					}).Warningln("event Data type mismatch")
					continue
				}

				header := EthHeaderFromTendermint(data.Header)

				api.filtersMu.RLock()
				for subID, wsSub := range api.filters {
					if wsSub.query != query {
						continue
					}
					// write to ws conn
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: subID,
							Result:       header,
						},
					}

					err = wsSub.wsConn.WriteJSON(res)
					if err != nil {
						api.logger.WithError(err).Error("error writing header, will drop peer")

						try(func() {
							api.filtersMu.RUnlock()
							api.filtersMu.Lock()
							defer func() {
								api.filtersMu.Unlock()
								api.filtersMu.RLock()
							}()

							if err != websocket.ErrCloseSent {
								_ = wsSub.wsConn.Close()
							}

							delete(api.filters, subID)
							close(wsSub.unsubscribed)
						}, api.logger, "closing websocket peer sub")
					}
				}
				api.filtersMu.RUnlock()
			case err, ok := <-errCh:
				if !ok {
					api.unsubscribe(subID)
					return
				}
				log.WithError(err).Debugln("dropping NewHeads WebSocket subscription", subID)
				api.unsubscribe(subID)
			case <-unsubscribed:
				return
			}
		}
	}(sub.eventCh, sub.Err())

	return subID, nil
}

func try(fn func(), l log.Logger, desc string) {
	if l == nil {
		l = log.DefaultLogger
	}

	defer func() {
		if x := recover(); x != nil {
			if err, ok := x.(error); ok {
				// debug.PrintStack()
				l.WithError(err).Debugln("panic during", desc)
				return
			}

			// debug.PrintStack()
			l.Debugf("panic during %s: %+v", desc, x)
			return
		}
	}()

	fn()
}

func (api *pubSubAPI) subscribeLogs(wsConn *wsConn, extra interface{}) (rpc.ID, error) {
	crit := filters.FilterCriteria{}

	if extra != nil {
		params, ok := extra.(map[string]interface{})
		if !ok {
			err := errors.New("invalid criteria")
			log.WithFields(log.Fields{
				"criteria": fmt.Sprintf("expected: map[string]interface{}, got %T", extra),
			}).Errorln(err)
			return "", err
		}

		if params["address"] != nil {
			address, ok := params["address"].(string)
			addresses, sok := params["address"].([]interface{})
			if !ok && !sok {
				err := errors.New("invalid address; must be address or array of addresses")
				log.WithFields(log.Fields{
					"address":   ok,
					"addresses": sok,
					"type":      fmt.Sprintf("%T", params["address"]),
				}).Errorln(err)
				return "", err
			}

			if ok {
				crit.Addresses = []common.Address{common.HexToAddress(address)}
			}

			if sok {
				crit.Addresses = []common.Address{}
				for _, addr := range addresses {
					address, ok := addr.(string)
					if !ok {
						err := errors.New("invalid address")
						log.WithField("type", fmt.Sprintf("%T", addr)).Errorln(err)
						return "", err
					}

					crit.Addresses = append(crit.Addresses, common.HexToAddress(address))
				}
			}
		}

		if params["topics"] != nil {
			topics, ok := params["topics"].([]interface{})
			if !ok {
				err := errors.New("invalid topics")

				log.WithFields(log.Fields{
					"type": fmt.Sprintf("expected []interface{}, got %T", params["topics"]),
				}).Errorln(err)

				return "", err
			}

			crit.Topics = make([][]common.Hash, len(topics))

			addCritTopic := func(topicIdx int, topic interface{}) error {
				tstr, ok := topic.(string)
				if !ok {
					err := errors.Errorf("invalid topic: %s", topic)
					log.WithField("type", fmt.Sprintf("expected string, got %T", topic)).Errorln(err)
					return err
				}

				crit.Topics[topicIdx] = []common.Hash{common.HexToHash(tstr)}
				return nil
			}

			for topicIdx, subtopics := range topics {
				if subtopics == nil {
					continue
				}

				// in case we don't have list, but a single topic value
				if topic, ok := subtopics.(string); ok {
					if err := addCritTopic(topicIdx, topic); err != nil {
						return "", err
					}

					continue
				}

				// in case we actually have a list of subtopics
				subtopicsList, ok := subtopics.([]interface{})
				if !ok {
					err := errors.New("invalid subtopics")

					log.WithFields(log.Fields{
						"type": fmt.Sprintf("expected []interface{}, got %T", subtopics),
					}).Errorln(err)

					return "", err
				}

				subtopicsCollect := make([]common.Hash, len(subtopicsList))
				for idx, subtopic := range subtopicsList {
					tstr, ok := subtopic.(string)
					if !ok {
						err := errors.Errorf("invalid subtopic: %s", subtopic)
						log.WithField("type", fmt.Sprintf("expected string, got %T", subtopic)).Errorln(err)
						return "", err
					}

					subtopicsCollect[idx] = common.HexToHash(tstr)
				}

				crit.Topics[topicIdx] = subtopicsCollect
			}
		}
	}

	critBz, err := json.Marshal(crit)
	if err != nil {
		log.WithError(err).Errorln("failed to JSON marshal criteria")
		return rpc.ID(""), err
	}

	var query = "subscribeLogs" + string(critBz)
	var subID = rpc.NewID()

	sub, _, err := api.events.SubscribeLogs(crit)
	if err != nil {
		log.WithError(err).Errorln("failed to subscribe logs")
		return rpc.ID(""), err
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[subID] = &wsSubscription{
		sub:          sub,
		wsConn:       wsConn,
		unsubscribed: unsubscribed,
		query:        query,
	}
	api.filtersMu.Unlock()

	go func(ch <-chan coretypes.ResultEvent, errCh <-chan error, subID rpc.ID) {
		for {
			select {
			case event, ok := <-ch:
				if !ok {
					api.unsubscribe(subID)
					return
				}

				dataTx, ok := event.Data.(tmtypes.EventDataTx)
				if !ok {
					log.WithFields(log.Fields{
						"expected": "tmtypes.EventDataTx",
						"actual":   fmt.Sprintf("%T", event.Data),
					}).Warningln("event Data type mismatch")
					continue
				}

				txResponse, err := evmtypes.DecodeTxResponse(dataTx.TxResult.Result.Data)
				if err != nil {
					log.WithError(err).Errorln("failed to decode tx response")
					return
				}

				logs := filterLogs(txResponse.TxLogs.EthLogs(), crit.FromBlock, crit.ToBlock, crit.Addresses, crit.Topics)
				if len(logs) == 0 {
					continue
				}

				api.filtersMu.RLock()
				wsSub, ok := api.filters[subID]
				if !ok {
					log.Warningln("subID not in filters", subID)
					return
				}
				api.filtersMu.RUnlock()

				for _, ethLog := range logs {
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: subID,
							Result:       ethLog,
						},
					}

					err = wsSub.wsConn.WriteJSON(res)
					if err != nil {
						try(func() {
							api.filtersMu.Lock()
							defer api.filtersMu.Unlock()

							if err != websocket.ErrCloseSent {
								_ = wsSub.wsConn.Close()
							}

							delete(api.filters, subID)
							close(wsSub.unsubscribed)
						}, api.logger, "closing websocket peer sub")
					}
				}
			case err, ok := <-errCh:
				if !ok {
					api.unsubscribe(subID)
					return
				}
				log.WithError(err).Warningln("dropping Logs WebSocket subscription", subID)
				api.unsubscribe(subID)
			case <-unsubscribed:
				return
			}
		}
	}(sub.eventCh, sub.Err(), subID)

	return subID, nil
}

func (api *pubSubAPI) subscribePendingTransactions(wsConn *wsConn) (rpc.ID, error) {
	var query = "subscribePendingTransactions"
	var subID = rpc.NewID()

	sub, _, err := api.events.SubscribePendingTxs()
	if err != nil {
		return "", errors.Wrap(err, "error creating block filter: %s")
	}

	unsubscribed := make(chan struct{})
	api.filtersMu.Lock()
	api.filters[subID] = &wsSubscription{
		sub:          sub,
		wsConn:       wsConn,
		unsubscribed: unsubscribed,
		query:        query,
	}
	api.filtersMu.Unlock()

	go func(txsCh <-chan coretypes.ResultEvent, errCh <-chan error) {
		for {
			select {
			case ev := <-txsCh:
				data, _ := ev.Data.(tmtypes.EventDataTx)
				txHash := common.BytesToHash(tmtypes.Tx(data.Tx).Hash())

				api.filtersMu.RLock()
				for subID, wsSub := range api.filters {
					if wsSub.query != query {
						continue
					}
					// write to ws conn
					res := &SubscriptionNotification{
						Jsonrpc: "2.0",
						Method:  "eth_subscription",
						Params: &SubscriptionResult{
							Subscription: subID,
							Result:       txHash,
						},
					}

					err = wsSub.wsConn.WriteJSON(res)
					if err != nil {
						api.logger.WithError(err).Warningln("error writing header, will drop peer")

						try(func() {
							api.filtersMu.Lock()
							defer api.filtersMu.Unlock()

							if err != websocket.ErrCloseSent {
								_ = wsSub.wsConn.Close()
							}

							delete(api.filters, subID)
							close(wsSub.unsubscribed)
						}, api.logger, "closing websocket peer sub")
					}
				}
				api.filtersMu.RUnlock()
			case err, ok := <-errCh:
				if !ok {
					api.unsubscribe(subID)
					return
				}
				log.WithError(err).Warningln("dropping PendingTransactions WebSocket subscription", subID)
				api.unsubscribe(subID)
			case <-unsubscribed:
				return
			}
		}
	}(sub.eventCh, sub.Err())

	return subID, nil
}

func (api *pubSubAPI) subscribeSyncing(wsConn *wsConn) (rpc.ID, error) {
	return "", nil
}

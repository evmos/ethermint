// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/evmos/ethermint/rpc/codec"
	"github.com/evmos/ethermint/rpc/types"
)

// handler handles JSON-RPC messages. There is one handler per connection. Note that
// handler is not safe for concurrent use. Message handling never blocks indefinitely
// because RPCs are processed on background goroutines launched by handler.
//
// The entry points for incoming messages are:
//
//	h.handleMsg(message)
//	h.handleBatch(message)
//
// Outgoing calls use the requestOp struct. Register the request before sending it
// on the connection:
//
//	op := &requestOp{ids: ...}
//	h.addRequestOp(op)
//
// Now send the request, then wait for the reply to be delivered through handleMsg:
//
//	if err := op.wait(...); err != nil {
//		h.removeRequestOp(op) // timeout, etc.
//	}
type Handler struct {
	reg            *ServiceRegistry
	unsubscribeCb  *types.Callback
	idgen          func() types.ID          // subscription ID generator
	respWait       map[string]*RequestOp    // active client requests
	clientSubs     map[string]*Subscription // active client subscriptions
	callWG         sync.WaitGroup           // pending call goroutines
	rootCtx        context.Context          // canceled by close()
	cancelRoot     func()                   // cancel function for rootCtx
	conn           codec.JSONWriter         // where responses will be sent
	Log            log.Logger
	AllowSubscribe bool

	subLock    sync.Mutex
	serverSubs map[types.ID]*types.Subscription
}

type callProc struct {
	ctx       context.Context
	notifiers []*Notifier
}

func NewHandler(connCtx context.Context, conn codec.JSONWriter, idgen func() types.ID, reg *ServiceRegistry) *Handler {
	rootCtx, cancelRoot := context.WithCancel(connCtx)
	h := &Handler{
		reg:            reg,
		idgen:          idgen,
		conn:           conn,
		respWait:       make(map[string]*RequestOp),
		clientSubs:     make(map[string]*Subscription),
		rootCtx:        rootCtx,
		cancelRoot:     cancelRoot,
		AllowSubscribe: true,
		serverSubs:     make(map[types.ID]*types.Subscription),
		Log:            log.Root(),
	}
	if conn.RemoteAddr() != "" {
		h.Log = h.Log.New("conn", conn.RemoteAddr())
	}
	h.unsubscribeCb = types.NewCallback(reflect.Value{}, reflect.ValueOf(h.unsubscribe))
	return h
}

// handleBatch executes all messages in a batch and returns the responses.
func (h *Handler) HandleBatch(msgs []*codec.JsonrpcMessage) {
	// Emit error response for empty batches:
	if len(msgs) == 0 {
		h.startCallProc(func(cp *callProc) {
			err := h.conn.WriteJSON(cp.ctx, codec.ErrorMessage(&types.InvalidRequestError{Message: "empty batch"}))
			if err != nil {
				log.Trace("batch handler failed to write json when empty msgs", "err", err)
			}
		})
		return
	}

	// Handle non-call messages first:
	calls := make([]*codec.JsonrpcMessage, 0, len(msgs))
	for _, msg := range msgs {
		if handled := h.handleImmediate(msg); !handled {
			calls = append(calls, msg)
		}
	}
	if len(calls) == 0 {
		return
	}
	// Process calls on a goroutine because they may block indefinitely:
	h.startCallProc(func(cp *callProc) {
		answers := make([]*codec.JsonrpcMessage, 0, len(msgs))
		for _, msg := range calls {
			if answer := h.handleCallMsg(cp, msg); answer != nil {
				answers = append(answers, answer)
			}
		}
		h.addSubscriptions(cp.notifiers)
		if len(answers) > 0 {
			err := h.conn.WriteJSON(cp.ctx, answers)
			if err != nil {
				log.Trace("batch handler failed to write json", "err", err)
			}
		}
		for _, n := range cp.notifiers {
			err := n.activate()
			if err != nil {
				log.Trace("call proc notifier failed to activate", "err", err)
			}
		}
	})
}

// handleMsg handles a single message.
func (h *Handler) HandleMsg(msg *codec.JsonrpcMessage) {
	if ok := h.handleImmediate(msg); ok {
		return
	}
	h.startCallProc(func(cp *callProc) {
		answer := h.handleCallMsg(cp, msg)
		h.addSubscriptions(cp.notifiers)
		if answer != nil {
			err := h.conn.WriteJSON(cp.ctx, answer)
			if err != nil {
				log.Trace("handler failed to write json", "err", err)
			}
		}
		for _, n := range cp.notifiers {
			err := n.activate()
			if err != nil {
				log.Trace("call proc notifier failed to activate", "err", err)
			}
		}
	})
}

// close cancels all requests except for inflightReq and waits for
// call goroutines to shut down.
func (h *Handler) Close(err error, inflightReq *RequestOp) {
	h.cancelAllRequests(err, inflightReq)
	h.callWG.Wait()
	h.cancelRoot()
	h.cancelServerSubscriptions(err)
}

// addRequestOp registers a request operation.
func (h *Handler) AddRequestOp(op *RequestOp) {
	for _, id := range op.ids {
		h.respWait[string(id)] = op
	}
}

// removeRequestOps stops waiting for the given request IDs.
func (h *Handler) RemoveRequestOp(op *RequestOp) {
	for _, id := range op.ids {
		delete(h.respWait, string(id))
	}
}

// cancelAllRequests unblocks and removes pending requests and active subscriptions.
func (h *Handler) cancelAllRequests(err error, inflightReq *RequestOp) {
	didClose := make(map[*RequestOp]bool)
	if inflightReq != nil {
		didClose[inflightReq] = true
	}

	for id, op := range h.respWait {
		// Remove the op so that later calls will not close op.resp again.
		delete(h.respWait, id)

		if !didClose[op] {
			op.err = err
			close(op.resp)
			didClose[op] = true
		}
	}
	for id, sub := range h.clientSubs {
		delete(h.clientSubs, id)
		sub.close(err)
	}
}

func (h *Handler) addSubscriptions(nn []*Notifier) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	for _, n := range nn {
		if sub := n.takeSubscription(); sub != nil {
			h.serverSubs[sub.ID] = sub
		}
	}
}

// cancelServerSubscriptions removes all subscriptions and closes their error channels.
func (h *Handler) cancelServerSubscriptions(err error) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	for id, s := range h.serverSubs {
		s.Error <- err
		close(s.Error)
		delete(h.serverSubs, id)
	}
}

// startCallProc runs fn in a new goroutine and starts tracking it in the h.calls wait group.
func (h *Handler) startCallProc(fn func(*callProc)) {
	h.callWG.Add(1)
	go func() {
		ctx, cancel := context.WithCancel(h.rootCtx)
		defer h.callWG.Done()
		defer cancel()
		fn(&callProc{ctx: ctx})
	}()
}

// handleImmediate executes non-call messages. It returns false if the message is a
// call or requires a reply.
func (h *Handler) handleImmediate(msg *codec.JsonrpcMessage) bool {
	start := time.Now()
	switch {
	case msg.IsNotification():
		if strings.HasSuffix(msg.Method, types.NotificationMethodSuffix) {
			h.handleSubscriptionResult(msg)
			return true
		}
		return false
	case msg.IsResponse():
		h.handleResponse(msg)
		h.Log.Trace("Handled RPC response", "reqid", idForLog{msg.ID}, "duration", time.Since(start))
		return true
	default:
		return false
	}
}

// handleSubscriptionResult processes subscription notifications.
func (h *Handler) handleSubscriptionResult(msg *codec.JsonrpcMessage) {
	var result codec.SubscriptionResult
	if err := json.Unmarshal(msg.Params, &result); err != nil {
		h.Log.Debug("Dropping invalid subscription message")
		return
	}
	if h.clientSubs[result.ID] != nil {
		h.clientSubs[result.ID].deliver(result.Result)
	}
}

// handleResponse processes method call responses.
func (h *Handler) handleResponse(msg *codec.JsonrpcMessage) {
	op := h.respWait[string(msg.ID)]
	if op == nil {
		h.Log.Debug("Unsolicited RPC response", "reqid", idForLog{msg.ID})
		return
	}
	delete(h.respWait, string(msg.ID))
	// For normal responses, just forward the reply to Call/BatchCall.
	if op.sub == nil {
		op.resp <- msg
		return
	}
	// For subscription responses, start the subscription if the server
	// indicates success. EthSubscribe gets unblocked in either case through
	// the op.resp channel.
	defer close(op.resp)
	if msg.Error != nil {
		op.err = msg.Error
		return
	}
	err := json.Unmarshal(msg.Result, &op.sub.Subid)
	op.err = err
	if err != nil {
		log.Trace("unmarshal msg result for json rpc failed", "err", err)
	} else {
		go op.sub.run()
		h.clientSubs[op.sub.Subid] = op.sub
	}
}

// handleCallMsg executes a call message and returns the answer.
func (h *Handler) handleCallMsg(ctx *callProc, msg *codec.JsonrpcMessage) *codec.JsonrpcMessage {
	start := time.Now()
	switch {
	case msg.IsNotification():
		h.handleCall(ctx, msg)
		h.Log.Debug("Served "+msg.Method, "duration", time.Since(start))
		return nil
	case msg.IsCall():
		resp := h.handleCall(ctx, msg)
		var ctx []interface{}
		ctx = append(ctx, "reqid", idForLog{msg.ID}, "duration", time.Since(start))
		if resp.Error != nil {
			ctx = append(ctx, "err", resp.Error.Message)
			if resp.Error.Data != nil {
				ctx = append(ctx, "errdata", resp.Error.Data)
			}
			h.Log.Warn("Served "+msg.Method, ctx...)
		} else {
			h.Log.Debug("Served "+msg.Method, ctx...)
		}
		return resp
	case msg.HasValidID():
		return msg.ErrorResponse(&types.InvalidRequestError{Message: "invalid request"})
	default:
		return codec.ErrorMessage(&types.InvalidRequestError{Message: "invalid request"})
	}
}

// handleCall processes method calls.
func (h *Handler) handleCall(cp *callProc, msg *codec.JsonrpcMessage) *codec.JsonrpcMessage {
	if msg.IsSubscribe() {
		return h.handleSubscribe(cp, msg)
	}
	var callb *types.Callback
	if msg.IsUnsubscribe() {
		callb = h.unsubscribeCb
	} else {
		callb = h.reg.callback(msg.Method)
	}
	if callb == nil {
		return msg.ErrorResponse(&types.MethodNotFoundError{Method: msg.Method})
	}
	args, err := codec.ParsePositionalArguments(msg.Params, callb.ArgTypes)
	if err != nil {
		return msg.ErrorResponse(&types.InvalidParamsError{Message: err.Error()})
	}
	start := time.Now()
	answer := h.runMethod(cp.ctx, msg, callb, args)

	// Collect the statistics for RPC calls if metrics is enabled.
	// We only care about pure rpc call. Filter out subscription.
	if callb != h.unsubscribeCb {
		if MetricsEnabled {
			rpcRequestGauge.Inc(1)
			if answer.Error != nil {
				failedRequestGauge.Inc(1)
			} else {
				successfulRequestGauge.Inc(1)
			}
			rpcServingTimer.UpdateSince(start)
			updateServeTimeHistogram(msg.Method, answer.Error == nil, time.Since(start))
		}
	}
	return answer
}

// handleSubscribe processes *_subscribe method calls.
func (h *Handler) handleSubscribe(cp *callProc, msg *codec.JsonrpcMessage) *codec.JsonrpcMessage {
	if !h.AllowSubscribe {
		return msg.ErrorResponse(&types.InternalServerError{
			Code:    types.ErrcodeNotificationsUnsupported,
			Message: types.ErrNotificationsUnsupported.Error(),
		})
	}

	// Subscription method name is first argument.
	name, err := codec.ParseSubscriptionName(msg.Params)
	if err != nil {
		return msg.ErrorResponse(&types.InvalidParamsError{Message: err.Error()})
	}
	namespace := msg.Namespace()
	callb := h.reg.subscription(namespace, name)
	if callb == nil {
		return msg.ErrorResponse(&types.SubscriptionNotFoundError{Namespace: namespace, Subscription: name})
	}

	// Parse subscription name arg too, but remove it before calling the callback.
	argTypes := append([]reflect.Type{types.StringType}, callb.ArgTypes...)
	args, err := codec.ParsePositionalArguments(msg.Params, argTypes)
	if err != nil {
		return msg.ErrorResponse(&types.InvalidParamsError{Message: err.Error()})
	}
	args = args[1:]

	// Install notifier in context so the subscription handler can find it.
	n := &Notifier{h: h, namespace: namespace}
	cp.notifiers = append(cp.notifiers, n)
	ctx := context.WithValue(cp.ctx, notifierKey{}, n)

	return h.runMethod(ctx, msg, callb, args)
}

// runMethod runs the Go callback for an RPC method.
func (h *Handler) runMethod(ctx context.Context, msg *codec.JsonrpcMessage, callb *types.Callback, args []reflect.Value) *codec.JsonrpcMessage {
	result, err := callb.Call(ctx, msg.Method, args)
	if err != nil {
		return msg.ErrorResponse(err)
	}
	return msg.Response(result)
}

// unsubscribe is the callback function for all *_unsubscribe calls.
func (h *Handler) unsubscribe(ctx context.Context, id types.ID) (bool, error) {
	h.subLock.Lock()
	defer h.subLock.Unlock()

	s := h.serverSubs[id]
	if s == nil {
		return false, types.ErrSubscriptionNotFound
	}
	close(s.Error)
	delete(h.serverSubs, id)
	return true, nil
}

type idForLog struct{ json.RawMessage }

func (id idForLog) String() string {
	if s, err := strconv.Unquote(string(id.RawMessage)); err == nil {
		return s
	}
	return string(id.RawMessage)
}

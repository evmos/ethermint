// Copyright 2016 The go-ethereum Authors
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
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/evmos/ethermint/rpc/codec"
	"github.com/evmos/ethermint/rpc/types"
)

type notifierKey struct{}

// NotifierFromContext returns the Notifier value stored in ctx, if any.
func NotifierFromContext(ctx context.Context) (*Notifier, bool) {
	n, ok := ctx.Value(notifierKey{}).(*Notifier)
	return n, ok
}

// Notifier is tied to a RPC connection that supports subscriptions.
// Server callbacks use the notifier to send notifications.
type Notifier struct {
	h         *Handler
	namespace string

	mu           sync.Mutex
	sub          *types.Subscription
	buffer       []json.RawMessage
	callReturned bool
	activated    bool
}

// CreateSubscription returns a new subscription that is coupled to the
// RPC connection. By default subscriptions are inactive and notifications
// are dropped until the subscription is marked as active. This is done
// by the RPC server after the subscription ID is send to the client.
func (n *Notifier) CreateSubscription() *types.Subscription {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.sub != nil {
		panic("can't create multiple subscriptions with Notifier")
	} else if n.callReturned {
		panic("can't create subscription after subscribe call has returned")
	}
	n.sub = &types.Subscription{ID: n.h.idgen(), Namespace: n.namespace, Error: make(chan error, 1)}
	return n.sub
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
func (n *Notifier) Notify(id types.ID, data interface{}) error {
	enc, err := json.Marshal(data)
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if n.sub == nil {
		panic("can't Notify before subscription is created")
	} else if n.sub.ID != id {
		panic("Notify with wrong ID")
	}
	if n.activated {
		return n.send(n.sub, enc)
	}
	n.buffer = append(n.buffer, enc)
	return nil
}

// Closed returns a channel that is closed when the RPC connection is closed.
// Deprecated: use subscription error channel
func (n *Notifier) Closed() <-chan interface{} {
	return n.h.conn.Closed()
}

// takeSubscription returns the subscription (if one has been created). No subscription can
// be created after this call.
func (n *Notifier) takeSubscription() *types.Subscription {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.callReturned = true
	return n.sub
}

// activate is called after the subscription ID was sent to client. Notifications are
// buffered before activation. This prevents notifications being sent to the client before
// the subscription ID is sent to the client.
func (n *Notifier) activate() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, data := range n.buffer {
		if err := n.send(n.sub, data); err != nil {
			return err
		}
	}
	n.activated = true
	return nil
}

func (n *Notifier) send(sub *types.Subscription, data json.RawMessage) error {
	params, err := json.Marshal(&codec.SubscriptionResult{ID: string(sub.ID), Result: data})
	if err != nil {
		log.Trace("marshal subscription result failed", "err", err)
		return err
	}
	ctx := context.Background()
	return n.h.conn.WriteJSON(ctx, &codec.JsonrpcMessage{
		Version: types.Vsn,
		Method:  n.namespace + types.NotificationMethodSuffix,
		Params:  params,
	})
}

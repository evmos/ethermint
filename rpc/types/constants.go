package types

import (
	"sync"
	"time"
)

const (
	Vsn                      = "2.0"
	ServiceMethodSeparator   = "_"
	SubscribeMethodSuffix    = "_subscribe"
	UnsubscribeMethodSuffix  = "_unsubscribe"
	NotificationMethodSuffix = "_subscription"

	DefaultWriteTimeout = 10 * time.Second // used if context has no deadline

	// Timeouts
	DefaultDialTimeout = 10 * time.Second // used if context has no deadline
	SubscribeTimeout   = 5 * time.Second  // overall timeout eth_subscribe, rpc_modules calls

	// Subscriptions are removed when the subscriber cannot keep up.
	//
	// This can be worked around by supplying a channel with sufficiently sized buffer,
	// but this can be inconvenient and hard to explain in the docs. Another issue with
	// buffered channels is that the buffer is static even though it might not be needed
	// most of the time.
	//
	// The approach taken here is to maintain a per-subscription linked list buffer
	// shrinks on demand. If the buffer reaches the size below, the subscription is
	// dropped.
	MaxClientSubscriptionBuffer = 20000

	WsReadBuffer       = 1024
	WsWriteBuffer      = 1024
	WsPingInterval     = 60 * time.Second
	WsPingWriteTimeout = 5 * time.Second
	WsPongTimeout      = 30 * time.Second
	WsMessageSizeLimit = 15 * 1024 * 1024

	MaxRequestContentLength = 1024 * 1024 * 5
	ContentType             = "application/json"
)

var WsBufferPool = new(sync.Pool)

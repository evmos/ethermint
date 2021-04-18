package rpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/ethereum/rpc/pubsub"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

var (
	txEvents     = tmtypes.QueryForEvent(tmtypes.EventTx).String()
	evmEvents    = tmquery.MustParse(fmt.Sprintf("%s='%s' AND %s.%s='%s'", tmtypes.EventTypeKey, tmtypes.EventTx, sdk.EventTypeMessage, sdk.AttributeKeyModule, evmtypes.ModuleName)).String()
	headerEvents = tmtypes.QueryForEvent(tmtypes.EventNewBlockHeader).String()
)

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria using the Tendermint's RPC client.
type EventSystem struct {
	ctx        context.Context
	tmWSClient *rpcclient.WSClient

	// light client mode
	lightMode bool

	index      filterIndex
	topicChans map[string]chan<- coretypes.ResultEvent
	indexMux   *sync.RWMutex

	// Channels
	install   chan *Subscription // install filter for event notification
	uninstall chan *Subscription // remove filter for event notification
	eventBus  pubsub.EventBus
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(tmWSClient *rpcclient.WSClient) *EventSystem {
	index := make(filterIndex)
	for i := filters.UnknownSubscription; i < filters.LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*Subscription)
	}

	es := &EventSystem{
		ctx:        context.Background(),
		tmWSClient: tmWSClient,
		lightMode:  false,
		index:      index,
		topicChans: make(map[string]chan<- coretypes.ResultEvent, len(index)),
		indexMux:   new(sync.RWMutex),
		install:    make(chan *Subscription),
		uninstall:  make(chan *Subscription),
		eventBus:   pubsub.NewEventBus(),
	}

	go es.eventLoop()
	go es.consumeEvents()
	return es
}

// WithContext sets a new context to the EventSystem. This is required to set a timeout context when
// a new filter is intantiated.
func (es *EventSystem) WithContext(ctx context.Context) {
	es.ctx = ctx
}

// subscribe performs a new event subscription to a given Tendermint event.
// The subscription creates a unidirectional receive event channel to receive the ResultEvent.
func (es *EventSystem) subscribe(sub *Subscription) (*Subscription, context.CancelFunc, error) {
	var (
		err      error
		cancelFn context.CancelFunc
	)

	es.ctx, cancelFn = context.WithCancel(context.Background())

	existingSubs := es.eventBus.Topics()
	for _, topic := range existingSubs {
		if topic == sub.event {
			eventCh, err := es.eventBus.Subscribe(sub.event)
			if err != nil {
				err := errors.Wrapf(err, "failed to subscribe to topic: %s", sub.event)
				return nil, cancelFn, err
			}

			sub.eventCh = eventCh
			return sub, cancelFn, nil
		}
	}

	switch sub.typ {
	case filters.LogsSubscription:
		err = es.tmWSClient.Subscribe(es.ctx, sub.event)
	case filters.BlocksSubscription:
		err = es.tmWSClient.Subscribe(es.ctx, sub.event)
	default:
		err = fmt.Errorf("invalid filter subscription type %d", sub.typ)
	}

	if err != nil {
		sub.err <- err
		return nil, cancelFn, err
	}

	// wrap events in a go routine to prevent blocking
	es.install <- sub
	<-sub.installed

	eventCh, err := es.eventBus.Subscribe(sub.event)
	if err != nil {
		err := errors.Wrapf(err, "failed to subscribe to topic after installed: %s", sub.event)
		return sub, cancelFn, err
	}

	sub.eventCh = eventCh
	return sub, cancelFn, nil
}

// SubscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel. Default value for the from and to
// block is "latest". If the fromBlock > toBlock an error is returned.
func (es *EventSystem) SubscribeLogs(crit filters.FilterCriteria) (*Subscription, context.CancelFunc, error) {
	var from, to rpc.BlockNumber
	if crit.FromBlock == nil {
		from = rpc.LatestBlockNumber
	} else {
		from = rpc.BlockNumber(crit.FromBlock.Int64())
	}
	if crit.ToBlock == nil {
		to = rpc.LatestBlockNumber
	} else {
		to = rpc.BlockNumber(crit.ToBlock.Int64())
	}

	switch {
	// only interested in new mined logs, mined logs within a specific block range, or
	// logs from a specific block number to new mined blocks
	case (from == rpc.LatestBlockNumber && to == rpc.LatestBlockNumber),
		(from >= 0 && to >= 0 && to >= from):
		return es.subscribeLogs(crit)

	default:
		return nil, nil, fmt.Errorf("invalid from and to block combination: from > to (%d > %d)", from, to)
	}
}

// subscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel.
func (es *EventSystem) subscribeLogs(crit filters.FilterCriteria) (*Subscription, context.CancelFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.LogsSubscription,
		event:     evmEvents,
		logsCrit:  crit,
		created:   time.Now().UTC(),
		logs:      make(chan []*ethtypes.Log),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

// SubscribeNewHeads subscribes to new block headers events.
func (es EventSystem) SubscribeNewHeads() (*Subscription, context.CancelFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.BlocksSubscription,
		event:     headerEvents,
		created:   time.Now().UTC(),
		headers:   make(chan *ethtypes.Header),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

// SubscribePendingTxs subscribes to new pending transactions events from the mempool.
func (es EventSystem) SubscribePendingTxs() (*Subscription, context.CancelFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.PendingTransactionsSubscription,
		event:     txEvents,
		created:   time.Now().UTC(),
		hashes:    make(chan []common.Hash),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

type filterIndex map[filters.Type]map[rpc.ID]*Subscription

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	for {
		select {
		case f := <-es.install:
			es.indexMux.Lock()
			es.index[f.typ][f.id] = f
			ch := make(chan coretypes.ResultEvent)
			es.topicChans[f.event] = ch
			if err := es.eventBus.AddTopic(f.event, ch); err != nil {
				log.WithField("topic", f.event).WithError(err).Errorln("failed to add event topic to event bus")
			}
			es.indexMux.Unlock()
			close(f.installed)
		case f := <-es.uninstall:
			es.indexMux.Lock()
			delete(es.index[f.typ], f.id)

			var channelInUse bool
			for _, sub := range es.index[f.typ] {
				if sub.event == f.event {
					channelInUse = true
					break
				}
			}

			// remove topic only when channel is not used by other subscriptions
			if !channelInUse {
				if err := es.tmWSClient.Unsubscribe(es.ctx, f.event); err != nil {
					log.WithError(err).WithField("query", f.event).Errorln("failed to unsubscribe from query")
				}

				ch, ok := es.topicChans[f.event]
				if ok {
					es.eventBus.RemoveTopic(f.event)
					close(ch)
					delete(es.topicChans, f.event)
				}
			}

			es.indexMux.Unlock()
			close(f.err)
		}
	}
}

func (es *EventSystem) consumeEvents() {
	for {
		for rpcResp := range es.tmWSClient.ResponsesCh {
			var ev coretypes.ResultEvent

			if rpcResp.Error != nil {
				time.Sleep(5 * time.Second)
				continue
			} else if err := tmjson.Unmarshal(rpcResp.Result, &ev); err != nil {
				log.WithError(err).Warningln("failed to JSON unmarshal ResponsesCh result event")
				continue
			}

			if len(ev.Query) == 0 {
				// skip empty responses
				continue
			}

			es.indexMux.RLock()
			ch, ok := es.topicChans[ev.Query]
			es.indexMux.RUnlock()
			if !ok {
				log.WithField("topic", ev.Query).Warningln("channel for subscription not found, lol")
				log.Infoln("available channels:", es.eventBus.Topics())
				continue
			}

			// gracefully handle lagging subscribers
			t := time.NewTimer(time.Second)
			select {
			case <-t.C:
				log.WithField("topic", ev.Query).Warningln("dropped event during lagging subscription")
			case ch <- ev:
			}
		}

		time.Sleep(time.Second)
	}
}

// Subscription defines a wrapper for the private subscription
type Subscription struct {
	id        rpc.ID
	typ       filters.Type
	event     string
	created   time.Time
	logsCrit  filters.FilterCriteria
	logs      chan []*ethtypes.Log
	hashes    chan []common.Hash
	headers   chan *ethtypes.Header
	installed chan struct{} // closed when the filter is installed
	eventCh   <-chan coretypes.ResultEvent
	err       chan error
}

// ID returns the underlying subscription RPC identifier.
func (s Subscription) ID() rpc.ID {
	return s.id
}

// Unsubscribe from the current subscription to Tendermint Websocket. It sends an error to the
// subscription error channel if unsubscription fails.
func (s *Subscription) Unsubscribe(es *EventSystem) {
	go func() {
	uninstallLoop:
		for {
			// write uninstall request and consume logs/hashes. This prevents
			// the eventLoop broadcast method to deadlock when writing to the
			// filter event channel while the subscription loop is waiting for
			// this method to return (and thus not reading these events).
			select {
			case es.uninstall <- s:
				break uninstallLoop
			case <-s.logs:
			case <-s.hashes:
			case <-s.headers:
			}
		}
	}()
}

// Err returns the error channel
func (s *Subscription) Err() <-chan error {
	return s.err
}

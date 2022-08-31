package filters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	tmquery "github.com/tendermint/tendermint/libs/pubsub/query"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/rpc/ethereum/pubsub"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var (
	txEvents  = tmtypes.QueryForEvent(tmtypes.EventTx).String()
	evmEvents = tmquery.MustCompile(fmt.Sprintf("%s='%s' AND %s.%s='%s'",
		tmtypes.EventTypeKey,
		tmtypes.EventTx,
		sdk.EventTypeMessage,
		sdk.AttributeKeyModule, evmtypes.ModuleName)).String()
	headerEvents = tmtypes.QueryForEvent(tmtypes.EventNewBlockHeader).String()
)

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria using the Tendermint's RPC client.
type EventSystem struct {
	logger log.Logger
	ctx    context.Context
	client rpcclient.Client

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
func NewEventSystem(logger log.Logger, client rpcclient.Client) *EventSystem {
	index := make(filterIndex)
	for i := filters.UnknownSubscription; i < filters.LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*Subscription)
	}

	es := &EventSystem{
		logger:     logger,
		ctx:        context.Background(),
		client:     client,
		lightMode:  false,
		index:      index,
		topicChans: make(map[string]chan<- coretypes.ResultEvent, len(index)),
		indexMux:   new(sync.RWMutex),
		install:    make(chan *Subscription),
		uninstall:  make(chan *Subscription),
		eventBus:   pubsub.NewEventBus(),
	}
	return es
}

// WithContext sets a new context to the EventSystem. This is required to set a timeout context when
// a new filter is intantiated.
func (es *EventSystem) WithContext(ctx context.Context) {
	es.ctx = ctx
}

// subscribe performs a new event subscription to a given Tendermint event.
// The subscription creates a unidirectional receive event channel to receive the ResultEvent.
func (es *EventSystem) subscribe(sub *Subscription) (*Subscription, pubsub.UnsubscribeFunc, error) {
	var (
		err      error
		cancelFn context.CancelFunc
	)

	ctx, cancelFn := context.WithCancel(context.Background())
	var query string
	switch sub.typ {
	case filters.LogsSubscription, filters.BlocksSubscription, filters.PendingTransactionsSubscription:
		query = sub.event
	default:
		err = fmt.Errorf("invalid filter subscription type %d", sub.typ)
	}

	if err != nil {
		sub.err <- err
		return nil, func() { cancelFn() }, err
	}

	eventCh := make(chan *coretypes.ResultEvents, 0)
	go func() {
		defer func() {
			close(eventCh)
		}()
		for {
			select {
			case <-ctx.Done():
				return

			default:
				filter := coretypes.EventFilter{Query: query}
				res, err := es.client.Events(ctx, &coretypes.RequestEvents{
					Filter:   &filter,
					MaxItems: 100,
					After:    sub.after,
					WaitTime: 1000 * time.Second,
				})
				if err != nil {
					sub.err <- err
					return
				}
				eventCh <- res

				if len(res.Items) > 0 {
					sub.after = res.Items[len(res.Items)-1].Cursor
				}
			}
		}
	}()
	sub.eventCh = eventCh
	return sub, func() { cancelFn() }, nil
}

// SubscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel. Default value for the from and to
// block is "latest". If the fromBlock > toBlock an error is returned.
func (es *EventSystem) SubscribeLogs(crit filters.FilterCriteria) (*Subscription, pubsub.UnsubscribeFunc, error) {
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
		(from >= 0 && to >= 0 && to >= from),
		(from >= 0 && to == rpc.LatestBlockNumber):
		return es.subscribeLogs(crit)

	default:
		return nil, nil, fmt.Errorf("invalid from and to block combination: from > to (%d > %d)", from, to)
	}
}

// subscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel.
func (es *EventSystem) subscribeLogs(crit filters.FilterCriteria) (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.LogsSubscription,
		event:     evmEvents,
		logsCrit:  crit,
		created:   time.Now().UTC(),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

// SubscribeNewHeads subscribes to new block headers events.
func (es EventSystem) SubscribeNewHeads() (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.BlocksSubscription,
		event:     headerEvents,
		created:   time.Now().UTC(),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

// SubscribePendingTxs subscribes to new pending transactions events from the mempool.
func (es EventSystem) SubscribePendingTxs() (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.PendingTransactionsSubscription,
		event:     txEvents,
		created:   time.Now().UTC(),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

type filterIndex map[filters.Type]map[rpc.ID]*Subscription

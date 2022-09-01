package filters

import (
	"time"

	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Subscription defines a wrapper for the private subscription
type Subscription struct {
	id       rpc.ID
	typ      filters.Type
	event    string
	created  time.Time
	logsCrit filters.FilterCriteria
	eventCh  <-chan *coretypes.ResultEvents
	err      chan error
	after    string
}

// ID returns the underlying subscription RPC identifier.
func (s Subscription) ID() rpc.ID {
	return s.id
}

// Unsubscribe from the current subscription to Tendermint Websocket. It sends an error to the
// subscription error channel if unsubscribe fails.
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
			}
		}
	}()
}

// Err returns the error channel
func (s *Subscription) Err() <-chan error {
	return s.err
}

// Event returns the tendermint result event channel
func (s *Subscription) Event() <-chan *coretypes.ResultEvents {
	return s.eventCh
}

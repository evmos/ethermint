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

// Keep life circle for posterity
func (s *Subscription) Unsubscribe(es *EventSystem) {
}

// Err returns the error channel
func (s *Subscription) Err() <-chan error {
	return s.err
}

// Event returns the tendermint result event channel
func (s *Subscription) Event() <-chan *coretypes.ResultEvents {
	return s.eventCh
}

package types

import (
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// ID defines a pseudo random number that is used to identify RPC subscriptions.
type ID string

var globalGen = RandomIDGenerator()

// NewID returns a new, random ID.
func NewID() ID {
	return globalGen()
}

// randomIDGenerator returns a function generates a random IDs.
func RandomIDGenerator() func() ID {
	buf := make([]byte, 8)
	var seed int64
	if _, err := crand.Read(buf); err == nil {
		seed = int64(binary.BigEndian.Uint64(buf))
	} else {
		seed = int64(time.Now().Nanosecond())
	}

	var (
		mu  sync.Mutex
		rng = rand.New(rand.NewSource(seed)) //nolint:gosec // copied as is from go-ethereum
	)
	return func() ID {
		mu.Lock()
		defer mu.Unlock()
		id := make([]byte, 16)
		_, err := rng.Read(id)
		if err != nil {
			log.Trace("generating random byes failed", "err", err)
		}
		return EncodeID(id)
	}
}

func EncodeID(b []byte) ID {
	id := hex.EncodeToString(b)
	id = strings.TrimLeft(id, "0")
	if id == "" {
		id = "0" // ID's are RPC quantities, no leading zero's and 0 is 0x0.
	}
	return ID("0x" + id)
}

// A Subscription is created by a notifier and tied to that notifier. The client can use
// this subscription to wait for an unsubscribe request for the client, see Err().
type Subscription struct {
	ID        ID
	Namespace string
	Error     chan error // closed on unsubscribe
}

// Err returns a channel that is closed when the client send an unsubscribe request.
func (s *Subscription) Err() <-chan error {
	return s.Error
}

// MarshalJSON marshals a subscription as its ID.
func (s *Subscription) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ID)
}

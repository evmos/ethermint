package codec

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/evmos/ethermint/rpc/types"
)

// JSONCodec reads and writes JSON-RPC messages to the underlying connection. It also has
// support for parsing arguments and serializing (result) objects.
type JSONCodec struct {
	remote  string
	closer  sync.Once                 // close closed channel once
	closeCh chan interface{}          // closed on Close
	decode  func(v interface{}) error // decoder to allow multiple transports
	encMu   sync.Mutex                // guards the encoder
	encode  func(v interface{}) error // encoder to allow multiple transports
	conn    deadlineCloser
}

// NewFuncCodec creates a codec which uses the given functions to read and write. If conn
// implements ConnRemoteAddr, log messages will use it to include the remote address of
// the connection.
func NewFuncCodec(conn deadlineCloser, encode, decode func(v interface{}) error) ServerCodec {
	codec := &JSONCodec{
		closeCh: make(chan interface{}),
		encode:  encode,
		decode:  decode,
		conn:    conn,
	}
	if ra, ok := conn.(ConnRemoteAddr); ok {
		codec.remote = ra.RemoteAddr()
	}
	return codec
}

// NewCodec creates a codec on the given connection. If conn implements ConnRemoteAddr, log
// messages will use it to include the remote address of the connection.
func NewCodec(conn Conn) ServerCodec {
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	dec.UseNumber()
	return NewFuncCodec(conn, enc.Encode, dec.Decode)
}

func (c *JSONCodec) PeerInfo() PeerInfo {
	// This returns "ipc" because all other built-in transports have a separate codec type.
	return PeerInfo{Transport: "ipc", RemoteAddr: c.remote}
}

func (c *JSONCodec) RemoteAddr() string {
	return c.remote
}

func (c *JSONCodec) ReadBatch() (messages []*JsonrpcMessage, batch bool, err error) {
	// Decode the next JSON object in the input stream.
	// This verifies basic syntax, etc.
	var rawmsg json.RawMessage
	if err := c.decode(&rawmsg); err != nil {
		return nil, false, err
	}
	messages, batch = parseMessage(rawmsg)
	for i, msg := range messages {
		if msg == nil {
			// Message is JSON 'null'. Replace with zero value so it
			// will be treated like any other invalid message.
			messages[i] = new(JsonrpcMessage)
		}
	}
	return messages, batch, nil
}

func (c *JSONCodec) WriteJSON(ctx context.Context, v interface{}) error {
	c.encMu.Lock()
	defer c.encMu.Unlock()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(types.DefaultWriteTimeout)
	}
	err := c.conn.SetWriteDeadline(deadline)
	if err != nil {
		log.Trace("set write deadline failed for json", "err", err)
	}
	return c.encode(v)
}

func (c *JSONCodec) Close() {
	c.closer.Do(func() {
		close(c.closeCh)
		err := c.conn.Close()
		if err != nil {
			log.Trace("jsoncodec close connection failed", "err", err)
		}
	})
}

// Closed returns a channel which will be closed when Close is called
func (c *JSONCodec) Closed() <-chan interface{} {
	return c.closeCh
}

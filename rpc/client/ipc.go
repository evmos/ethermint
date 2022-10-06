package client

import (
	"context"

	"github.com/evmos/ethermint/rpc/codec"
)

// DialIPC create a new IPC client that connects to the given endpoint. On Unix it assumes
// the endpoint is the full path to a unix socket, and Windows the endpoint is an
// identifier for a named pipe.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialIPC(ctx context.Context, endpoint string) (*Client, error) {
	return NewClient(ctx, newClientTransportIPC(endpoint))
}

func newClientTransportIPC(endpoint string) ReconnectFunc {
	return func(ctx context.Context) (codec.ServerCodec, error) {
		conn, err := newIPCConnection(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		return codec.NewCodec(conn), err
	}
}

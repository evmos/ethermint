package client

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"

	"github.com/evmos/ethermint/rpc/codec"
	rpcTypes "github.com/evmos/ethermint/rpc/types"
)

// DialWebsocketWithDialer creates a new RPC client using WebSocket.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
//
// Deprecated: use DialOptions and the WithWebsocketDialer option.
func DialWebsocketWithDialer(ctx context.Context, endpoint, origin string, dialer websocket.Dialer) (*Client, error) {
	cfg := new(clientConfig)
	cfg.wsDialer = &dialer
	if origin != "" {
		cfg.setHeader("origin", origin)
	}
	connect, err := NewClientTransportWS(endpoint, cfg)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, connect)
}

// DialWebsocket creates a new RPC client that communicates with a JSON-RPC server
// that is listening on the given endpoint.
//
// The context is used for the initial connection establishment. It does not
// affect subsequent interactions with the client.
func DialWebsocket(ctx context.Context, endpoint, origin string) (*Client, error) {
	cfg := new(clientConfig)
	if origin != "" {
		cfg.setHeader("origin", origin)
	}
	connect, err := NewClientTransportWS(endpoint, cfg)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, connect)
}

func NewClientTransportWS(endpoint string, cfg *clientConfig) (ReconnectFunc, error) {
	dialer := cfg.wsDialer
	if dialer == nil {
		dialer = &websocket.Dialer{
			ReadBufferSize:  rpcTypes.WsReadBuffer,
			WriteBufferSize: rpcTypes.WsWriteBuffer,
			WriteBufferPool: rpcTypes.WsBufferPool,
		}
	}

	dialURL, header, err := WsClientHeaders(endpoint, "")
	if err != nil {
		return nil, err
	}
	for key, values := range cfg.httpHeaders {
		header[key] = values
	}

	connect := func(ctx context.Context) (codec.ServerCodec, error) {
		header := header.Clone()
		if cfg.httpAuth != nil {
			if err := cfg.httpAuth(header); err != nil {
				return nil, err
			}
		}
		conn, resp, err := dialer.DialContext(ctx, dialURL, header) //nolint:bodyclose // not fixed as imported from go-ethereum
		if err != nil {
			hErr := rpcTypes.WsHandshakeError{Err: err}
			if resp != nil {
				hErr.Status = resp.Status
			}
			return nil, hErr
		}
		return codec.NewWebsocketCodec(conn, dialURL, header), nil
	}
	return connect, nil
}

func WsClientHeaders(endpoint, origin string) (string, http.Header, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return endpoint, nil, err
	}
	header := make(http.Header)
	if origin != "" {
		header.Add("origin", origin)
	}
	if endpointURL.User != nil {
		b64auth := base64.StdEncoding.EncodeToString([]byte(endpointURL.User.String()))
		header.Add("authorization", "Basic "+b64auth)
		endpointURL.User = nil
	}
	return endpointURL.String(), header, nil
}

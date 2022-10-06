package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/evmos/ethermint/rpc/codec"
	"github.com/evmos/ethermint/rpc/types"
)

// DialHTTP creates a new RPC client that connects to an RPC server over HTTP.
func DialHTTP(endpoint string) (*Client, error) {
	return DialHTTPWithClient(endpoint, new(http.Client))
}

// DialHTTPWithClient creates a new RPC client that connects to an RPC server over HTTP
// using the provided HTTP Client.
//
// Deprecated: use DialOptions and the WithHTTPClient option.
func DialHTTPWithClient(endpoint string, client *http.Client) (*Client, error) {
	// Sanity check URL so we don't end up with a client that will fail every request.
	_, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var cfg clientConfig
	fn := newClientTransportHTTP(endpoint, &cfg)
	return NewClient(context.Background(), fn)
}

func newClientTransportHTTP(endpoint string, cfg *clientConfig) ReconnectFunc {
	headers := make(http.Header, 2+len(cfg.httpHeaders))
	headers.Set("accept", types.ContentType)
	headers.Set("content-type", types.ContentType)
	for key, values := range cfg.httpHeaders {
		headers[key] = values
	}

	client := cfg.httpClient
	if client == nil {
		client = new(http.Client)
	}

	hc := &codec.HTTPConn{
		Client:  client,
		Headers: headers,
		URL:     endpoint,
		Auth:    cfg.httpAuth,
		CloseCh: make(chan interface{}),
	}

	return func(ctx context.Context) (codec.ServerCodec, error) {
		return hc, nil
	}
}

func (c *Client) sendHTTP(ctx context.Context, op *RequestOp, msg interface{}) error {
	hc := c.writeConn.(*codec.HTTPConn)
	respBody, err := hc.DoRequest(ctx, msg)
	if err != nil {
		return err
	}
	defer respBody.Close()

	var respmsg codec.JsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsg); err != nil {
		return err
	}
	op.resp <- &respmsg
	return nil
}

func (c *Client) sendBatchHTTP(ctx context.Context, op *RequestOp, msgs []*codec.JsonrpcMessage) error {
	hc := c.writeConn.(*codec.HTTPConn)
	respBody, err := hc.DoRequest(ctx, msgs)
	if err != nil {
		return err
	}
	defer respBody.Close()
	var respmsgs []codec.JsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsgs); err != nil {
		return err
	}
	for i := 0; i < len(respmsgs); i++ {
		op.resp <- &respmsgs[i]
	}
	return nil
}

// httpServerConn turns a HTTP connection into a Conn.
type httpServerConn struct {
	io.Reader
	io.Writer
	r *http.Request
}

func NewHTTPServerConn(r *http.Request, w http.ResponseWriter) codec.ServerCodec {
	body := io.LimitReader(r.Body, types.MaxRequestContentLength)
	conn := &httpServerConn{Reader: body, Writer: w, r: r}
	return codec.NewCodec(conn)
}

// Close does nothing and always returns nil.
func (t *httpServerConn) Close() error { return nil }

// RemoteAddr returns the peer address of the underlying connection.
func (t *httpServerConn) RemoteAddr() string {
	return t.r.RemoteAddr
}

// SetWriteDeadline does nothing and always returns nil.
func (t *httpServerConn) SetWriteDeadline(time.Time) error { return nil }

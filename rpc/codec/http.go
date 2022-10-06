// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package codec

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/evmos/ethermint/rpc/types"
)

// A HTTPAuth function is called by the client whenever a HTTP request is sent.
// The function must be safe for concurrent use.
//
// Usually, HTTPAuth functions will call h.Set("authorization", "...") to add
// auth information to the request.
type HTTPAuth func(h http.Header) error

type HTTPConn struct {
	Client    *http.Client
	URL       string
	CloseOnce sync.Once
	CloseCh   chan interface{}
	Mu        sync.Mutex // protects headers
	Headers   http.Header
	Auth      HTTPAuth
}

// HttpConn implements ServerCodec, but it is treated specially by Client
// and some methods don't work. The panic() stubs here exist to ensure
// this special treatment is correct.

func (hc *HTTPConn) WriteJSON(context.Context, interface{}) error {
	panic("writeJSON called on HTTPConn")
}

func (hc *HTTPConn) PeerInfo() PeerInfo {
	panic("peerInfo called on HTTPConn")
}

func (hc *HTTPConn) RemoteAddr() string {
	return hc.URL
}

func (hc *HTTPConn) ReadBatch() ([]*JsonrpcMessage, bool, error) {
	<-hc.CloseCh
	return nil, false, io.EOF
}

func (hc *HTTPConn) Close() {
	hc.CloseOnce.Do(func() { close(hc.CloseCh) })
}

func (hc *HTTPConn) Closed() <-chan interface{} {
	return hc.CloseCh
}

func (hc *HTTPConn) DoRequest(ctx context.Context, msg interface{}) (io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", hc.URL, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }

	// set headers
	hc.Mu.Lock()
	req.Header = hc.Headers.Clone()
	hc.Mu.Unlock()
	if hc.Auth != nil {
		if err := hc.Auth(req.Header); err != nil {
			return nil, err
		}
	}

	// do request
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		var body []byte
		if _, err := buf.ReadFrom(resp.Body); err == nil {
			body = buf.Bytes()
		}

		return nil, types.HTTPError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return resp.Body, nil
}

// HTTPTimeouts represents the configuration params for the HTTP RPC server.
type HTTPTimeouts struct {
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If ReadHeaderTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, there is no timeout.
	ReadHeaderTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration
}

// DefaultHTTPTimeouts represents the default timeout values used if further
// configuration is not provided.
var DefaultHTTPTimeouts = HTTPTimeouts{
	ReadTimeout:       30 * time.Second,
	ReadHeaderTimeout: 30 * time.Second,
	WriteTimeout:      30 * time.Second,
	IdleTimeout:       120 * time.Second,
}

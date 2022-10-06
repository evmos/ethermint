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

package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"sync/atomic"

	mapset "github.com/deckarep/golang-set"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"

	"github.com/evmos/ethermint/rpc/client"
	rpcCodec "github.com/evmos/ethermint/rpc/codec"
	"github.com/evmos/ethermint/rpc/types"
)

const (
	MetadataAPI = "rpc"
	EngineAPI   = "engine"
)

// https://www.jsonrpc.org/historical/json-rpc-over-http.html#id13
var acceptedContentTypes = []string{types.ContentType, "application/json-rpc", "application/jsonrequest"}

// CodecOption specifies which type of messages a codec supports.
//
// Deprecated: this option is no longer honored by Server.
type CodecOption int

const (
	// OptionMethodInvocation is an indication that the codec supports RPC method calls
	OptionMethodInvocation CodecOption = 1 << iota

	// OptionSubscriptions is an indication that the codec supports RPC notifications
	OptionSubscriptions = 1 << iota // support pub sub
)

// Server is an RPC server.
type Server struct {
	services client.ServiceRegistry
	Idgen    func() types.ID
	run      int32
	codecs   mapset.Set
}

// NewServer creates a new server instance with no registered handlers.
func NewServer() *Server {
	server := &Server{Idgen: types.RandomIDGenerator(), codecs: mapset.NewSet(), run: 1}
	// Register the default service providing meta information about the RPC service such
	// as the services and methods it offers.
	Service := &Service{server}
	err := server.RegisterName(MetadataAPI, Service)
	if err != nil {
		log.Trace("register name for server failed", "err", err)
	}
	return server
}

// ServeHTTP serves JSON-RPC requests over HTTP.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Permit dumb empty requests for remote health-checks (AWS)
	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if code, err := ValidateRequest(r); err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	// Create request-scoped context.
	connInfo := rpcCodec.PeerInfo{Transport: "http", RemoteAddr: r.RemoteAddr}
	connInfo.HTTP.Version = r.Proto
	connInfo.HTTP.Host = r.Host
	connInfo.HTTP.Origin = r.Header.Get("Origin")
	connInfo.HTTP.UserAgent = r.Header.Get("User-Agent")
	ctx := r.Context()
	ctx = context.WithValue(ctx, rpcCodec.PeerInfoContextKey{}, connInfo)

	// All checks passed, create a codec that reads directly from the request body
	// until EOF, writes the response to w, and orders the server to process a
	// single request.
	w.Header().Set("content-type", types.ContentType)
	codec := client.NewHTTPServerConn(r, w)
	defer codec.Close()
	s.serveSingleRequest(ctx, codec)
}

// WithMetrics enables metrics collection for server
func (s *Server) WithMetrics(r metrics.Registry) {
	client.EnableMetrics(r)
}

// RegisterName creates a service for the given receiver type under the given name. When no
// methods on the given receiver match the criteria to be either a RPC method or a
// subscription an error is returned. Otherwise a new service is created and added to the
// service collection this server provides to clients.
func (s *Server) RegisterName(name string, receiver interface{}) error {
	return s.services.RegisterName(name, receiver)
}

// ServeCodec reads incoming requests from codec, calls the appropriate callback and writes
// the response back using the given codec. It will block until the codec is closed or the
// server is stopped. In either case the codec is closed.
//
// Note that codec options are no longer supported.
func (s *Server) ServeCodec(codec rpcCodec.ServerCodec, options CodecOption) {
	defer codec.Close()

	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	// Add the codec to the set so it can be closed by Stop.
	s.codecs.Add(codec)
	defer s.codecs.Remove(codec)

	c := client.InitClient(codec, s.Idgen, &s.services)
	<-codec.Closed()
	c.Close()
}

// serveSingleRequest reads and processes a single RPC request from the given codec. This
// is used to serve HTTP connections. Subscriptions and reverse calls are not allowed in
// this mode.
func (s *Server) serveSingleRequest(ctx context.Context, codec rpcCodec.ServerCodec) {
	// Don't serve if server is stopped.
	if atomic.LoadInt32(&s.run) == 0 {
		return
	}

	h := client.NewHandler(ctx, codec, s.Idgen, &s.services)
	h.AllowSubscribe = false
	defer h.Close(io.EOF, nil)

	reqs, batch, err := codec.ReadBatch()
	if err != nil {
		if err != io.EOF {
			err := codec.WriteJSON(ctx, rpcCodec.ErrorMessage(&types.InvalidMessageError{Message: "parse error"}))
			if err != nil {
				log.Trace("write json for codec failed", "err", err)
			}
		}
		return
	}
	if batch {
		h.HandleBatch(reqs)
	} else {
		h.HandleMsg(reqs[0])
	}
}

// Stop stops reading new requests, waits for stopPendingRequestTimeout to allow pending
// requests to finish, then closes all codecs which will cancel pending requests and
// subscriptions.
func (s *Server) Stop() {
	if atomic.CompareAndSwapInt32(&s.run, 1, 0) {
		log.Debug("RPC server shutting down")
		s.codecs.Each(func(c interface{}) bool {
			c.(rpcCodec.ServerCodec).Close()
			return true
		})
	}
}

// Service gives meta information about the server.
// e.g. gives information about the loaded modules.
type Service struct {
	server *Server
}

// Modules returns the list of RPC services with their version number
func (s *Service) Modules() map[string]string {
	s.server.services.Mu.Lock()
	defer s.server.services.Mu.Unlock()

	modules := make(map[string]string)

	serviceNames := make([]string, 0, len(s.server.services.Services))
	for name := range s.server.services.Services {
		serviceNames = append(serviceNames, name)
	}
	for _, name := range serviceNames {
		modules[name] = "1.0"
	}
	return modules
}

// validateRequest returns a non-zero response code and error message if the
// request is invalid.
func ValidateRequest(r *http.Request) (int, error) {
	if r.Method == http.MethodPut || r.Method == http.MethodDelete {
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	}
	if r.ContentLength > types.MaxRequestContentLength {
		err := fmt.Errorf("content length too large (%d>%d)", r.ContentLength, types.MaxRequestContentLength)
		return http.StatusRequestEntityTooLarge, err
	}
	// Allow OPTIONS (regardless of content-type)
	if r.Method == http.MethodOptions {
		return 0, nil
	}
	// Check content-type
	if mt, _, err := mime.ParseMediaType(r.Header.Get("content-type")); err == nil {
		for _, accepted := range acceptedContentTypes {
			if accepted == mt {
				return 0, nil
			}
		}
	}
	// Invalid content-type
	err := fmt.Errorf("invalid content type, only %s is supported", types.ContentType)
	return http.StatusUnsupportedMediaType, err
}

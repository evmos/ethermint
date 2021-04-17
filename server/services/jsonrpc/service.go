package jsonrpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/cosmos/ethermint/server/config"
	"github.com/ethereum/go-ethereum/rpc"
)

type Service struct {
	rpcServer *rpc.Server
	apis      []rpc.API
	http      *http.Server
}

// NewService creates a new JSON-RPC server instance over http with public Ethereum APIs
func NewService(apis []rpc.API) *Service {
	s := &Service{
		rpcServer: rpc.NewServer(),
		apis:      apis,
		http:      &http.Server{},
	}
	s.http.Handler = s.rpcServer

	return s
}

// Name returns the JSON-RPC service name
func (Service) Name() string {
	return "JSON-RPC"
}

// RegisterRoutes registers the JSON-RPC server to the application. It fails if any of the
// API names fail to register.
func (s *Service) RegisterRoutes() error {
	for _, api := range s.apis {
		if err := s.rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			return err
		}
	}

	return nil
}

// Start starts the JSON-RPC server on the address defined on the configuration.
func (s *Service) Start(cfg config.Config) error {
	u, err := url.Parse(cfg.JSONRPC.Address)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", u.Host)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		err = s.http.Serve(listener)
		if err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(5 * time.Second): // assume server started successfully
		return nil
	}
}

// Stop stops the JSON-RPC service by no longer reading new requests, waits for
// stopPendingRequestTimeout to allow pending requests to finish, then closes all codecs which will
// cancel pending requests and subscriptions.
func (s *Service) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.http.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

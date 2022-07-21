package mocks

/*
package grpc returns a GRPC query client implementation that
accepts various (mock) implementations of the various methods.
This implementation is useful for using in tests, when you don't
need a real server, but want a high-level of control about
the server response you want to mock (eg. error handling),
or if you just want to record the calls to verify in your tests.

E.g. JSON-PRC-CLIENT -> BACKEND -> GRPC CLIENT -> APP
*/

import (
	"testing"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	rpc "github.com/evmos/ethermint/rpc/types"
)

// QueryClient defines a mocked object that implements the grpc QueryCLient
// interface. It's used on tests to test the JSON-RPC without running a grpc
// client server.
var _ evmtypes.QueryClient = &QueryClient{}

func TestQueryClient(t *testing.T) {
	queryClient := NewQueryClient(t)

	var header metadata.MD
	queryClient.On("Params", rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(&header)).Return(&evmtypes.QueryParamsResponse{}, nil)

	// mock calls for abstraction
	_, err := queryClient.Params(rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	require.NoError(t, err)
}

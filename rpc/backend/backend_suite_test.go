package backend

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	rpc "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

type BackendTestSuite struct {
	suite.Suite

	backend *Backend
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}

// SetupTest is executed before every Test
func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})
	clientCtx := client.Context{}.WithChainID("ethermint_9000-1").WithHeight(1)
	allowUnprotectedTxs := false

	suite.backend = NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)

	queryClient := NewMockQueryClient(suite.T())
	var header metadata.MD
	RegisterMockQueries(queryClient, &header)

	suite.backend.queryClient.QueryClient = queryClient
	suite.backend.ctx = rpc.ContextWithHeight(1)
}

// QueryClient defines a mocked object that implements the grpc QueryCLient
// interface. It's used on tests to test the JSON-RPC without running a grpc
// client server.
//
// E.g. JSON-PRC-CLIENT -> BACKEND -> Mock GRPC CLIENT -> APP
var _ evmtypes.QueryClient = &MockQueryClient{}

// RegisterMockQueries registers the queries and their respective responses,
// so that they can be called in tests using the queryClient
func RegisterMockQueries(queryClient *MockQueryClient, header *metadata.MD) {
	queryClient.On("Params", rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(header)).
		Return(&evmtypes.QueryParamsResponse{}, nil).
		Run(func(args mock.Arguments) {
			// If Params call is successful, also update the header height
			arg := args.Get(2).(grpc.HeaderCallOption)
			h := metadata.MD{}
			h.Set(grpctypes.GRPCBlockHeightHeader, "1")
			*arg.HeaderAddr = h
		})
}

func TestQueryClient(t *testing.T) {
	queryClient := NewMockQueryClient(t)
	var header metadata.MD
	RegisterMockQueries(queryClient, &header)

	// mock calls for abstraction
	_, err := queryClient.Params(rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	require.NoError(t, err)
}

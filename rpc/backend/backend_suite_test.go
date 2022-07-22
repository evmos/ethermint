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

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpc "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/x/evm/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

type BackendTestSuite struct {
	suite.Suite
	backend *Backend
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}

// SetupTest is executed before every BackendTestSuite test
func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})
	clientCtx := client.Context{}.WithChainID("ethermint_9000-1").WithHeight(1)
	allowUnprotectedTxs := false

	suite.backend = NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)

	queryClient := mocks.NewQueryClient(suite.T())
	var header metadata.MD
	RegisterMockQueries(queryClient, &header)

	suite.backend.queryClient.QueryClient = queryClient
	suite.backend.ctx = rpc.ContextWithHeight(1)
}

// QueryClient defines a mocked object that implements the grpc QueryCLient
// interface. It's used on tests to test the JSON-RPC without running a grpc
// client server. E.g. JSON-PRC-CLIENT -> BACKEND -> Mock GRPC CLIENT -> APP
var _ evmtypes.QueryClient = &mocks.QueryClient{}

// RegisterMockQueries registers the queries and their respective responses,
// so that they can be called in tests using the queryClient
func RegisterMockQueries(queryClient *mocks.QueryClient, header *metadata.MD) {
	queryClient.On("Params", rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(header)).
		Return(&evmtypes.QueryParamsResponse{}, nil).
		Run(func(args mock.Arguments) {
			// If Params call is successful, also update the header height
			arg := args.Get(2).(grpc.HeaderCallOption)
			h := metadata.MD{}
			h.Set(grpctypes.GRPCBlockHeightHeader, "1")
			*arg.HeaderAddr = h
		})

	baseFee := sdk.NewInt(1)
	queryClient.On("BaseFee", rpc.ContextWithHeight(1), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{BaseFee: &baseFee}, nil)
	// Base fee not enabled
	queryClient.On("BaseFee", rpc.ContextWithHeight(0), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{}, nil)
	// Base fee returns error
	queryClient.On("BaseFee", rpc.ContextWithHeight(-1), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{}, types.ErrInvalidBaseFee)
}

func TestQueryClient(t *testing.T) {
	queryClient := mocks.NewQueryClient(t)
	var header metadata.MD
	RegisterMockQueries(queryClient, &header)

	// mock calls for abstraction
	_, err := queryClient.Params(rpc.ContextWithHeight(1), &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	require.NoError(t, err)
	_, err = queryClient.BaseFee(rpc.ContextWithHeight(1), &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)
	res, err := queryClient.BaseFee(rpc.ContextWithHeight(0), &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryBaseFeeResponse{}, res)
	_, err = queryClient.BaseFee(rpc.ContextWithHeight(-1), &evmtypes.QueryBaseFeeRequest{})
	require.Error(t, err)
}

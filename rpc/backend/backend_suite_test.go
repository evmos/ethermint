package backend

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"

	"github.com/evmos/ethermint/rpc/backend/mocks"
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

// SetupTest is executed before every BackendTestSuite test
func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})
	clientCtx := client.Context{}.WithChainID("ethermint_9000-1").WithHeight(1).WithOffline(true)
	allowUnprotectedTxs := false

	suite.backend = NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)
	suite.backend.queryClient.QueryClient = mocks.NewQueryClient(suite.T())
	suite.backend.clientCtx.Client = mocks.NewClient(suite.T())
	suite.backend.ctx = rpc.ContextWithHeight(1)
}

// Client defines a mocked object that implements the tendermint rpc CLient
// interface. It's used on tests to test the JSON-RPC without running a
// tendermint rpc client server. E.g. JSON-PRC-CLIENT -> BACKEND -> Mock GRPC
// CLIENT -> APP
var _ tmrpcclient.Client = &mocks.Client{}

func TestClient(t *testing.T) {
	client := mocks.NewClient(t)

	// Register the queries and their respective responses, so that they can be
	// called in tests using the client
	height := rpc.BlockNumber(1).Int64()
	RegisterBlockQueries(client, height)

	block := types.Block{Header: types.Header{Height: height}}
	res, err := client.Block(rpc.ContextWithHeight(height), &height)
	require.Equal(t, res, &tmrpctypes.ResultBlock{Block: &block})
	require.NoError(t, err)
}

func RegisterBlockQueries(client *mocks.Client, height int64) {
	block := types.Block{Header: types.Header{Height: height}}
	client.On("Block", rpc.ContextWithHeight(height), mock.AnythingOfType("*int64")).
		Return(&tmrpctypes.ResultBlock{Block: &block}, nil)
}

// QueryClient defines a mocked object that implements the grpc QueryCLient
// interface. It's used on tests to test the JSON-RPC without running a grpc
// client server. E.g. JSON-PRC-CLIENT -> BACKEND -> Mock GRPC CLIENT -> APP
var _ evmtypes.QueryClient = &mocks.QueryClient{}

func TestQueryClient(t *testing.T) {
	queryClient := mocks.NewQueryClient(t)
	var header metadata.MD

	// Register the queries and their respective responses, so that they can be
	// called in tests using the queryClient
	height := int64(1)
	RegisterParamsQueries(queryClient, &header, height)
	_, err := queryClient.Params(rpc.ContextWithHeight(height), &evmtypes.QueryParamsRequest{}, grpc.Header(&header))
	require.NoError(t, err)
	blockHeightHeader := header.Get(grpctypes.GRPCBlockHeightHeader)
	headerHeight, err := strconv.ParseInt(blockHeightHeader[0], 10, 64)
	require.Equal(t, height, headerHeight)

	RegisterBaseFeeQueries(queryClient)
	_, err = queryClient.BaseFee(rpc.ContextWithHeight(1), &evmtypes.QueryBaseFeeRequest{})
	require.NoError(t, err)
	res, err := queryClient.BaseFee(rpc.ContextWithHeight(0), &evmtypes.QueryBaseFeeRequest{})
	require.Equal(t, &evmtypes.QueryBaseFeeResponse{}, res)
	require.NoError(t, err)
	_, err = queryClient.BaseFee(rpc.ContextWithHeight(-1), &evmtypes.QueryBaseFeeRequest{})
	require.Error(t, err)
}

func RegisterParamsQueries(queryClient *mocks.QueryClient, header *metadata.MD, height int64) {
	queryClient.On("Params", rpc.ContextWithHeight(height), &evmtypes.QueryParamsRequest{}, grpc.Header(header)).
		Return(&evmtypes.QueryParamsResponse{}, nil).
		Run(func(args mock.Arguments) {
			// If Params call is successful, also update the header height
			arg := args.Get(2).(grpc.HeaderCallOption)
			h := metadata.MD{}
			h.Set(grpctypes.GRPCBlockHeightHeader, fmt.Sprint(height))
			*arg.HeaderAddr = h
		})
}

func RegisterParamsQueriesError(queryClient *mocks.QueryClient, header *metadata.MD, height int64) {
	queryClient.On("Params", rpc.ContextWithHeight(height), &evmtypes.QueryParamsRequest{}, grpc.Header(header)).
		Return(nil, sdkerrors.ErrInvalidRequest)
}

func RegisterBaseFeeQueries(queryClient *mocks.QueryClient) {
	baseFee := sdk.NewInt(1)
	queryClient.On("BaseFee", rpc.ContextWithHeight(1), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{BaseFee: &baseFee}, nil)
	// Base fee not enabled
	queryClient.On("BaseFee", rpc.ContextWithHeight(0), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{}, nil)
	// Base fee returns error
	queryClient.On("BaseFee", rpc.ContextWithHeight(-1), &evmtypes.QueryBaseFeeRequest{}).
		Return(&evmtypes.QueryBaseFeeResponse{}, evmtypes.ErrInvalidBaseFee)
}

package backend

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/mocks"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

type BackendTestSuite struct {
	suite.Suite

	backend *Backend
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}

// Setup Test runs automatically
func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})

	clientCtx := client.Context{}.WithChainID("ethermint_9000-1")

	allowUnprotectedTxs := false

	suite.backend = NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)

	queryClient := mocks.NewQueryClient(suite.T())
	queryClient.On("Params", context.Background(), &evmtypes.QueryParamsRequest{}, grpc.HeaderCallOption{}).Return(&evmtypes.QueryParamsResponse{}, nil)

	suite.backend.queryClient.QueryClient = queryClient
}

func (suite *BackendTestSuite) TestBlockNumber() {
	testCases := []struct {
		mame           string
		malleate       func()
		expBlockNumber hexutil.Uint64
		expPass        bool
	}{
		{
			"pass",
			func() {},
			hexutil.Uint64(0x9),
			true,
		},
	}
	for _, tc := range testCases {
		blockNumber, err := suite.backend.BlockNumber()

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(1, blockNumber)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

package backend_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/backend"
	"github.com/stretchr/testify/suite"
)

type BackendTestSuite struct {
	suite.Suite
	backend *backend.Backend
}

func TestBackendTestSuite(t *testing.T) {
	suite.Run(t, new(BackendTestSuite))
}

func (suite *BackendTestSuite) SetupTest() {
	ctx := server.NewDefaultContext()
	ctx.Viper.Set("telemetry.global-labels", []interface{}{})
	clientCtx := client.Context{
		Height:  1,
		ChainID: "ethermint_9000-1",
	}
	allowUnprotectedTxs := false
	suite.backend = backend.NewBackend(ctx, ctx.Logger, clientCtx, allowUnprotectedTxs)

	// queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	// types.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	// suite.queryClient = types.NewQueryClient(queryHelper)

}

func (suite *BackendTestSuite) TestBlockNumber() {
	testCases := []struct {
		mame           string
		malleate       func()
		expBlockNumber hexutil.Uint64
		expPass        bool
	}{
		{
			"test",
			func() {},
			hexutil.Uint64(0x00),
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest()

		blockNumber, err := suite.backend.BlockNumber()

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBlockNumber, blockNumber)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

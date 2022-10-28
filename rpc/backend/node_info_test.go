package backend

import (
	"fmt"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	ethermint "github.com/evmos/ethermint/types"
)

func (suite *BackendTestSuite) TestRPCMinGasPrice() {
	testCases := []struct {
		name           string
		registerMock   func()
		expMinGasPrice int64
		expPass        bool
	}{
		{
			"pass - default gas price",
			func() {
				//client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsWithoutHeaderError(queryClient, 1)
			},
			ethermint.DefaultGasPrice,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			minPrice := suite.backend.RPCMinGasPrice()
			if tc.expPass {
				suite.Require().Equal(tc.expMinGasPrice, minPrice)
			} else {
				suite.Require().NotEqual(tc.expMinGasPrice, minPrice)
			}
		})
	}
}

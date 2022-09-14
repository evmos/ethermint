package backend

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/tests"
)

func (suite *BackendTestSuite) TestGetCode() {

	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))
	contract := []byte("0xef616c92f3cfc9e92dc270d6acff9cea213cecc7020a76ee4395af09bdceb4837a1ebdb5735e11e7d3adb6104e0c3ac55180b4ddf5e54d022cc5e8837f6a4f971b")

	testCases := []struct {
		name          string
		addr          common.Address
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(common.Address)
		expPass       bool
		expCode       hexutil.Bytes
	}{
		{
			"pass",
			tests.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterCode(queryClient, addr, contract)
			},
			true,
			contract,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.registerMock(tc.addr)

			code, err := suite.backend.GetCode(tc.addr, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expCode, code)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

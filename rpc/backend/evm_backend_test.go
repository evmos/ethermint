package backend

import "github.com/ethereum/go-ethereum/common/hexutil"

func (suite *BackendTestSuite) TestBlockNumber() {
	testCases := []struct {
		mame           string
		malleate       func()
		expBlockNumber hexutil.Uint64
		expPass        bool
	}{
		{
			"pass",
			func() {
			},
			0x1,
			true,
		},
	}
	for _, tc := range testCases {
		blockNumber, err := suite.backend.BlockNumber()

		if tc.expPass {
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expBlockNumber, blockNumber)
		} else {
			suite.Require().Error(err)
		}
	}
}

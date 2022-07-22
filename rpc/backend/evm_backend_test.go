package backend

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/tendermint/tendermint/abci/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

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
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBlockNumber, blockNumber)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

// TODO add Tendermint rpc Client mock
// func (suite *BackendTestSuite) TestGetTendermintBlockByNumber() {
// 	testCases := []struct {
// 		mame        string
// 		malleate    func()
// 		blocknumber types.BlockNumber
// 		expBlock    *coretypes.ResultBlock
// 		expPass     bool
// 	}{
// 		{
// 			"pass",
// 			func() {},
// 			types.BlockNumber(1),
// 			nil,
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		block, err := suite.backend.GetTendermintBlockByNumber(tc.blocknumber)

// 		if tc.expPass {
// 			suite.Require().Nil(err)
// 			suite.Require().Equal(tc.expBlock, block)
// 		} else {
// 			suite.Require().NotNil(err)
// 		}
// 	}
// }

func (suite *BackendTestSuite) TestBlockBloom() {
	testCases := []struct {
		mame          string
		malleate      func()
		blockRes      *tmrpctypes.ResultBlockResults
		expBlockBloom ethtypes.Bloom
		expPass       bool
	}{
		{
			"fail - empty block result",
			func() {},
			&tmrpctypes.ResultBlockResults{},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - non block bloom event type",
			func() {},
			&tmrpctypes.ResultBlockResults{
				EndBlockEvents: []types.Event{{Type: evmtypes.EventTypeEthereumTx}},
			},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - nonblock bloom attribute key",
			func() {},
			&tmrpctypes.ResultBlockResults{
				EndBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
						Attributes: []types.EventAttribute{
							{Key: []byte(evmtypes.AttributeKeyEthereumTxHash)},
						},
					},
				},
			},
			ethtypes.Bloom{},
			false,
		},
		{
			"pass - nonblock bloom attribute key",
			func() {},
			&tmrpctypes.ResultBlockResults{
				EndBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
						Attributes: []types.EventAttribute{
							{Key: []byte(bAttributeKeyEthereumBloom)},
						},
					},
				},
			},
			ethtypes.Bloom{},
			true,
		},
	}
	for _, tc := range testCases {
		blockBloom, err := suite.backend.BlockBloom(tc.blockRes)

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBlockBloom, blockBloom)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

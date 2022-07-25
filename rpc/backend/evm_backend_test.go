package backend

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/abci/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/metadata"

	"github.com/evmos/ethermint/rpc/backend/mocks"
	ethrpc "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

func (suite *BackendTestSuite) TestBlockNumber() {

	testCases := []struct {
		mame                string
		registerMockQueries func()
		expBlockNumber      hexutil.Uint64
		expPass             bool
	}{
		{
			"pass - app state header height 1",
			func() {
				// Register mock queries
				height := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsQueries(queryClient, &header, int64(height))
			},
			0x1,
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries
		tc.registerMockQueries()

		blockNumber, err := suite.backend.BlockNumber()

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBlockNumber, blockNumber)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

func (suite *BackendTestSuite) TestGetTendermintBlockByNumber() {
	var block tmtypes.Block

	testCases := []struct {
		mame                string
		blocknumber         ethrpc.BlockNumber
		registerMockQueries func(ethrpc.BlockNumber)
		expPass             bool
	}{
		{
			"fail - blockNum < 0 with app state height error",
			ethrpc.BlockNumber(-1),
			func(blockNum ethrpc.BlockNumber) {
				// QueryClient.Params
				appHeight := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsQueriesError(queryClient, &header, appHeight)
			},
			false,
		},
		{
			"pass - blockNum < 0 with app state height >= 1",
			ethrpc.BlockNumber(-1),
			func(blockNum ethrpc.BlockNumber) {
				// QueryClient.Params
				appHeight := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsQueries(queryClient, &header, appHeight)

				// Client.Block
				tmHeight := appHeight
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockQueries(client, tmHeight)

				block = tmtypes.Block{Header: tmtypes.Header{Height: tmHeight}}
			},
			true,
		},
		{
			"pass - blockNum = 0 (defaults to blockNum = 1 due to a difference between tendermint heights and geth heights",
			ethrpc.BlockNumber(0),
			func(blockNum ethrpc.BlockNumber) {
				// Client.Block
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockQueries(client, height)

				block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
			},
			true,
		},
		{
			"pass - blockNum = 1",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				// Client.Block
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockQueries(client, height)

				block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
			},
			true,
		},
		// TODO why does the "x-cosmos-block-height" always have to be  "1"?
		// {
		// 	"pass - blockNumber > 1",
		// 	ethrpc.BlockNumber(5),
		// 	func(blockNum ethrpc.BlockNumber) {
		// 		// Client.Block
		// 		height := blockNum.Int64()
		// 		client := suite.backend.clientCtx.Client.(*mocks.Client)
		// 		RegisterBlockQueries(client, height)

		// 		block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries

		tc.registerMockQueries(tc.blocknumber)
		resultBlock, err := suite.backend.GetTendermintBlockByNumber(tc.blocknumber)

		if tc.expPass {
			expResultBlock := &tmrpctypes.ResultBlock{Block: &block}
			suite.Require().Nil(err)
			suite.Require().Equal(expResultBlock, resultBlock)
			suite.Require().Equal(expResultBlock.Block.Header.Height, resultBlock.Block.Header.Height)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

func (suite *BackendTestSuite) TestBlockBloom() {
	testCases := []struct {
		mame          string
		blockRes      *tmrpctypes.ResultBlockResults
		expBlockBloom ethtypes.Bloom
		expPass       bool
	}{
		{
			"fail - empty block result",
			&tmrpctypes.ResultBlockResults{},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - non block bloom event type",
			&tmrpctypes.ResultBlockResults{
				EndBlockEvents: []types.Event{{Type: evmtypes.EventTypeEthereumTx}},
			},
			ethtypes.Bloom{},
			false,
		},
		{
			"fail - nonblock bloom attribute key",
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

func (suite *BackendTestSuite) TestBaseFee() {
	// Register mock queries
	queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
	RegisterBaseFeeQueries(queryClient)

	baseFee := sdk.NewInt(1)

	testCases := []struct {
		mame       string
		blockRes   *tmrpctypes.ResultBlockResults
		expBaseFee *big.Int
		expPass    bool
	}{
		{
			"fail - grpc BaseFee error - ",
			// query client mock returns err for height -1
			&tmrpctypes.ResultBlockResults{Height: -1},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with non feeemarket block event",
			&tmrpctypes.ResultBlockResults{
				Height: -1,
				BeginBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with feeemarket block event",
			&tmrpctypes.ResultBlockResults{
				Height: -1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with feeemarket block event with wrong attribute value",
			&tmrpctypes.ResultBlockResults{
				Height: -1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
						Attributes: []types.EventAttribute{
							{Value: []byte{0x1}},
						},
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with feeemarket block event with baseFee attribute value",
			&tmrpctypes.ResultBlockResults{
				Height: -1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
						Attributes: []types.EventAttribute{
							{Value: []byte(baseFee.String())},
						},
					},
				},
			},
			baseFee.BigInt(),
			true,
		},
		{
			"fail - base fee or london fork not enabled",
			&tmrpctypes.ResultBlockResults{Height: 0},
			nil,
			true,
		},
		{
			"pass",
			&tmrpctypes.ResultBlockResults{Height: 1},
			baseFee.BigInt(),
			true,
		},
	}
	for _, tc := range testCases {
		baseFee, err := suite.backend.BaseFee(tc.blockRes)

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBaseFee, baseFee)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

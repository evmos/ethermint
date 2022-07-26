package backend

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/abci/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/metadata"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	ethrpc "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

func (suite *BackendTestSuite) TestBlockNumber() {
	testCases := []struct {
		name           string
		registerMock   func()
		expBlockNumber hexutil.Uint64
		expPass        bool
	}{
		{
			"fail - invalid block header height",
			func() {
				height := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsInvalidHeight(queryClient, &header, int64(height))
			},
			0x0,
			false,
		},
		{
			"fail - invalid block header",
			func() {
				height := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsInvalidHeader(queryClient, &header, int64(height))
			},
			0x0,
			false,
		},
		{
			"pass - app state header height 1",
			func() {
				height := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParams(queryClient, &header, int64(height))
			},
			0x1,
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries
		tc.registerMock()

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
		name         string
		blocknumber  ethrpc.BlockNumber
		registerMock func(ethrpc.BlockNumber)
		found        bool
		expPass      bool
	}{
		{
			"fail - client error",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, height)
			},
			false,
			false,
		},
		{
			"noop - block not found",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockNotFound(client, height)
			},
			false,
			true,
		},
		{
			"fail - blockNum < 0 with app state height error",
			ethrpc.BlockNumber(-1),
			func(_ ethrpc.BlockNumber) {
				appHeight := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParamsError(queryClient, &header, appHeight)
			},
			false,
			false,
		},
		{
			"pass - blockNum < 0 with app state height >= 1",
			ethrpc.BlockNumber(-1),
			func(blockNum ethrpc.BlockNumber) {
				appHeight := int64(1)
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterParams(queryClient, &header, appHeight)

				tmHeight := appHeight
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlock(client, tmHeight)

				block = tmtypes.Block{Header: tmtypes.Header{Height: tmHeight}}
			},
			true,
			true,
		},
		{
			"pass - blockNum = 0 (defaults to blockNum = 1 due to a difference between tendermint heights and geth heights",
			ethrpc.BlockNumber(0),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlock(client, height)

				block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
			},
			true,
			true,
		},
		{
			"pass - blockNum = 1",
			ethrpc.BlockNumber(1),
			func(blockNum ethrpc.BlockNumber) {
				height := blockNum.Int64()
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlock(client, height)

				block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
			},
			true,
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
		// 		RegisterBlock(client, height)

		// 		block = tmtypes.Block{Header: tmtypes.Header{Height: height}}
		// 	},
		// 	true,
		// },
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries

		tc.registerMock(tc.blocknumber)
		resultBlock, err := suite.backend.GetTendermintBlockByNumber(tc.blocknumber)

		if tc.expPass {
			suite.Require().Nil(err)

			if !tc.found {
				suite.Require().Nil(resultBlock)
			} else {
				expResultBlock := &tmrpctypes.ResultBlock{Block: &block}
				suite.Require().Equal(expResultBlock, resultBlock)
				suite.Require().Equal(expResultBlock.Block.Header.Height, resultBlock.Block.Header.Height)
			}
		} else {
			suite.Require().NotNil(err)
		}
	}
}

func (suite *BackendTestSuite) TestBlockBloom() {
	testCases := []struct {
		name          string
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

func (suite *BackendTestSuite) TestEthBlockFromTendermint() {
	testCases := []struct {
		name         string
		resBlock     *tmrpctypes.ResultBlock
		blockRes     *tmrpctypes.ResultBlockResults
		fullTx       bool
		registerMock func(sdk.Int, sdk.AccAddress, int64)
		expPass      bool
	}{
		{
			"pass - block without tx",
			&tmrpctypes.ResultBlock{
				Block: tmtypes.MakeBlock(
					1,
					[]tmtypes.Tx{},
					nil,
					nil,
				),
			},
			&tmrpctypes.ResultBlockResults{
				Height: 1,
				TxsResults: []*types.ResponseDeliverTx{
					{
						Code:    0,
						Log:     ExceedBlockGasLimitError,
						GasUsed: 0,
					},
				},
			},
			true,
			func(baseFee sdk.Int, validator sdk.AccAddress, height int64) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFee(queryClient, baseFee)
				RegisterValidatorAccount(queryClient, validator)

				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterConsensusParams(client, height)
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries

		// Setup mock with given values
		baseFee := sdk.NewInt(1)
		validator := sdk.AccAddress(tests.GenerateAddress().Bytes())
		height := int64(1)
		tc.registerMock(baseFee, validator, height)

		block, err := suite.backend.EthBlockFromTendermint(tc.resBlock, tc.blockRes, tc.fullTx)

		if tc.expPass {
			header := tc.resBlock.Block.Header
			gasLimit := int64(^uint32(0)) // for `MaxGas = -1` (DefaultConsensusParams)
			gasUsed := new(big.Int).SetUint64(uint64(tc.blockRes.TxsResults[0].GasUsed))

			root := common.Hash{}.Bytes()
			receipt := ethtypes.NewReceipt(root, false, gasUsed.Uint64())
			bloom := ethtypes.CreateBloom(ethtypes.Receipts{receipt})

			var transactionsRoot common.Hash
			if len(tc.resBlock.Block.Txs) == 0 {
				transactionsRoot = ethtypes.EmptyRootHash
			} else {
				transactionsRoot = common.BytesToHash(header.DataHash)
			}

			ethRPCTxs := []interface{}{} // TODO Change for tests with txs

			expBlock := map[string]interface{}{
				"number":           hexutil.Uint64(header.Height),
				"hash":             hexutil.Bytes(header.Hash()),
				"parentHash":       common.BytesToHash(header.LastBlockID.Hash.Bytes()),
				"nonce":            ethtypes.BlockNonce{},   // PoW specific
				"sha3Uncles":       ethtypes.EmptyUncleHash, // No uncles in Tendermint
				"logsBloom":        bloom,
				"stateRoot":        hexutil.Bytes(header.AppHash),
				"miner":            common.BytesToAddress(validator.Bytes()),
				"mixHash":          common.Hash{},
				"difficulty":       (*hexutil.Big)(big.NewInt(0)),
				"extraData":        "0x",
				"size":             hexutil.Uint64(tc.resBlock.Block.Size()),
				"gasLimit":         hexutil.Uint64(gasLimit), // Static gas limit
				"gasUsed":          (*hexutil.Big)(gasUsed),
				"timestamp":        hexutil.Uint64(header.Time.Unix()),
				"transactionsRoot": transactionsRoot,
				"receiptsRoot":     ethtypes.EmptyRootHash,

				"uncles":          []common.Hash{},
				"transactions":    ethRPCTxs,
				"totalDifficulty": (*hexutil.Big)(big.NewInt(0)),
				"baseFeePerGas":   (*hexutil.Big)(sdk.NewInt(1).BigInt()),
			}

			suite.Require().Equal(expBlock, block)
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *BackendTestSuite) TestBaseFee() {
	baseFee := sdk.NewInt(1)

	testCases := []struct {
		name         string
		blockRes     *tmrpctypes.ResultBlockResults
		registerMock func()
		expBaseFee   *big.Int
		expPass      bool
	}{
		{
			"fail - grpc BaseFee error",
			&tmrpctypes.ResultBlockResults{Height: 1},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeError(queryClient)
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with non feemarket block event",
			&tmrpctypes.ResultBlockResults{
				Height: 1,
				BeginBlockEvents: []types.Event{
					{
						Type: evmtypes.EventTypeBlockBloom,
					},
				},
			},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeError(queryClient)
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with feemarket block event",
			&tmrpctypes.ResultBlockResults{
				Height: 1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
					},
				},
			},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeError(queryClient)
			},
			nil,
			false,
		},
		{
			"fail - grpc BaseFee error - with feemarket block event with wrong attribute value",
			&tmrpctypes.ResultBlockResults{
				Height: 1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
						Attributes: []types.EventAttribute{
							{Value: []byte{0x1}},
						},
					},
				},
			},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeError(queryClient)
			},
			nil,
			false,
		},
		{
			"fail - grpc baseFee error - with feemarket block event with baseFee attribute value",
			&tmrpctypes.ResultBlockResults{
				Height: 1,
				BeginBlockEvents: []types.Event{
					{
						Type: feemarkettypes.EventTypeFeeMarket,
						Attributes: []types.EventAttribute{
							{Value: []byte(baseFee.String())},
						},
					},
				},
			},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeError(queryClient)
			},
			baseFee.BigInt(),
			true,
		},
		{
			"fail - base fee or london fork not enabled",
			&tmrpctypes.ResultBlockResults{Height: 1},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFeeDisabled(queryClient)
			},
			nil,
			true,
		},
		{
			"pass",
			&tmrpctypes.ResultBlockResults{Height: 1},
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterBaseFee(queryClient, baseFee)
			},
			baseFee.BigInt(),
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries
		tc.registerMock()

		baseFee, err := suite.backend.BaseFee(tc.blockRes)

		if tc.expPass {
			suite.Require().Nil(err)
			suite.Require().Equal(tc.expBaseFee, baseFee)
		} else {
			suite.Require().NotNil(err)
		}
	}
}

func (suite *BackendTestSuite) TestGetEthereumMsgsFromTendermintBlock() {
	msgEthereumTx := evmtypes.NewTx(
		big.NewInt(1),
		uint64(0),
		&common.Address{},
		big.NewInt(0),
		100000,
		big.NewInt(1),
		nil,
		nil,
		[]byte{},
		nil,
	)

	txBuilder := suite.backend.clientCtx.TxConfig.NewTxBuilder()
	ethSigner := ethtypes.LatestSignerForChainID(big.NewInt(1))

	address, priv := tests.NewAddrKey()
	privKey := priv.(*ethsecp256k1.PrivKey)

	// A valid msg should have empty `From`
	msgEthereumTx.From = address.Hex()

	err := msgEthereumTx.Sign(ethSigner, tests.NewSigner(privKey))
	suite.Require().NoError(err)

	err = txBuilder.SetMsgs(msgEthereumTx)
	suite.Require().NoError(err)

	bz, err := suite.backend.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	suite.Require().NoError(err)

	testCases := []struct {
		name     string
		resBlock *tmrpctypes.ResultBlock
		blockRes *tmrpctypes.ResultBlockResults
		expMsgs  []*evmtypes.MsgEthereumTx
	}{
		{
			// TODO understand the !TxSuccessOrExceedsBlockGasLimit check
			"tx in not included in block - result code 1",
			&tmrpctypes.ResultBlock{
				Block: tmtypes.MakeBlock(
					1,
					[]tmtypes.Tx{bz},
					nil,
					nil,
				),
			},
			&tmrpctypes.ResultBlockResults{
				TxsResults: []*types.ResponseDeliverTx{
					{
						Code: 1,
						// Log:  ExceedBlockGasLimitError,
					},
				},
			},
			[]*evmtypes.MsgEthereumTx(nil),
		},
		// TODO DEBUG why the resulting MsgEtherum TX has additional data V: ([]uint8) (len=1)
		// {
		// 	"pass",
		// 	&tmrpctypes.ResultBlock{
		// 		Block: tmtypes.MakeBlock(
		// 			1,
		// 			[]tmtypes.Tx{bz},
		// 			nil,
		// 			nil,
		// 		),
		// 	},
		// 	&tmrpctypes.ResultBlockResults{
		// 		TxsResults: []*types.ResponseDeliverTx{
		// 			{
		// 				Code: 0,
		// 				Log:  ExceedBlockGasLimitError,
		// 			},
		// 		},
		// 	},
		// 	[]*evmtypes.MsgEthereumTx{msgEthereumTx},
		// },
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset test and queries

		msgs := suite.backend.GetEthereumMsgsFromTendermintBlock(tc.resBlock, tc.blockRes)
		suite.Require().Equal(tc.expMsgs, msgs)
	}
}

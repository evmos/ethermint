package backend

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/indexer"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func (suite *BackendTestSuite) TestTraceTransaction() {
	msgEthereumTx, _ := suite.buildEthereumTx()
	msgEthereumTx2, _ := suite.buildEthereumTx()

	txHash := msgEthereumTx.AsTransaction().Hash()
	txHash2 := msgEthereumTx2.AsTransaction().Hash()

	priv, _ := ethsecp256k1.GenerateKey()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())

	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	RegisterParamsWithoutHeader(queryClient, 1)

	armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
	suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
	ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())

	txEncoder := suite.backend.clientCtx.TxConfig.TxEncoder()

	msgEthereumTx.From = from.String()
	msgEthereumTx.Sign(ethSigner, suite.signer)
	tx, _ := msgEthereumTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), "aphoton")
	txBz, _ := txEncoder(tx)

	msgEthereumTx2.From = from.String()
	msgEthereumTx2.Sign(ethSigner, suite.signer)
	tx2, _ := msgEthereumTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), "aphoton")
	txBz2, _ := txEncoder(tx2)

	testCases := []struct {
		name          string
		registerMock  func()
		block         *types.Block
		responseBlock []*abci.ResponseDeliverTx
		expResult     interface{}
		expPass       bool
	}{
		{
			"fail - tx not found",
			func() {},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
							{Key: []byte("txIndex"), Value: []byte("0")},
							{Key: []byte("amount"), Value: []byte("1000")},
							{Key: []byte("txGasUsed"), Value: []byte("21000")},
							{Key: []byte("txHash"), Value: []byte("")},
							{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						}},
					},
				},
			},
			nil,
			false,
		},
		{
			"fail - block not found",
			func() {
				//var header metadata.MD
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, 1)
			},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
							{Key: []byte("txIndex"), Value: []byte("0")},
							{Key: []byte("amount"), Value: []byte("1000")},
							{Key: []byte("txGasUsed"), Value: []byte("21000")},
							{Key: []byte("txHash"), Value: []byte("")},
							{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			false,
		},
		{
			"pass - transaction found in a block with multiple transactions",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockMultipleTxs(client, 1, []types.Tx{txBz, txBz2})
				RegisterTraceTransactionWithPredecessors(queryClient, msgEthereumTx, []*evmtypes.MsgEthereumTx{msgEthereumTx})
			},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz, txBz2}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
							{Key: []byte("txIndex"), Value: []byte("0")},
							{Key: []byte("amount"), Value: []byte("1000")},
							{Key: []byte("txGasUsed"), Value: []byte("21000")},
							{Key: []byte("txHash"), Value: []byte("")},
							{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						}},
					},
				},
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: []byte("ethereumTxHash"), Value: []byte(txHash2.Hex())},
							{Key: []byte("txIndex"), Value: []byte("1")},
							{Key: []byte("amount"), Value: []byte("1000")},
							{Key: []byte("txGasUsed"), Value: []byte("21000")},
							{Key: []byte("txHash"), Value: []byte("")},
							{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
		{
			"pass - transaction found",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlock(client, 1, txBz)
				RegisterTraceTransaction(queryClient, msgEthereumTx)
			},
			&types.Block{Header: types.Header{Height: 1}, Data: types.Data{Txs: []types.Tx{txBz}}},
			[]*abci.ResponseDeliverTx{
				{
					Code: 0,
					Events: []abci.Event{
						{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
							{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
							{Key: []byte("txIndex"), Value: []byte("0")},
							{Key: []byte("amount"), Value: []byte("1000")},
							{Key: []byte("txGasUsed"), Value: []byte("21000")},
							{Key: []byte("txHash"), Value: []byte("")},
							{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						}},
					},
				},
			},
			map[string]interface{}{"test": "hello"},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			db := dbm.NewMemDB()
			suite.backend.indexer = indexer.NewKVIndexer(db, tmlog.NewNopLogger(), suite.backend.clientCtx)

			err := suite.backend.indexer.IndexBlock(tc.block, tc.responseBlock)
			txResult, err := suite.backend.TraceTransaction(txHash, nil)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expResult, txResult)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestTraceBlock() {
	msgEthTx, bz := suite.buildEthereumTx()
	emptyBlock := tmtypes.MakeBlock(1, []tmtypes.Tx{}, nil, nil)
	filledBlock := tmtypes.MakeBlock(1, []tmtypes.Tx{bz}, nil, nil)
	resBlockEmpty := tmrpctypes.ResultBlock{Block: emptyBlock, BlockID: emptyBlock.LastBlockID}
	resBlockFilled := tmrpctypes.ResultBlock{Block: filledBlock, BlockID: filledBlock.LastBlockID}

	testCases := []struct {
		name            string
		registerMock    func()
		expTraceResults []*evmtypes.TxTraceResult
		resBlock        *tmrpctypes.ResultBlock
		config          *evmtypes.TraceConfig
		expPass         bool
	}{
		{
			"pass - no transaction returning empty array",
			func() {},
			[]*evmtypes.TxTraceResult{},
			&resBlockEmpty,
			&evmtypes.TraceConfig{},
			true,
		},
		{
			"fail - cannot unmarshal data",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterTraceBlock(queryClient, []*evmtypes.MsgEthereumTx{msgEthTx})

			},
			[]*evmtypes.TxTraceResult{},
			&resBlockFilled,
			&evmtypes.TraceConfig{},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			traceResults, err := suite.backend.TraceBlock(1, tc.config, tc.resBlock)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expTraceResults, traceResults)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

package backend

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/indexer"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/tests"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc/metadata"
	"math/big"
)

func (suite *BackendTestSuite) TestGetTransactionByHash() {
	msgEthereumTx, bz := suite.buildEthereumTx()
	rpcTransaction, _ := rpctypes.NewRPCTransaction(msgEthereumTx.AsTransaction(), common.Hash{}, 0, 0, big.NewInt(1))

	testCases := []struct {
		name         string
		registerMock func()
		tx           *evmtypes.MsgEthereumTx
		expRPCTx     *rpctypes.RPCTransaction
		expPass      bool
	}{
		{
			"pass - Transaction not found, register unconfirmed transaction error",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterUnconfirmedTxsError(client, nil)
			},
			msgEthereumTx,
			nil,
			true,
		},
		{
			"pass - Transaction not found, empty unconfirmed transaction",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterUnconfirmedTxsEmpty(client, nil)
			},
			msgEthereumTx,
			nil,
			true,
		},
		{
			"pass - Gets the pending transaction by hash with from the EVMTxIndexer",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsWithoutHeader(queryClient, 1)
				ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())
				msgEthereumTx.From = suite.acc.String()
				msgEthereumTx.Sign(ethSigner, suite.signer)
				rlpEncodedBz, _ := rlp.EncodeToBytes(msgEthereumTx.AsTransaction())
				suite.backend.SendRawTransaction(rlpEncodedBz)
				RegisterUnconfirmedTxs(client, nil, []types.Tx{bz})
			},
			msgEthereumTx,
			rpcTransaction,
			true,
		},
		{
			"pass - ethereum transaction not found in msgs at block 0, index 0",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsWithoutHeader(queryClient, 1)
				suite.backend.indexer = nil
				ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())
				msgEthereumTx.From = suite.acc.String()
				msgEthereumTx.Sign(ethSigner, suite.signer)
				rlpEncodedBz, _ := rlp.EncodeToBytes(msgEthereumTx.AsTransaction())
				suite.backend.SendRawTransaction(rlpEncodedBz)
				RegisterUnconfirmedTxs(client, nil, []types.Tx{bz})
				query := fmt.Sprintf("%s.%s='%s'", evmtypes.TypeMsgEthereumTx, evmtypes.AttributeKeyEthereumTxHash, msgEthereumTx.Hash)
				RegisterTxSearch(client, query, bz)

			},
			msgEthereumTx,
			rpcTransaction,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			rpcTx, err := suite.backend.GetTransactionByHash(common.HexToHash(tc.tx.Hash))

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rpcTx, tc.expRPCTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionByBlockHashAndIndex() {
	_, bz := suite.buildEthereumTx()

	testCases := []struct {
		name         string
		registerMock func()
		blockHash    common.Hash
		expRPCTx     *rpctypes.RPCTransaction
		expPass      bool
	}{
		{
			"pass - block not found",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockByHashError(client, common.Hash{}, bz)
			},
			common.Hash{},
			nil,
			true,
		},
		{
			"pass - Block results error",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockByHash(client, common.Hash{}, bz)
				RegisterBlockResultsError(client, 1)
			},
			common.Hash{},
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			rpcTx, err := suite.backend.GetTransactionByBlockHashAndIndex(tc.blockHash, 1)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rpcTx, tc.expRPCTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionByBlockAndIndex() {
	msgEthTx, bz := suite.buildEthereumTx()
	defaultBlock := types.MakeBlock(1, []types.Tx{bz}, nil, nil)
	txFromMsg, _ := rpctypes.NewTransactionFromMsg(
		msgEthTx,
		common.BytesToHash(defaultBlock.Hash().Bytes()),
		1,
		0,
		big.NewInt(1),
	)
	testCases := []struct {
		name         string
		registerMock func()
		block        *tmrpctypes.ResultBlock
		idx          hexutil.Uint
		expRPCTx     *rpctypes.RPCTransaction
		expPass      bool
	}{
		{
			"pass - block txs index out of bound ",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockResults(client, 1)
			},
			&tmrpctypes.ResultBlock{Block: types.MakeBlock(1, []types.Tx{bz}, nil, nil)},
			1,
			nil,
			true,
		},
		{
			"pass - Can't fetch base fee",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockResults(client, 1)
				RegisterBaseFeeError(queryClient)
			},
			&tmrpctypes.ResultBlock{Block: defaultBlock},
			0,
			txFromMsg,
			true,
		},
		{
			"pass - returns the Ethereum format transaction by the Ethereum hash",
			func() {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, sdk.NewInt(1))
			},
			&tmrpctypes.ResultBlock{Block: defaultBlock},
			0,
			txFromMsg,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			rpcTx, err := suite.backend.GetTransactionByBlockAndIndex(tc.block, tc.idx)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rpcTx, tc.expRPCTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionByBlockNumberAndIndex() {
	msgEthTx, bz := suite.buildEthereumTx()
	defaultBlock := types.MakeBlock(1, []types.Tx{bz}, nil, nil)
	txFromMsg, _ := rpctypes.NewTransactionFromMsg(
		msgEthTx,
		common.BytesToHash(defaultBlock.Hash().Bytes()),
		1,
		0,
		big.NewInt(1),
	)
	testCases := []struct {
		name         string
		registerMock func()
		blockNum     rpctypes.BlockNumber
		idx          hexutil.Uint
		expRPCTx     *rpctypes.RPCTransaction
		expPass      bool
	}{
		{
			"fail -  block not found return nil",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlockError(client, 1)
			},
			0,
			0,
			nil,
			true,
		},
		{
			"pass - returns the transaction identified by block number and index",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterBlock(client, 1, bz)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, sdk.NewInt(1))
			},
			0,
			0,
			txFromMsg,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			rpcTx, err := suite.backend.GetTransactionByBlockNumberAndIndex(tc.blockNum, tc.idx)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rpcTx, tc.expRPCTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionByTxIndex() {
	_, bz := suite.buildEthereumTx()

	testCases := []struct {
		name         string
		registerMock func()
		height       int64
		index        uint
		expTxResult  *ethermint.TxResult
		expPass      bool
	}{
		{
			"fail - Ethereum tx with query not found",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				suite.backend.indexer = nil
				RegisterTxSearch(client, "tx.height=0 AND ethereum_tx.txIndex=0", bz)
			},
			0,
			0,
			&ethermint.TxResult{},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			txResults, err := suite.backend.GetTxByTxIndex(tc.height, tc.index)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(txResults, tc.expTxResult)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestQueryTendermintTxIndexer() {
	testCases := []struct {
		name         string
		registerMock func()
		txGetter     func(*rpctypes.ParsedTxs) *rpctypes.ParsedTx
		query        string
		expTxResult  *ethermint.TxResult
		expPass      bool
	}{
		{
			"fail - Ethereum tx with query not found",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterTxSearchEmpty(client, "")
			},
			func(txs *rpctypes.ParsedTxs) *rpctypes.ParsedTx {
				return &rpctypes.ParsedTx{}
			},
			"",
			&ethermint.TxResult{},
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			txResults, err := suite.backend.queryTendermintTxIndexer(tc.query, tc.txGetter)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(txResults, tc.expTxResult)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetTransactionReceipt() {
	msgEthereumTx, _ := suite.buildEthereumTx()
	txHash := msgEthereumTx.AsTransaction().Hash()

	priv, _ := ethsecp256k1.GenerateKey()
	from := common.BytesToAddress(priv.PubKey().Address().Bytes())
	signer := tests.NewSigner(priv)

	queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
	RegisterParamsWithoutHeader(queryClient, 1)
	armor := crypto.EncryptArmorPrivKey(priv, "", "eth_secp256k1")
	suite.backend.clientCtx.Keyring.ImportPrivKey("test_key", armor, "")
	ethSigner := ethtypes.LatestSigner(suite.backend.ChainConfig())
	msgEthereumTx.From = from.String()
	msgEthereumTx.Sign(ethSigner, signer)
	tx, _ := msgEthereumTx.BuildTx(suite.backend.clientCtx.TxConfig.NewTxBuilder(), "aphoton")
	txEncoder := suite.backend.clientCtx.TxConfig.TxEncoder()
	txBz, _ := txEncoder(tx)

	testCases := []struct {
		name         string
		registerMock func()
		tx           *evmtypes.MsgEthereumTx
		block        *types.Block
		blockResult  []*abci.ResponseDeliverTx
		expTxReceipt map[string]interface{}
		expPass      bool
	}{
		{
			"fail - Receipts do not match ",
			func() {
				var header metadata.MD
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterParams(queryClient, &header, 1)
				RegisterParamsWithoutHeader(queryClient, 1)
				RegisterBlock(client, 1, txBz)
				RegisterBlockResults(client, 1)

			},
			msgEthereumTx,
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
			map[string]interface{}(nil),
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			db := dbm.NewMemDB()
			suite.backend.indexer = indexer.NewKVIndexer(db, tmlog.NewNopLogger(), suite.backend.clientCtx)
			err := suite.backend.indexer.IndexBlock(tc.block, tc.blockResult)
			suite.backend.indexer.GetByTxHash(common.HexToHash(tc.tx.Hash))
			//suite.T().Log(result, err)

			txReceipt, err := suite.backend.GetTransactionReceipt(common.HexToHash(tc.tx.Hash))
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(txReceipt, tc.expTxReceipt)
			} else {
				suite.Require().NotEqual(txReceipt, tc.expTxReceipt)
				//suite.Require().Error(err)
			}
		})
	}
}

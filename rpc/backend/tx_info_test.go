package backend

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/tendermint/tendermint/types"
)

func (suite *BackendTestSuite) TestGetTransactionByHash() {
	msgEthereumTx, bz := suite.buildEthereumTx()

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
			"pass - ",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				RegisterParamsWithoutHeader(queryClient, 1)
				signedBz, err := suite.backend.Sign(common.BytesToAddress(suite.acc.Bytes()), bz)
				signer := ethtypes.LatestSigner(suite.backend.ChainConfig())
				msgEthereumTx.From = common.BytesToAddress(suite.acc.Bytes()).String()
				signErr := msgEthereumTx.Sign(signer, suite.backend.clientCtx.Keyring)
				rlpEncodedBz, _ := rlp.EncodeToBytes(msgEthereumTx.AsTransaction())
				hash, err := suite.backend.SendRawTransaction(rlpEncodedBz)
				suite.T().Log("err", err, hash, signedBz, signErr)
				RegisterUnconfirmedTxs(client, nil, []types.Tx{bz})
			},
			msgEthereumTx,
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			tc.registerMock()

			rpcTx, err := suite.backend.GetTransactionByHash(common.HexToHash(tc.tx.Hash))

			//suite.T().Log("rpcTx", rpcTx)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(rpcTx, tc.expRPCTx)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

//func (suite *BackendTestSuite) TestGetTransactionByHashPending() {
//	// TODO
//}
//
//func (suite *BackendTestSuite) TestGetTransactionReceipt() {
//	// TODO
//}

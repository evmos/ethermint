package keeper_test

import (
	"fmt"

	"google.golang.org/grpc/metadata"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

func (suite *KeeperTestSuite) TestQueryAccount() {
	var (
		req        *types.QueryAccountRequest
		expAccount *types.QueryAccountResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"zero address",
			func() {
				suite.app.BankKeeper.SetBalance(suite.ctx, suite.address.Bytes(), ethermint.NewPhotonCoinInt64(0))
				expAccount = &types.QueryAccountResponse{
					Balance:  "0",
					CodeHash: ethcrypto.Keccak256(nil),
					Nonce:    0,
				}
				req = &types.QueryAccountRequest{
					Address: ethcmn.Address{}.String(),
				}
			},
			false,
		},
		{
			"success",
			func() {
				suite.app.BankKeeper.SetBalance(suite.ctx, suite.address.Bytes(), ethermint.NewPhotonCoinInt64(100))
				expAccount = &types.QueryAccountResponse{
					Balance:  "100",
					CodeHash: ethcrypto.Keccak256(nil),
					Nonce:    0,
				}
				req = &types.QueryAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.Account(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expAccount, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCosmosAccount() {
	var (
		req        *types.QueryCosmosAccountRequest
		expAccount *types.QueryCosmosAccountResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"zero address",
			func() {
				suite.app.BankKeeper.SetBalance(suite.ctx, suite.address.Bytes(), ethermint.NewPhotonCoinInt64(0))
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(ethcmn.Address{}.Bytes()).String(),
				}
				req = &types.QueryCosmosAccountRequest{
					Address: ethcmn.Address{}.String(),
				}
			},
			false,
		},
		{
			"success",
			func() {
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:      0,
					AccountNumber: 0,
				}
				req = &types.QueryCosmosAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
		{
			"success with seq and account number",
			func() {
				acc := suite.app.AccountKeeper.GetAccount(suite.ctx, suite.address.Bytes())
				suite.Require().NoError(acc.SetSequence(10))
				suite.Require().NoError(acc.SetAccountNumber(1))
				suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:      10,
					AccountNumber: 1,
				}
				req = &types.QueryCosmosAccountRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.CosmosAccount(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expAccount, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryBalance() {
	var (
		req        *types.QueryBalanceRequest
		expBalance string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"zero address",
			func() {
				suite.app.BankKeeper.SetBalance(suite.ctx, suite.address.Bytes(), ethermint.NewPhotonCoinInt64(0))
				expBalance = "0"
				req = &types.QueryBalanceRequest{
					Address: ethcmn.Address{}.String(),
				}
			},
			false,
		},
		{
			"success",
			func() {
				suite.app.BankKeeper.SetBalance(suite.ctx, suite.address.Bytes(), ethermint.NewPhotonCoinInt64(100))
				expBalance = "100"
				req = &types.QueryBalanceRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.Balance(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expBalance, res.Balance)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryStorage() {
	var (
		req      *types.QueryStorageRequest
		expValue string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"zero address",
			func() {
				req = &types.QueryStorageRequest{
					Address: ethcmn.Address{}.String(),
				}
			},
			false,
		},
		{"empty hash",
			func() {
				req = &types.QueryStorageRequest{
					Address: suite.address.String(),
					Key:     ethcmn.Hash{}.String(),
				}
			},
			false,
		},
		{
			"success",
			func() {
				key := ethcmn.BytesToHash([]byte("key"))
				value := ethcmn.BytesToHash([]byte("value"))
				expValue = value.String()
				suite.app.EvmKeeper.CommitStateDB.SetState(suite.address, key, value)
				req = &types.QueryStorageRequest{
					Address: suite.address.String(),
					Key:     key.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.Storage(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expValue, res.Value)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryCode() {
	var (
		req     *types.QueryCodeRequest
		expCode []byte
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"zero address",
			func() {
				req = &types.QueryCodeRequest{
					Address: ethcmn.Address{}.String(),
				}
				exp := &types.QueryCodeResponse{}
				expCode = exp.Code
			},
			false,
		},
		{
			"success",
			func() {
				expCode = []byte("code")
				suite.app.EvmKeeper.CommitStateDB.SetCode(suite.address, expCode)

				req = &types.QueryCodeRequest{
					Address: suite.address.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.Code(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expCode, res.Code)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryTxLogs() {
	var (
		req     *types.QueryTxLogsRequest
		expLogs []*types.Log
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty hash",
			func() {
				req = &types.QueryTxLogsRequest{
					Hash: ethcmn.Hash{}.String(),
				}
			},
			false,
		},
		{"logs not found",
			func() {
				hash := ethcmn.BytesToHash([]byte("hash"))
				req = &types.QueryTxLogsRequest{
					Hash: hash.String(),
				}
			},
			true,
		},
		{
			"success",
			func() {
				hash := ethcmn.BytesToHash([]byte("tx_hash"))

				expLogs = []*types.Log{
					{
						Address:     suite.address.String(),
						Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      hash.String(),
						TxIndex:     1,
						BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
						Index:       0,
						Removed:     false,
					},
				}

				suite.app.EvmKeeper.CommitStateDB.SetLogs(hash, types.LogsToEthereum(expLogs))

				req = &types.QueryTxLogsRequest{
					Hash: hash.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.TxLogs(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expLogs, res.Logs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryBlockLogs() {
	var (
		req     *types.QueryBlockLogsRequest
		expLogs []types.TransactionLogs
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty hash",
			func() {
				req = &types.QueryBlockLogsRequest{
					Hash: ethcmn.Hash{}.String(),
				}
			},
			false,
		},
		{"logs not found",
			func() {
				hash := ethcmn.BytesToHash([]byte("hash"))
				req = &types.QueryBlockLogsRequest{
					Hash: hash.String(),
				}
			},
			true,
		},
		{
			"success",
			func() {

				hash := ethcmn.BytesToHash([]byte("block_hash"))
				expLogs = []types.TransactionLogs{
					{
						Hash: ethcmn.BytesToHash([]byte("tx_hash_0")).String(),
						Logs: []*types.Log{
							{
								Address:     suite.address.String(),
								Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
								Data:        []byte("data"),
								BlockNumber: 1,
								TxHash:      ethcmn.BytesToHash([]byte("tx_hash_0")).String(),
								TxIndex:     1,
								BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
								Index:       0,
								Removed:     false,
							},
						},
					},
					{
						Hash: ethcmn.BytesToHash([]byte("tx_hash_1")).String(),
						Logs: []*types.Log{
							{
								Address:     suite.address.String(),
								Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
								Data:        []byte("data"),
								BlockNumber: 1,
								TxHash:      ethcmn.BytesToHash([]byte("tx_hash_1")).String(),
								TxIndex:     1,
								BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
								Index:       0,
								Removed:     false,
							},
							{
								Address:     suite.address.String(),
								Topics:      []string{ethcmn.BytesToHash([]byte("topic_1")).String()},
								Data:        []byte("data_1"),
								BlockNumber: 1,
								TxHash:      ethcmn.BytesToHash([]byte("tx_hash_1")).String(),
								TxIndex:     1,
								BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
								Index:       0,
								Removed:     false,
							},
						},
					},
				}

				suite.app.EvmKeeper.CommitStateDB.SetLogs(ethcmn.BytesToHash([]byte("tx_hash_0")), types.LogsToEthereum(expLogs[0].Logs))
				suite.app.EvmKeeper.CommitStateDB.SetLogs(ethcmn.BytesToHash([]byte("tx_hash_1")), types.LogsToEthereum(expLogs[1].Logs))

				req = &types.QueryBlockLogsRequest{
					Hash: hash.String(),
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.BlockLogs(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expLogs, res.TxLogs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryTxReceipt() {
	var (
		req    *types.QueryTxReceiptRequest
		expRes *types.QueryTxReceiptResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty hash",
			func() {
				req = &types.QueryTxReceiptRequest{}
			},
			false,
		},
		{"tx receipt not found for hash",
			func() {
				hash := ethcmn.BytesToHash([]byte("thash"))
				req = &types.QueryTxReceiptRequest{
					Hash: hash.Hex(),
				}
			},
			false,
		},
		{"success",
			func() {
				hash := ethcmn.BytesToHash([]byte("thash"))
				receipt := &types.TxReceipt{
					Hash:        hash.Hex(),
					From:        suite.address.Hex(),
					BlockHeight: uint64(suite.ctx.BlockHeight()),
					BlockHash:   ethcmn.BytesToHash(suite.ctx.BlockHeader().DataHash).Hex(),
				}

				suite.app.EvmKeeper.SetTxReceiptToHash(suite.ctx, hash, receipt)
				req = &types.QueryTxReceiptRequest{
					Hash: hash.Hex(),
				}

				expRes = &types.QueryTxReceiptResponse{
					Receipt: receipt,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			res, err := suite.queryClient.TxReceipt(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryTxReceiptByBlockHeight() {
	var (
		req    = &types.QueryTxReceiptsByBlockHeightRequest{}
		expRes *types.QueryTxReceiptsByBlockHeightResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty response",
			func() {
				expRes = &types.QueryTxReceiptsByBlockHeightResponse{
					Receipts: nil,
				}
			},
			true,
		},
		{"success",
			func() {
				hash := ethcmn.BytesToHash([]byte("thash"))
				receipt := &types.TxReceipt{
					Hash:        hash.Hex(),
					From:        suite.address.Hex(),
					BlockHeight: uint64(suite.ctx.BlockHeight()),
					BlockHash:   ethcmn.BytesToHash(suite.ctx.BlockHeader().DataHash).Hex(),
				}

				suite.app.EvmKeeper.AddTxHashToBlock(suite.ctx, suite.ctx.BlockHeight(), hash)
				suite.app.EvmKeeper.SetTxReceiptToHash(suite.ctx, hash, receipt)
				expRes = &types.QueryTxReceiptsByBlockHeightResponse{
					Receipts: []*types.TxReceipt{receipt},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", suite.ctx.BlockHeight()))

			res, err := suite.queryClient.TxReceiptsByBlockHeight(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expRes, res)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryBlockBloom() {
	var (
		req      *types.QueryBlockBloomRequest
		expBloom []byte
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"marshal error",
			func() {},
			false,
		},
		{"bloom not found for height",
			func() {
				req = &types.QueryBlockBloomRequest{}
				bloom := ethtypes.BytesToBloom([]byte("bloom"))
				expBloom = bloom.Bytes()
				suite.ctx = suite.ctx.WithBlockHeight(10)
				suite.app.EvmKeeper.SetBlockBloom(suite.ctx, 2, bloom)
			},
			false,
		},
		{
			"success",
			func() {
				req = &types.QueryBlockBloomRequest{}
				bloom := ethtypes.BytesToBloom([]byte("bloom"))
				expBloom = bloom.Bytes()
				suite.ctx = suite.ctx.WithBlockHeight(1)
				suite.app.EvmKeeper.SetBlockBloom(suite.ctx, 1, bloom)
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)
			ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", suite.ctx.BlockHeight()))
			res, err := suite.queryClient.BlockBloom(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				suite.Require().Equal(expBloom, res.Bloom)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	expParams := types.DefaultParams()

	res, err := suite.queryClient.Params(ctx, &types.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(expParams, res.Params)
}

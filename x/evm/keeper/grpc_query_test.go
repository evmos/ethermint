package keeper_test

import (
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

//Not valid Ethereum address
const invalidAddress = "0x0000"

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
		{
			"invalid address",
			func() {
				expAccount = &types.QueryAccountResponse{
					Balance:  "0",
					CodeHash: common.BytesToHash(ethcrypto.Keccak256(nil)).Hex(),
					Nonce:    0,
				}
				req = &types.QueryAccountRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				amt := sdk.Coins{ethermint.NewPhotonCoinInt64(100)}
				err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
				suite.Require().NoError(err)

				expAccount = &types.QueryAccountResponse{
					Balance:  "100",
					CodeHash: common.BytesToHash(ethcrypto.Keccak256(nil)).Hex(),
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
		{"invalid address",
			func() {
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(ethcmn.Address{}.Bytes()).String(),
				}
				req = &types.QueryCosmosAccountRequest{
					Address: invalidAddress,
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
		{"invalid address",
			func() {
				expBalance = "0"
				req = &types.QueryBalanceRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				amt := sdk.Coins{ethermint.NewPhotonCoinInt64(100)}
				err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
				suite.Require().NoError(err)
				err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
				suite.Require().NoError(err)

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
		{"invalid address",
			func() {
				req = &types.QueryStorageRequest{
					Address: invalidAddress,
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
				suite.app.EvmKeeper.SetState(suite.address, key, value)
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
		{"invalid address",
			func() {
				req = &types.QueryCodeRequest{
					Address: invalidAddress,
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
				suite.app.EvmKeeper.SetCode(suite.address, expCode)

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

				suite.app.EvmKeeper.SetLogs(hash, types.LogsToEthereum(expLogs))

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

				suite.app.EvmKeeper.SetLogs(ethcmn.BytesToHash([]byte("tx_hash_0")), types.LogsToEthereum(expLogs[0].Logs))
				suite.app.EvmKeeper.SetLogs(ethcmn.BytesToHash([]byte("tx_hash_1")), types.LogsToEthereum(expLogs[1].Logs))

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

func (suite *KeeperTestSuite) TestQueryValidatorAccount() {
	var (
		req        *types.QueryValidatorAccountRequest
		expAccount *types.QueryValidatorAccountResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"invalid address",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(ethcmn.Address{}.Bytes()).String(),
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: "",
				}
			},
			false,
		},
		{
			"success",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:       0,
					AccountNumber:  0,
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: suite.consAddress.String(),
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

				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(suite.address.Bytes()).String(),
					Sequence:       10,
					AccountNumber:  1,
				}
				req = &types.QueryValidatorAccountRequest{
					ConsAddress: suite.consAddress.String(),
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
			res, err := suite.queryClient.ValidatorAccount(ctx, req)

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

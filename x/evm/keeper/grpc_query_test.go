package keeper_test

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

// Not valid Ethereum address
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
					CodeHash: common.BytesToHash(crypto.Keccak256(nil)).Hex(),
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
					CodeHash: common.BytesToHash(crypto.Keccak256(nil)).Hex(),
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
		{
			"invalid address",
			func() {
				expAccount = &types.QueryCosmosAccountResponse{
					CosmosAddress: sdk.AccAddress(common.Address{}.Bytes()).String(),
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
		{
			"invalid address",
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
		{
			"invalid address",
			func() {
				req = &types.QueryStorageRequest{
					Address: invalidAddress,
				}
			},
			false,
		},
		{
			"success",
			func() {
				key := common.BytesToHash([]byte("key"))
				value := common.BytesToHash([]byte("value"))
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
		{
			"invalid address",
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
		txHash  common.Hash
		expLogs []*types.Log
	)

	testCases := []struct {
		msg      string
		malleate func()
	}{
		{
			"empty logs",
			func() {
				txHash = common.BytesToHash([]byte("hash"))
				expLogs = nil
			},
		},
		{
			"success",
			func() {
				txHash = common.BytesToHash([]byte("tx_hash"))

				expLogs = []*types.Log{
					{
						Address:     suite.address.String(),
						Topics:      []string{common.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      txHash.String(),
						TxIndex:     1,
						BlockHash:   common.BytesToHash(suite.ctx.HeaderHash()).Hex(),
						Index:       0,
						Removed:     false,
					},
				}

				suite.app.EvmKeeper.SetTxHashTransient(txHash)
				suite.app.EvmKeeper.IncreaseTxIndexTransient()
				for _, log := range types.LogsToEthereum(expLogs) {
					suite.app.EvmKeeper.AddLog(log)
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			logs := suite.app.EvmKeeper.GetTxLogsTransient(txHash)
			suite.Require().Equal(expLogs, types.NewLogsFromEth(logs))
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
		{
			"invalid address",
			func() {
				expAccount = &types.QueryValidatorAccountResponse{
					AccountAddress: sdk.AccAddress(common.Address{}.Bytes()).String(),
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

func (suite *KeeperTestSuite) TestEstimateGas() {
	gasHelper := hexutil.Uint64(20000)

	var (
		args   types.TransactionArgs
		gasCap uint64
	)
	testCases := []struct {
		msg          string
		malleate     func()
		expPass      bool
		expGas       uint64
		dynamicTxFee bool
	}{
		// should success, because transfer value is zero
		{"default args", func() {
			args = types.TransactionArgs{To: &common.Address{}}
		}, true, 21000, false},
		// should fail, because the default From address(zero address) don't have fund
		{"not enough balance", func() {
			args = types.TransactionArgs{To: &common.Address{}, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0, false},
		// should success, enough balance now
		{"enough balance", func() {
			args = types.TransactionArgs{To: &common.Address{}, From: &suite.address, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0, false},
		// should success, because gas limit lower than 21000 is ignored
		{"gas exceed allowance", func() {
			args = types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
		}, true, 21000, false},
		// should fail, invalid gas cap
		{"gas exceed global allowance", func() {
			args = types.TransactionArgs{To: &common.Address{}}
			gasCap = 20000
		}, false, 0, false},
		// estimate gas of an erc20 contract deployment, the exact gas number is checked with geth
		{"contract deployment", func() {
			ctorArgs, err := types.ERC20Contract.ABI.Pack("", &suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Require().NoError(err)
			data := append(types.ERC20Contract.Bin, ctorArgs...)
			args = types.TransactionArgs{
				From: &suite.address,
				Data: (*hexutil.Bytes)(&data),
			}
		}, true, 1186778, false},
		// estimate gas of an erc20 transfer, the exact gas number is checked with geth
		{"erc20 transfer", func() {
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			transferData, err := types.ERC20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
			suite.Require().NoError(err)
			args = types.TransactionArgs{To: &contractAddr, From: &suite.address, Data: (*hexutil.Bytes)(&transferData)}
		}, true, 51880, false},

		// repeated tests with dynamicTxFee
		{"default args w/ dynamicTxFee", func() {
			args = types.TransactionArgs{To: &common.Address{}}
		}, true, 21000, true},
		{"not enough balance w/ dynamicTxFee", func() {
			args = types.TransactionArgs{To: &common.Address{}, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0, true},
		{"enough balance w/ dynamicTxFee", func() {
			args = types.TransactionArgs{To: &common.Address{}, From: &suite.address, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0, true},
		{"gas exceed allowance w/ dynamicTxFee", func() {
			args = types.TransactionArgs{To: &common.Address{}, Gas: &gasHelper}
		}, true, 21000, true},
		{"gas exceed global allowance w/ dynamicTxFee", func() {
			args = types.TransactionArgs{To: &common.Address{}}
			gasCap = 20000
		}, false, 0, true},
		{"contract deployment w/ dynamicTxFee", func() {
			ctorArgs, err := types.ERC20Contract.ABI.Pack("", &suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Require().NoError(err)
			data := append(types.ERC20Contract.Bin, ctorArgs...)
			args = types.TransactionArgs{
				From: &suite.address,
				Data: (*hexutil.Bytes)(&data),
			}
		}, true, 1186778, true},
		{"erc20 transfer w/ dynamicTxFee", func() {
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			transferData, err := types.ERC20Contract.ABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
			suite.Require().NoError(err)
			args = types.TransactionArgs{To: &contractAddr, From: &suite.address, Data: (*hexutil.Bytes)(&transferData)}
		}, true, 51880, true},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.dynamicTxFee = tc.dynamicTxFee
			suite.SetupTest()
			gasCap = 25_000_000
			tc.malleate()

			args, err := json.Marshal(&args)
			suite.Require().NoError(err)
			req := types.EthCallRequest{
				Args:   args,
				GasCap: gasCap,
			}

			rsp, err := suite.queryClient.EstimateGas(sdk.WrapSDKContext(suite.ctx), &req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expGas, rsp.Gas)
			} else {
				suite.Require().Error(err)
			}
		})
	}
	suite.dynamicTxFee = false // reset flag
}

func (suite *KeeperTestSuite) TestTraceTx() {
	// TODO deploy contract that triggers internal transactions
	var (
		txMsg        *types.MsgEthereumTx
		traceConfig  *types.TraceConfig
		txIndex      uint64
		predecessors []*types.MsgEthereumTx
	)

	testCases := []struct {
		msg           string
		malleate      func()
		expPass       bool
		traceResponse []byte
		dynamicTxFee  bool
	}{
		{
			msg: "default trace",
			malleate: func() {
				txIndex = 0
				traceConfig = nil
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: []byte{0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61},
		},
		{
			msg: "default trace with filtered response",
			malleate: func() {
				txIndex = 0
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: []byte{0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x33, 0x2c, 0x22, 0x67},
			dynamicTxFee:  false,
		}, {
			msg: "javascript tracer",
			malleate: func() {
				txIndex = 0
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x5d},
		},
		{
			msg: "default trace with dynamicTxFee",
			malleate: func() {
				txIndex = 0
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: []byte{0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x33, 0x2c, 0x22, 0x67},
			dynamicTxFee:  true,
		}, {
			msg: "javascript tracer with dynamicTxFee",
			malleate: func() {
				txIndex = 0
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
				predecessors = []*types.MsgEthereumTx{}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x5d},
			dynamicTxFee:  true,
		},
		{
			msg: "default tracer with predecessors",
			malleate: func() {
				txIndex = 1
				traceConfig = nil

				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				// Generate token transfer transaction
				firstTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
				txMsg = suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
				suite.Commit()

				predecessors = append(predecessors, firstTx)
			},
			expPass:       true,
			traceResponse: []byte{0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x39, 0x31, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73},
			dynamicTxFee:  false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.dynamicTxFee = tc.dynamicTxFee
			suite.SetupTest()
			// Deploy contract
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			// Generate token transfer transaction
			txMsg = suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
			suite.Commit()

			tc.malleate()
			traceReq := types.QueryTraceTxRequest{
				Msg:          txMsg,
				TraceConfig:  traceConfig,
				TxIndex:      txIndex,
				Predecessors: predecessors,
			}
			res, err := suite.queryClient.TraceTx(sdk.WrapSDKContext(suite.ctx), &traceReq)
			suite.Require().NoError(err)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, res.Data[:150])
				} else {
					suite.Require().Equal(tc.traceResponse, res.Data)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}

	suite.dynamicTxFee = false // reset flag
}

func (suite *KeeperTestSuite) TestTraceBlock() {
	var (
		txs         []*types.MsgEthereumTx
		traceConfig *types.TraceConfig
	)

	testCases := []struct {
		msg           string
		malleate      func()
		expPass       bool
		traceResponse []byte
		dynamicTxFee  bool
	}{
		{
			msg: "default trace",
			malleate: func() {
				traceConfig = nil
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a},
		},
		{
			msg: "filtered trace",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61},
		}, {
			msg: "javascript tracer",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d},
		},
		{
			msg: "default trace with dynamicTxFee and filtered return",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					DisableStack:   true,
					DisableStorage: true,
					EnableMemory:   false,
				}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61},
			dynamicTxFee:  true,
		}, {
			msg: "javascript tracer with dynamicTxFee",
			malleate: func() {
				traceConfig = &types.TraceConfig{
					Tracer: "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == \"CALL\") this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}",
				}
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x5d},
			dynamicTxFee:  true,
		},
		{
			msg: "tracer with multiple transactions",
			malleate: func() {
				traceConfig = nil
				contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
				suite.Commit()
				// create multiple transactions in the same block
				firstTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
				secondTx := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
				suite.Commit()
				// overwrite txs to include only the ones on new block
				txs = append([]*types.MsgEthereumTx{}, firstTx, secondTx)
			},
			expPass:       true,
			traceResponse: []byte{0x5b, 0x7b, 0x22, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x3a, 0x7b, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x34, 0x38, 0x32, 0x38, 0x2c, 0x22, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x22, 0x3a, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0x22, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3a, 0x22, 0x22, 0x2c, 0x22, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x4c, 0x6f, 0x67, 0x73, 0x22, 0x3a, 0x5b, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a, 0x22, 0x50, 0x55, 0x53, 0x48, 0x31, 0x22, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x22, 0x3a, 0x33, 0x30, 0x32, 0x39, 0x36, 0x2c, 0x22, 0x67, 0x61, 0x73, 0x43, 0x6f, 0x73, 0x74, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x64, 0x65, 0x70, 0x74, 0x68, 0x22, 0x3a, 0x31, 0x2c, 0x22, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x22, 0x3a, 0x5b, 0x5d, 0x7d, 0x2c, 0x7b, 0x22, 0x70, 0x63, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x6f, 0x70, 0x22, 0x3a},
			dynamicTxFee:  false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			txs = []*types.MsgEthereumTx{}
			suite.dynamicTxFee = tc.dynamicTxFee
			suite.SetupTest()
			// Deploy contract
			contractAddr := suite.DeployTestContract(suite.T(), suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			// Generate token transfer transaction
			txMsg := suite.TransferERC20Token(suite.T(), contractAddr, suite.address, common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), sdk.NewIntWithDecimal(1, 18).BigInt())
			suite.Commit()

			txs = append(txs, txMsg)

			tc.malleate()
			traceReq := types.QueryTraceBlockRequest{
				Txs:         txs,
				TraceConfig: traceConfig,
			}
			res, err := suite.queryClient.TraceBlock(sdk.WrapSDKContext(suite.ctx), &traceReq)

			if tc.expPass {
				suite.Require().NoError(err)
				// if data is to big, slice the result
				if len(res.Data) > 150 {
					suite.Require().Equal(tc.traceResponse, res.Data[:150])
				} else {
					suite.Require().Equal(tc.traceResponse, res.Data)
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}

	suite.dynamicTxFee = false // reset flag
}

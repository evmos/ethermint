package keeper_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"

	"google.golang.org/grpc/metadata"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

//Not valid Ethereum address
const invalidAddress = "0x0000"

var (
	//go:embed ERC20Contract.json
	compiledContractJSON []byte
	contractBin          []byte
	contractABI          abi.ABI
)

func init() {
	var tmp struct {
		Abi string
		Bin string
	}
	err := json.Unmarshal(compiledContractJSON, &tmp)
	if err != nil {
		panic(err)
	}
	contractBin = common.FromHex(tmp.Bin)
	err = json.Unmarshal([]byte(tmp.Abi), &contractABI)
	if err != nil {
		panic(err)
	}
}

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
		{"bad height",
			func() {
				req = &types.QueryBlockBloomRequest{Height: -2}
			},
			false,
		},
		{
			"bloom from transient store",
			func() {
				req = &types.QueryBlockBloomRequest{Height: 1}
				bloom := ethtypes.BytesToBloom([]byte("bloom"))
				expBloom = bloom.Bytes()
				suite.app.EvmKeeper.WithContext(suite.ctx.WithBlockHeight(1))
				suite.app.EvmKeeper.SetBlockBloomTransient(bloom.Big())
			},
			true,
		},
		{"bloom not found for height",
			func() {
				req = &types.QueryBlockBloomRequest{Height: 100}
				bloom := ethtypes.BytesToBloom([]byte("bloom"))
				expBloom = bloom.Bytes()
				suite.ctx = suite.ctx.WithBlockHeight(100)
				suite.app.EvmKeeper.SetBlockBloom(suite.ctx, 2, bloom)
			},
			false,
		},
		{
			"success",
			func() {
				req = &types.QueryBlockBloomRequest{Height: 3}
				bloom := ethtypes.BytesToBloom([]byte("bloom"))
				expBloom = bloom.Bytes()
				suite.ctx = suite.ctx.WithBlockHeight(3)
				suite.app.EvmKeeper.SetBlockBloom(suite.ctx, 3, bloom)
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

// DeployTestContract deploy a test erc20 contract and returns the contract address
func (suite *KeeperTestSuite) deployTestContract(owner common.Address, supply *big.Int) common.Address {
	ctx := sdk.WrapSDKContext(suite.ctx)
	chainID := suite.app.EvmKeeper.ChainID()

	ctorArgs, err := contractABI.Pack("", owner, supply)
	suite.Require().NoError(err)

	data := append(contractBin, ctorArgs...)
	args, err := json.Marshal(&types.CallArgs{
		From: &suite.address,
		Data: (*hexutil.Bytes)(&data),
	})
	suite.Require().NoError(err)

	res, err := suite.queryClient.EstimateGas(ctx, &types.EthCallRequest{
		Args:   args,
		GasCap: uint64(ethermint.DefaultRPCGasLimit),
	})
	suite.Require().NoError(err)

	nonce := suite.app.EvmKeeper.GetNonce(suite.address)
	erc20DeployTx := types.NewTxContract(
		chainID,
		nonce,
		nil,     // amount
		res.Gas, // gasLimit
		nil,     // gasPrice
		data,    // input
		nil,     // accesses
	)
	erc20DeployTx.From = suite.address.Hex()
	err = erc20DeployTx.Sign(ethtypes.LatestSignerForChainID(chainID), suite.signer)
	suite.Require().NoError(err)
	rsp, err := suite.app.EvmKeeper.EthereumTx(ctx, erc20DeployTx)
	suite.Require().NoError(err)
	suite.Require().Empty(rsp.VmError)
	return crypto.CreateAddress(suite.address, nonce)
}

func (suite *KeeperTestSuite) TestEstimateGas() {
	ctx := sdk.WrapSDKContext(suite.ctx)
	gasHelper := hexutil.Uint64(20000)

	var (
		args   types.CallArgs
		gasCap uint64
	)
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		expGas   uint64
	}{
		// should success, because transfer value is zero
		{"default args", func() {
			args = types.CallArgs{To: &common.Address{}}
		}, true, 21000},
		// should fail, because the default From address(zero address) don't have fund
		{"not enough balance", func() {
			args = types.CallArgs{To: &common.Address{}, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0},
		// should success, enough balance now
		{"enough balance", func() {
			args = types.CallArgs{To: &common.Address{}, From: &suite.address, Value: (*hexutil.Big)(big.NewInt(100))}
		}, false, 0},
		// should success, because gas limit lower than 21000 is ignored
		{"gas exceed allowance", func() {
			args = types.CallArgs{To: &common.Address{}, Gas: &gasHelper}
		}, true, 21000},
		// should fail, invalid gas cap
		{"gas exceed global allowance", func() {
			args = types.CallArgs{To: &common.Address{}}
			gasCap = 20000
		}, false, 0},
		// estimate gas of an erc20 contract deployment, the exact gas number is checked with geth
		{"contract deployment", func() {
			ctorArgs, err := contractABI.Pack("", &suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Require().NoError(err)
			data := append(contractBin, ctorArgs...)
			args = types.CallArgs{
				From: &suite.address,
				Data: (*hexutil.Bytes)(&data),
			}
		}, true, 1144643},
		// estimate gas of an erc20 transfer, the exact gas number is checked with geth
		{"erc20 transfer", func() {
			contractAddr := suite.deployTestContract(suite.address, sdk.NewIntWithDecimal(1000, 18).BigInt())
			suite.Commit()
			transferData, err := contractABI.Pack("transfer", common.HexToAddress("0x378c50D9264C63F3F92B806d4ee56E9D86FfB3Ec"), big.NewInt(1000))
			suite.Require().NoError(err)
			args = types.CallArgs{To: &contractAddr, From: &suite.address, Data: (*hexutil.Bytes)(&transferData)}
		}, true, 51880},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			gasCap = ethermint.DefaultRPCGasLimit
			tc.malleate()

			args, err := json.Marshal(&args)
			suite.Require().NoError(err)
			req := types.EthCallRequest{
				Args:   args,
				GasCap: gasCap,
			}

			rsp, err := suite.queryClient.EstimateGas(ctx, &req)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expGas, rsp.Gas)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

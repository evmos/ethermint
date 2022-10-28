package backend

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"google.golang.org/grpc/metadata"
)

func (suite *BackendTestSuite) TestResend() {
	txNonce := (hexutil.Uint64)(1)
	baseFee := sdk.NewInt(1)
	gasPrice := new(hexutil.Big)
	toAddr := tests.GenerateAddress()
	callArgs := evmtypes.TransactionArgs{
		From:                 nil,
		To:                   &toAddr,
		Gas:                  nil,
		GasPrice:             nil,
		MaxFeePerGas:         gasPrice,
		MaxPriorityFeePerGas: gasPrice,
		Value:                gasPrice,
		Nonce:                nil,
		Input:                nil,
		Data:                 nil,
		AccessList:           nil,
	}

	testCases := []struct {
		name         string
		registerMock func()
		args         evmtypes.TransactionArgs
		gasPrice     *hexutil.Big
		gasLimit     *hexutil.Uint64
		expHash      common.Hash
		expPass      bool
	}{
		{
			"fail - Missing transaction nonce ",
			func() {},
			evmtypes.TransactionArgs{
				Nonce: nil,
			},
			nil,
			nil,
			common.Hash{},
			false,
		},
		{
			"fail - Can't set Tx defaults ",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFeeDisabled(queryClient)
			},
			evmtypes.TransactionArgs{
				Nonce: &txNonce,
			},
			nil,
			nil,
			common.Hash{},
			true,
		},
		{
			"fail - Can't set Tx defaults ",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, baseFee)
			},
			evmtypes.TransactionArgs{
				Nonce: &txNonce,
			},
			nil,
			nil,
			common.Hash{},
			true,
		},
		{
			"fail - MaxFeePerGas is nil",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFeeDisabled(queryClient)
			},
			evmtypes.TransactionArgs{
				Nonce:                &txNonce,
				MaxPriorityFeePerGas: nil,
				GasPrice:             nil,
				MaxFeePerGas:         nil,
			},
			nil,
			nil,
			common.Hash{},
			true,
		},
		{
			"fail - GasPrice and (MaxFeePerGas or MaxPriorityPerGas specified",
			func() {},
			evmtypes.TransactionArgs{
				Nonce:                &txNonce,
				MaxPriorityFeePerGas: nil,
				GasPrice:             gasPrice,
				MaxFeePerGas:         gasPrice,
			},
			nil,
			nil,
			common.Hash{},
			false,
		},
		{
			"fail - Block error",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlockError(client, 1)
			},
			evmtypes.TransactionArgs{
				Nonce: &txNonce,
			},
			nil,
			nil,
			common.Hash{},
			false,
		},
		{
			"fail - MaxFeePerGas is nil",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, baseFee)
			},
			evmtypes.TransactionArgs{
				Nonce:                &txNonce,
				GasPrice:             nil,
				MaxPriorityFeePerGas: gasPrice,
				MaxFeePerGas:         gasPrice,
			},
			nil,
			nil,
			common.Hash{},
			true,
		},
		{
			"pass - Chain Id is nil",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, baseFee)
			},
			evmtypes.TransactionArgs{
				Nonce:                &txNonce,
				MaxPriorityFeePerGas: gasPrice,
				ChainID:              nil,
			},
			nil,
			nil,
			common.Hash{},
			true,
		},
		{
			"pass - Gas is nil",
			func() {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.EVMQueryClient)
				var header metadata.MD
				RegisterParams(queryClient, &header, 1)
				RegisterBlock(client, 1, nil)
				RegisterBlockResults(client, 1)
				RegisterBaseFee(queryClient, baseFee)
				RegisterEstimateGas(queryClient, callArgs)
			},
			evmtypes.TransactionArgs{
				Nonce:                &txNonce,
				To:                   &toAddr,
				MaxFeePerGas:         gasPrice,
				MaxPriorityFeePerGas: gasPrice,
				Gas:                  nil,
			},
			gasPrice,
			nil,
			common.Hash{},
			false,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("case %s", tc.name), func() {
			suite.SetupTest() // reset test and queries
			tc.registerMock()

			hash, err := suite.backend.Resend(tc.args, tc.gasPrice, tc.gasLimit)
			suite.T().Log("hash", hash, "err", err)
			if tc.expPass {
				suite.Require().Equal(tc.expHash, hash)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

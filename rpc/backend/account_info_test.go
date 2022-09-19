package backend

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/tests"
)

func (suite *BackendTestSuite) TestGetCode() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))
	contractCode := []byte("0xef616c92f3cfc9e92dc270d6acff9cea213cecc7020a76ee4395af09bdceb4837a1ebdb5735e11e7d3adb6104e0c3ac55180b4ddf5e54d022cc5e8837f6a4f971b")

	testCases := []struct {
		name          string
		addr          common.Address
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(common.Address)
		expPass       bool
		expCode       hexutil.Bytes
	}{
		{
			"fail - BlockHash and BlockNumber are both nil ",
			tests.GenerateAddress(),
			rpctypes.BlockNumberOrHash{},
			func(addr common.Address) {},
			false,
			nil,
		},
		{
			"fail - query client errors on getting Code",
			tests.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterCodeError(queryClient, addr)
			},
			false,
			nil,
		},
		{
			"pass",
			tests.GenerateAddress(),
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterCode(queryClient, addr, contractCode)
			},
			true,
			contractCode,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset
			tc.registerMock(tc.addr)

			code, err := suite.backend.GetCode(tc.addr, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expCode, code)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetProof() {
	blockNrInvalid := rpctypes.NewBlockNumber(big.NewInt(1))
	// blockNr := rpctypes.NewBlockNumber(big.NewInt(4))
	// _, bz := suite.buildEthereumTx()

	testCases := []struct {
		name          string
		addr          common.Address
		storageKeys   []string
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(rpctypes.BlockNumber, common.Address)
		expPass       bool
		expAccRes     *rpctypes.AccountResult
	}{
		// fail - invalidBlockNumber
		{
			"fail - BlockNumeber = 1",
			tests.GenerateAddress(),
			[]string{},
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNrInvalid},
			func(bn rpctypes.BlockNumber, addr common.Address) {
				client := suite.backend.clientCtx.Client.(*mocks.Client)
				RegisterBlock(client, bn.Int64(), nil)
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterAccount(queryClient, addr, blockNrInvalid.Int64())
			},
			false,
			&rpctypes.AccountResult{},
		},
		// TODO How can I pass block height >=2 here? RegisterBlock doesn't accept it
		// {
		// 	"pass",
		// 	tests.GenerateAddress(),
		// 	[]string{},
		// 	rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
		// 	func(addr common.Address) {
		// 		client := suite.backend.clientCtx.Client.(*mocks.Client)
		// 		RegisterBlock(client, blockNr.Int64(), nil)
		// 		queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
		// 		RegisterAccount(queryClient, addr, blockNr.Int64())
		// 	},
		// 	true,
		// 	&rpctypes.AccountResult{},
		// },
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.registerMock(*tc.blockNrOrHash.BlockNumber, tc.addr)

			accRes, err := suite.backend.GetProof(tc.addr, tc.storageKeys, tc.blockNrOrHash)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expAccRes, accRes)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *BackendTestSuite) TestGetStorageAt() {
	blockNr := rpctypes.NewBlockNumber(big.NewInt(1))

	testCases := []struct {
		name          string
		addr          common.Address
		key           string
		blockNrOrHash rpctypes.BlockNumberOrHash
		registerMock  func(common.Address, string, string)
		expPass       bool
		expStorage    hexutil.Bytes
	}{
		{
			"fail - BlockHash and BlockNumber are both nil",
			tests.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{},
			func(addr common.Address, key string, storage string) {},
			false,
			nil,
		},
		{
			"fail - query client errors on getting Storage",
			tests.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address, key string, storage string) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterStorageAtError(queryClient, addr, key)
			},
			false,
			nil,
		},
		{
			"pass",
			tests.GenerateAddress(),
			"0x0",
			rpctypes.BlockNumberOrHash{BlockNumber: &blockNr},
			func(addr common.Address, key string, storage string) {
				queryClient := suite.backend.queryClient.QueryClient.(*mocks.QueryClient)
				RegisterStorageAt(queryClient, addr, key, storage)
			},
			true,
			hexutil.Bytes{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest()
			tc.registerMock(tc.addr, tc.key, tc.expStorage.String())

			storage, err := suite.backend.GetStorageAt(tc.addr, tc.key, tc.blockNrOrHash)
			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expStorage, storage)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

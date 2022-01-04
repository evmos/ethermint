package e2e_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"

	// . "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/testutil/network"
	ethermint "github.com/tharsis/ethermint/types"
)

// var _ = Describe("E2e", func() {
// })

// func TestJsonRpc(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "JSON-RPC Suite")
// }

// TODO: migrate to Ginkgo BDD
type IntegrationTestSuite struct {
	suite.Suite

	ctx     context.Context
	cfg     network.Config
	network *network.Network

	gethClient *gethclient.Client
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	cfg := network.DefaultConfig()
	cfg.JSONRPCAddress = config.DefaultJSONRPCAddress
	cfg.NumValidators = 1

	s.ctx = context.Background()
	s.cfg = cfg
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)
	s.Require().NotNil(s.network)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	address := fmt.Sprintf("http://%s", s.network.Validators[0].AppConfig.JSONRPC.Address)

	if s.network.Validators[0].JSONRPCClient == nil {
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}

	rpcClient, err := rpc.DialContext(s.ctx, address)
	s.Require().NoError(err)
	s.gethClient = gethclient.New(rpcClient)
	s.Require().NotNil(s.gethClient)
}

func (s *IntegrationTestSuite) TestChainID() {
	genesisRes, err := s.network.Validators[0].RPCClient.Genesis(s.ctx)
	s.Require().NoError(err)

	chainID, err := s.network.Validators[0].JSONRPCClient.ChainID(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(chainID)

	s.T().Log(chainID.Int64())

	eip155ChainID, err := ethermint.ParseChainID(s.network.Config.ChainID)
	s.Require().NoError(err)
	eip155ChainIDGen, err := ethermint.ParseChainID(genesisRes.Genesis.ChainID)
	s.Require().NoError(err)

	s.Require().Equal(chainID, eip155ChainID)
	s.Require().Equal(eip155ChainID, eip155ChainIDGen)
}

func (s *IntegrationTestSuite) TestNodeInfo() {
	// Not implemented
	info, err := s.gethClient.GetNodeInfo(s.ctx)
	s.Require().Error(err)
	s.Require().Empty(info)
}

func (s *IntegrationTestSuite) TestCreateAccessList() {
	// Not implemented
	accessList, _, _, err := s.gethClient.CreateAccessList(s.ctx, ethereum.CallMsg{})
	s.Require().Error(err)
	s.Require().Nil(accessList)
}

func (s *IntegrationTestSuite) TestBlock() {
	blockNum, err := s.network.Validators[0].JSONRPCClient.BlockNumber(s.ctx)
	s.Require().NoError(err)
	s.Require().NotZero(blockNum)

	bn := int64(blockNum)

	block, err := s.network.Validators[0].RPCClient.Block(s.ctx, &bn)
	s.Require().NoError(err)
	s.Require().NotNil(block)

	blockByNum, err := s.network.Validators[0].JSONRPCClient.BlockByNumber(s.ctx, new(big.Int).SetUint64(blockNum))
	s.Require().NoError(err)
	s.Require().NotNil(blockByNum)

	// compare the ethereum header with the tendermint header
	s.Require().Equal(len(block.Block.Txs), len(blockByNum.Body().Transactions))
	s.Require().Equal(block.Block.LastBlockID.Hash.Bytes(), blockByNum.Header().ParentHash.Bytes())

	hash := common.BytesToHash(block.Block.Hash())
	block, err = s.network.Validators[0].RPCClient.BlockByHash(s.ctx, hash.Bytes())
	s.Require().NoError(err)
	s.Require().NotNil(block)

	blockByHash, err := s.network.Validators[0].JSONRPCClient.BlockByHash(s.ctx, hash)
	s.Require().NoError(err)
	s.Require().NotNil(blockByHash)
	s.Require().Equal(blockByNum, blockByHash)

	// TODO: parse Tm block to Ethereum and compare
}

func (s *IntegrationTestSuite) TestHeader() {
	blockNum, err := s.network.Validators[0].JSONRPCClient.BlockNumber(s.ctx)
	s.Require().NoError(err)
	s.Require().NotZero(blockNum)

	bn := int64(blockNum)

	block, err := s.network.Validators[0].RPCClient.Block(s.ctx, &bn)
	s.Require().NoError(err)
	s.Require().NotNil(block)

	hash := common.BytesToHash(block.Block.Hash())

	headerByNum, err := s.network.Validators[0].JSONRPCClient.HeaderByNumber(s.ctx, new(big.Int).SetUint64(blockNum))
	s.Require().NoError(err)
	s.Require().NotNil(headerByNum)

	headerByHash, err := s.network.Validators[0].JSONRPCClient.HeaderByHash(s.ctx, hash)
	s.Require().NoError(err)
	s.Require().NotNil(headerByHash)
	s.Require().Equal(headerByNum, headerByHash)

	// TODO: we need to convert the ethereum block and return the header
	// header := rpctypes.EthHeaderFromTendermint(block.Block.Header, ethtypes.Bloom{}, headerByHash.BaseFee)
	// s.Require().NotNil(header)
	// s.Require().Equal(headerByHash, header)
}

func (s *IntegrationTestSuite) TestSendRawTransaction() {
	testCases := []struct {
		name           string
		data           string
		expEncodingErr bool
		expError       bool
	}{
		{
			"rlp: expected input list for types.LegacyTx",
			"0x85b7119c978b22ac5188a554916d5eb9000567b87b3b8a536222c3c2e6549b98",
			true,
			false,
		},
		{
			"transaction type not supported",
			"0x1238b01bfc01e946ffdf8ccb087a072298cf9f141899c5c586550cc910b8c5aa",
			true,
			false,
		},
		{
			"rlp: element is larger than containing list",
			"0xd46e8dd67c5d32be8d46e8dd67c5d32be8058bb8eb970870f072445675058bb8eb970870f072445675",
			true,
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var data hexutil.Bytes

			err := data.UnmarshalText([]byte(tc.data))
			s.Require().NoError(err, data)

			tx := new(ethtypes.Transaction)
			err = tx.UnmarshalBinary(data)
			if tc.expEncodingErr {
				s.Require().Error(err)
				s.Require().Equal(tc.name, err.Error())
				return
			}

			s.Require().NoError(err)
			s.Require().NotEmpty(tx)

			hash := tx.Hash()

			err = s.network.Validators[0].JSONRPCClient.SendTransaction(s.ctx, tx)
			if tc.expError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)

			err = s.network.WaitForNextBlock()
			s.Require().NoError(err)

			expTx, isPending, err := s.network.Validators[0].JSONRPCClient.TransactionByHash(s.ctx, hash)

			if tc.expError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().False(isPending)
			s.Require().Equal(tx, expTx)
		})
	}
}

func (s *IntegrationTestSuite) TestEstimateGasContractDeployment() {
	bytecode := "0x608060405234801561001057600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a260d08061004d6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063eb8ac92114602d575b600080fd5b606060048036036040811015604157600080fd5b8101908080359060200190929190803590602001909291905050506062565b005b8160008190555080827ff3ca124a697ba07e8c5e80bebcfcc48991fc16a63170e8a9206e30508960d00360405160405180910390a3505056fea265627a7a723158201d94d2187aaf3a6790527b615fcc40970febf0385fa6d72a2344848ebd0df3e964736f6c63430005110032"
	expectedGas := uint64(0x1879c)

	var data hexutil.Bytes

	err := data.UnmarshalText([]byte(bytecode))

	s.Require().NoError(err, data)

	gas, err := s.network.Validators[0].JSONRPCClient.EstimateGas(s.ctx, ethereum.CallMsg{
		Data: data,
	})

	s.Require().NoError(err)
	s.Require().Equal(expectedGas, gas)
}

func (s *IntegrationTestSuite) TestSendTransactionContractDeploymentNoGas() {
	bytecode := "0x6080604052348015600f57600080fd5b5060117f775a94827b8fd9b519d36cd827093c664f93347070a554f65e4a6f56cd73889860405160405180910390a2603580604b6000396000f3fe6080604052600080fdfea165627a7a723058206cab665f0f557620554bb45adf266708d2bd349b8a4314bdff205ee8440e3c240029"

	var data hexutil.Bytes
	err := data.UnmarshalText([]byte(bytecode))

	var testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	tx := ethtypes.NewContractCreation(0, nil, 0x5208, nil, data)
	tx, _ = ethtypes.SignTx(tx, ethtypes.HomesteadSigner{}, testKey)

	err = s.network.Validators[0].JSONRPCClient.SendTransaction(s.ctx, tx)
	s.Require().Error(err)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

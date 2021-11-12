package e2e_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	// . "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"

	"github.com/stretchr/testify/suite"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

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

	if s.network.Validators[0].JSONRPCClient == nil {
		address := fmt.Sprintf("http://%s", s.network.Validators[0].AppConfig.JSONRPC.Address)
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(address)
		s.Require().NoError(err)
	}
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

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

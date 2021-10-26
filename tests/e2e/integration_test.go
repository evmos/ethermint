package e2e

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"

	rpctypes "github.com/tharsis/ethermint/rpc/ethereum/types"
	"github.com/tharsis/ethermint/server/config"
	"github.com/tharsis/ethermint/testutil/network"
	ethermint "github.com/tharsis/ethermint/types"
)

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
		s.network.Validators[0].JSONRPCClient, err = ethclient.Dial(s.network.Validators[0].JSONRPCAddress)
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

	blockByNum, err := s.network.Validators[0].JSONRPCClient.BlockByNumber(s.ctx, new(big.Int).SetUint64(blockNum))
	s.Require().NoError(err)
	s.Require().NotNil(blockByNum)

	hash := blockByNum.Hash()
	blockByHash, err := s.network.Validators[0].JSONRPCClient.BlockByHash(s.ctx, hash)
	s.Require().NoError(err)
	s.Require().NotNil(blockByHash)
	s.Require().Equal(blockByNum, blockByHash)

	block, err := s.network.Validators[0].RPCClient.BlockByHash(s.ctx, hash.Bytes())
	s.Require().NoError(err)
	s.Require().NotNil(block)

	// TODO: parse Tm block to Ethereum and compare
}

func (s *IntegrationTestSuite) TestHash() {
	blockNum, err := s.network.Validators[0].JSONRPCClient.BlockNumber(s.ctx)
	s.Require().NoError(err)
	s.Require().NotZero(blockNum)

	headerByNum, err := s.network.Validators[0].JSONRPCClient.HeaderByNumber(s.ctx, new(big.Int).SetUint64(blockNum))
	s.Require().NoError(err)
	s.Require().NotNil(headerByNum)

	hash := headerByNum.Hash()
	headerByHash, err := s.network.Validators[0].JSONRPCClient.HeaderByHash(s.ctx, hash)
	s.Require().NoError(err)
	s.Require().NotNil(headerByHash)
	s.Require().Equal(headerByNum, headerByHash)

	block, err := s.network.Validators[0].RPCClient.BlockByHash(s.ctx, hash.Bytes())
	s.Require().NoError(err)
	s.Require().NotNil(block)

	header := rpctypes.EthHeaderFromTendermint(block.Block.Header, headerByHash.BaseFee)
	s.Require().NotNil(header)
	s.Require().Equal(headerByHash, header)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

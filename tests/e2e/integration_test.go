package e2e

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/suite"

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

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	s.ctx = context.Background()
	s.cfg = cfg
	s.network = network.New(s.T(), cfg)
	s.Require().NotNil(s.network)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	cl, err := ethclient.Dial(s.network.Validators[0].JSONRPCAddress)
	s.Require().NoError(err, "failed to dial JSON-RPC at %s", s.network.Validators[0].JSONRPCAddress)
	s.network.Validators[0].JSONRPCClient = cl
}

func (s *IntegrationTestSuite) TestChainID() {
	chainID, err := s.network.Validators[0].JSONRPCClient.ChainID(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(chainID)

	s.T().Log(chainID.Int64())

	eip155ChainID, err := ethermint.ParseChainID(s.network.Config.ChainID)
	s.Require().NoError(err)
	s.Require().Equal(chainID, eip155ChainID)
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/ethermint/version"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type apisTestSuite struct {
	suite.Suite
	Stop context.CancelFunc
	Port int
}

func (s *apisTestSuite) SetupSuite() {
	stop, port, err := startAPIServer()
	require.Nil(s.T(), err, "unexpected error")
	s.Stop = stop
	s.Port = port
}

func (s *apisTestSuite) TearDownSuite() {
	s.Stop()
}

func (s *apisTestSuite) TestPublicWeb3APIClientVersion() {
	res, err := rpcCall(s.Port, "web3_clientVersion", []string{})
	require.Nil(s.T(), err, "unexpected error")
	require.Equal(s.T(), version.ClientVersion(), res)
}

func (s *apisTestSuite) TestPublicWeb3APISha3() {
	res, err := rpcCall(s.Port, "web3_sha3", []string{"0x67656c6c6f20776f726c64"})
	require.Nil(s.T(), err, "unexpected error")
	require.Equal(s.T(), "0x1b84adea42d5b7d192fd8a61a85b25abe0757e9a65cab1da470258914053823f", res)
}

func (s *apisTestSuite) TestMiningAPIs() {
	res, err := rpcCall(s.Port, "eth_mining", nil)
	require.Nil(s.T(), err, "unexpected error")
	require.Equal(s.T(), false, res)

	res, err = rpcCall(s.Port, "eth_hashrate", nil)
	require.Nil(s.T(), err, "unexpected error")
	require.Equal(s.T(), "0x0", res)
}

func TestAPIsTestSuite(t *testing.T) {
	suite.Run(t, new(apisTestSuite))
}

func startAPIServer() (context.CancelFunc, int, error) {
	config := &Config{
		RPCAddr: "127.0.0.1",
		RPCPort: randomPort(),
	}
	timeouts := rpc.HTTPTimeouts{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err := StartHTTPEndpoint(ctx, config, GetRPCAPIs(), timeouts)
	if err != nil {
		return cancel, 0, err
	}

	return cancel, config.RPCPort, nil
}

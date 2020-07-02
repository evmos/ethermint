package rpc

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// PublicNetAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicNetAPI struct {
	networkVersion uint64
}

// NewPersonalEthAPI creates an instance of the public ETH Web3 API.
func NewPublicNetAPI(cliCtx context.CLIContext) *PublicNetAPI {
	chainID := viper.GetString(flags.FlagChainID)
	// parse the chainID from a integer string
	intChainID, err := strconv.ParseUint(chainID, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid chainID: %s, must be integer format", chainID))
	}

	return &PublicNetAPI{
		networkVersion: intChainID,
	}
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

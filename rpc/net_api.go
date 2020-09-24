package rpc

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	ethermint "github.com/cosmos/ethermint/types"
)

// PublicNetAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicNetAPI struct {
	networkVersion uint64
}

// NewPublicNetAPI creates an instance of the public Net Web3 API.
func NewPublicNetAPI(_ context.CLIContext) *PublicNetAPI {
	chainID := viper.GetString(flags.FlagChainID)
	// parse the chainID from a integer string
	chainIDEpoch, err := ethermint.ParseChainID(chainID)
	if err != nil {
		panic(err)
	}

	return &PublicNetAPI{
		networkVersion: chainIDEpoch.Uint64(),
	}
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

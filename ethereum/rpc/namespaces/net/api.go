package net

import (
	"fmt"

	ethermint "github.com/tharsis/ethermint/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// PublicAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicAPI struct {
	networkVersion uint64
}

// NewPublicAPI creates an instance of the public Net Web3 API.
func NewPublicAPI(clientCtx client.Context) *PublicAPI {
	// parse the chainID from a integer string
	chainIDEpoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	return &PublicAPI{
		networkVersion: chainIDEpoch.Uint64(),
	}
}

// Version returns the current ethereum protocol version.
func (s *PublicAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

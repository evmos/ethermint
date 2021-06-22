package net

import (
	"fmt"

	ethermint "github.com/tharsis/ethermint/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// PublicApi is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicApi struct {
	networkVersion uint64
}

// NewPublicApi creates an instance of the public Net Web3 API.
func NewPublicApi(clientCtx client.Context) *PublicApi {
	// parse the chainID from a integer string
	chainIDEpoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	return &PublicApi{
		networkVersion: chainIDEpoch.Uint64(),
	}
}

// Version returns the current ethereum protocol version.
func (s *PublicApi) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

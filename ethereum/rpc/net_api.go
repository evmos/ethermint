package rpc

import (
	"fmt"

	ethermint "github.com/cosmos/ethermint/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// PublicNetAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicNetAPI struct {
	networkVersion uint64
}

// NewPublicNetAPI creates an instance of the public Net Web3 API.
func NewPublicNetAPI(clientCtx client.Context) *PublicNetAPI {
	// parse the chainID from a integer string
	chainIDEpoch, err := ethermint.ParseChainID(clientCtx.ChainID)
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

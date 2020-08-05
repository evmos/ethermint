package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSetBech32Prefixes(t *testing.T) {
	config := sdk.GetConfig()
	require.Equal(t, sdk.Bech32PrefixAccAddr, config.GetBech32AccountAddrPrefix())
	require.Equal(t, sdk.Bech32PrefixAccPub, config.GetBech32AccountPubPrefix())
	require.Equal(t, sdk.Bech32PrefixValAddr, config.GetBech32ValidatorAddrPrefix())
	require.Equal(t, sdk.Bech32PrefixValPub, config.GetBech32ValidatorPubPrefix())
	require.Equal(t, sdk.Bech32PrefixConsAddr, config.GetBech32ConsensusAddrPrefix())
	require.Equal(t, sdk.Bech32PrefixConsPub, config.GetBech32ConsensusPubPrefix())

	SetBech32Prefixes(config)
	require.Equal(t, Bech32PrefixAccAddr, config.GetBech32AccountAddrPrefix())
	require.Equal(t, Bech32PrefixAccPub, config.GetBech32AccountPubPrefix())
	require.Equal(t, Bech32PrefixValAddr, config.GetBech32ValidatorAddrPrefix())
	require.Equal(t, Bech32PrefixValPub, config.GetBech32ValidatorPubPrefix())
	require.Equal(t, Bech32PrefixConsAddr, config.GetBech32ConsensusAddrPrefix())
	require.Equal(t, Bech32PrefixConsPub, config.GetBech32ConsensusPubPrefix())
}

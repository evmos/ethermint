package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
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

	require.Equal(t, sdk.GetConfig().GetBech32AccountAddrPrefix(), config.GetBech32AccountAddrPrefix())
	require.Equal(t, sdk.GetConfig().GetBech32AccountPubPrefix(), config.GetBech32AccountPubPrefix())
	require.Equal(t, sdk.GetConfig().GetBech32ValidatorAddrPrefix(), config.GetBech32ValidatorAddrPrefix())
	require.Equal(t, sdk.GetConfig().GetBech32ValidatorPubPrefix(), config.GetBech32ValidatorPubPrefix())
	require.Equal(t, sdk.GetConfig().GetBech32ConsensusAddrPrefix(), config.GetBech32ConsensusAddrPrefix())
	require.Equal(t, sdk.GetConfig().GetBech32ConsensusPubPrefix(), config.GetBech32ConsensusPubPrefix())
}

func TestSetCoinType(t *testing.T) {
	config := sdk.GetConfig()
	require.Equal(t, sdk.CoinType, int(config.GetCoinType()))
	require.Equal(t, sdk.FullFundraiserPath, config.GetFullFundraiserPath())

	SetBip44CoinType(config)
	require.Equal(t, Bip44CoinType, int(config.GetCoinType()))
	require.Equal(t, sdk.GetConfig().GetCoinType(), config.GetCoinType())
	require.Equal(t, sdk.GetConfig().GetFullFundraiserPath(), config.GetFullFundraiserPath())
}

func TestHDPath(t *testing.T) {
	params := *hd.NewFundraiserParams(0, Bip44CoinType, 0)
	// need to prepend "m/" because the below method provided by the sdk does not add the proper prepending
	hdPath := "m/" + params.String()
	require.Equal(t, "m/44'/60'/0'/0/0", hdPath)
	require.Equal(t, hdPath, BIP44HDPath)
}

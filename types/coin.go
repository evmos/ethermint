package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AttoPhoton defines the default coin denomination used in Ethermint in:
	//
	// - Staking parameters: denomination used as stake in the dPoS chain
	// - Mint parameters: denomination minted due to fee distribution rewards
	// - Governance parameters: denomination used for spam prevention in proposal deposits
	// - Crisis parameters: constant fee denomination used for spam prevention to check broken invariant
	// - EVM parameters: denomination used for running EVM state transitions in Ethermint.
	AttoPhoton string = "aphoton"

	// BaseDenomUnit defines the base denomination unit for Photons.
	// 1 photon = 1x10^{BaseDenomUnit} aphoton
	BaseDenomUnit = 18
)

// NewPhotonCoin is a utility function that returns an "aphoton" coin with the given sdk.Int amount.
// The function will panic if the provided amount is negative.
func NewPhotonCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(AttoPhoton, amount)
}

// NewPhotonDecCoin is a utility function that returns an "aphoton" decimal coin with the given sdk.Int amount.
// The function will panic if the provided amount is negative.
func NewPhotonDecCoin(amount sdk.Int) sdk.DecCoin {
	return sdk.NewDecCoin(AttoPhoton, amount)
}

// NewPhotonCoinInt64 is a utility function that returns an "aphoton" coin with the given int64 amount.
// The function will panic if the provided amount is negative.
func NewPhotonCoinInt64(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(AttoPhoton, amount)
}

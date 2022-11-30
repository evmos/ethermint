package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AttoENTGL defines the default coin denomination used in Ethermint in:
	//
	// - Staking parameters: denomination used as stake in the dPoS chain
	// - Mint parameters: denomination minted due to fee distribution rewards
	// - Governance parameters: denomination used for spam prevention in proposal deposits
	// - Crisis parameters: constant fee denomination used for spam prevention to check broken invariant
	// - EVM parameters: denomination used for running EVM state transitions in Ethermint.
	AttoENTGL string = "aENTGL"

	// BaseDenomUnit defines the base denomination unit for ENTGLs.
	// 1 ENTGL = 1x10^{BaseDenomUnit} aENTGL
	BaseDenomUnit = 18

	// DefaultGasPrice is default gas price for evm transactions
	DefaultGasPrice = 20
)

// PowerReduction defines the default power reduction value for staking
var PowerReduction = sdkmath.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(BaseDenomUnit), nil))

// NewENTGLCoin is a utility function that returns an "aENTGL" coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewENTGLCoin(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(AttoENTGL, amount)
}

// NewENTGLDecCoin is a utility function that returns an "aENTGL" decimal coin with the given sdkmath.Int amount.
// The function will panic if the provided amount is negative.
func NewENTGLDecCoin(amount sdkmath.Int) sdk.DecCoin {
	return sdk.NewDecCoin(AttoENTGL, amount)
}

// NewENTGLCoinInt64 is a utility function that returns an "aENTGL" coin with the given int64 amount.
// The function will panic if the provided amount is negative.
func NewENTGLCoinInt64(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(AttoENTGL, amount)
}

package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	// "github.com/ethereum/go-ethereum/params"

	"github.com/tharsis/ethermint/x/evm/types"
)

// CalculateBaseFee calculates the base fee for the current block. This is only calculated once per
// block during EndBlock. If the NoBaseFee parameter is enabled, this function returns nil.
// NOTE: This code is inspired from the go-ethereum EIP1559 implementation and adapted to Cosmos SDK-based
// chains. For the canonical code refer to: https://github.com/ethereum/go-ethereum/blob/master/consensus/misc/eip1559.go
func (k Keeper) CalculateBaseFee(ctx sdk.Context) *big.Int {
	consParams := ctx.ConsensusParams()
	params := k.GetParams(ctx)

	if params.NoBaseFee {
		return nil
	}

	chainConfig := params.ChainConfig
	cfg := chainConfig.EthereumConfig(k.eip155ChainID)

	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	// if !chainConfig.EthereumConfig(k.eip155ChainID).IsLondon(big.NewInt(ctx.BlockHeight())) {
	// 	return new(big.Int).SetUint64(types.InitialBaseFee)
	// }

	// FIXME: remove and uncomment line above
	height := big.NewInt(ctx.BlockHeight())

	rules := cfg.Rules(height)
	if rules.IsBerlin || cfg.IsMuirGlacier(height) || rules.IsIstanbul || rules.IsByzantium || rules.IsConstantinople {
		return new(big.Int).SetUint64(types.InitialBaseFee)
	}

	// get the block gas used and the base fee values for the parent block.
	parentBaseFee := k.GetBaseFee(ctx)
	parentGasUsed := k.GetBlockGasUsed(ctx)

	gasLimit := new(big.Int).SetUint64(math.MaxUint64)
	if consParams != nil && consParams.Block.MaxGas > -1 {
		gasLimit = big.NewInt(consParams.Block.MaxGas)
	}

	parentGasTargetBig := new(big.Int).Div(gasLimit, big.NewInt(types.ElasticityMultiplier)) // TODO: update to geth
	if !parentGasTargetBig.IsUint64() {
		return new(big.Int).SetUint64(types.InitialBaseFee) // TODO: update to geth
	}

	parentGasTarget := parentGasTargetBig.Uint64()
	baseFeeChangeDenominator := new(big.Int).SetUint64(types.BaseFeeChangeDenominator) // TODO: update to geth

	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parentGasUsed == parentGasTarget {
		return new(big.Int).Set(parentBaseFee)
	}

	if parentGasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		gasUsedDelta := new(big.Int).SetUint64(parentGasUsed - parentGasTarget)
		x := new(big.Int).Mul(parentBaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := math.BigMax(
			x.Div(y, baseFeeChangeDenominator),
			common.Big1,
		)

		return x.Add(parentBaseFee, baseFeeDelta)
	}

	// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
	gasUsedDelta := new(big.Int).SetUint64(parentGasTarget - parentGasUsed)
	x := new(big.Int).Mul(parentBaseFee, gasUsedDelta)
	y := x.Div(x, parentGasTargetBig)
	baseFeeDelta := x.Div(y, baseFeeChangeDenominator)

	return math.BigMax(
		x.Sub(parentBaseFee, baseFeeDelta),
		common.Big0,
	)
}

package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
)

// CalculateBaseFee calculates the base fee for the current block. This is only calculated once per
// block during EndBlock.
func (k Keeper) CalculateBaseFee(ctx sdk.Context) *big.Int {
	consParams := ctx.ConsensusParams()
	chainConfig := k.GetParams(ctx).ChainConfig

	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	// TODO: create util function
	if !chainConfig.EthereumConfig(k.eip155ChainID).IsLondon(big.NewInt(ctx.BlockHeight())) {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}

	// get the block gas used and the base fee values for the parent block.
	parentBaseFee := k.GetBaseFee(ctx)
	parentGasUsed := k.GetBlockGasUsed(ctx)

	gasLimit := new(big.Int).SetUint64(math.MaxUint64)
	if consParams != nil && consParams.Block.MaxGas > -1 {
		gasLimit = big.NewInt(consParams.Block.MaxGas)
	}

	parentGasTargetBig := new(big.Int).Div(gasLimit, big.NewInt(params.ElasticityMultiplier))
	if !parentGasTargetBig.IsUint64() {
		return new(big.Int).SetUint64(params.InitialBaseFee)
	}

	parentGasTarget := parentGasTargetBig.Uint64()
	baseFeeChangeDenominator := new(big.Int).SetUint64(params.BaseFeeChangeDenominator)

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

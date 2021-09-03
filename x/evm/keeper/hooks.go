package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/x/evm/types"
)

var (
	_ types.EvmHooks = MultiEvmHooks{}
)

// MultiEvmHooks combine multiple evm hooks, all hook functions are run in array sequence
type MultiEvmHooks []types.EvmHooks

// NewMultiEvmHooks combine multiple evm hooks
func NewMultiEvmHooks(hooks ...types.EvmHooks) MultiEvmHooks {
	return hooks
}

// PostTxProcessing delegate the call to underlying hooks
func (mh MultiEvmHooks) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	for i := range mh {
		if err := mh[i].PostTxProcessing(ctx, txHash, logs); err != nil {
			return sdkerrors.Wrapf(err, "EVM hook %T failed", mh[i])
		}
	}
	return nil
}

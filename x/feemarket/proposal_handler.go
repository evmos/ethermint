package feemarket

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/ethermint/x/feemarket/keeper"
	"github.com/tharsis/ethermint/x/feemarket/types"
	"math/big"
)

// NewBaseFeeChangeProposalHandler creates a new governance Handler for a BaseFeeChangeProposal
func NewBaseFeeChangeProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.BaseFeeChangeProposal:
			k.SetBaseFee(ctx, new(big.Int).SetUint64(c.BaseFee))
			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized cronos proposal content type: %T", c)
		}
	}
}

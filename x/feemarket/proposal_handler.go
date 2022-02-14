package feemarket

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/tharsis/ethermint/x/feemarket/keeper"
	"github.com/tharsis/ethermint/x/feemarket/types"
)

// NewBaseFeeChangeProposalHandler creates a new governance Handler for a BaseFeeChangeProposal
func NewBaseFeeChangeProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.BaseFeeChangeProposal:
			k.SetBaseFee(ctx, new(big.Int).SetUint64(c.BaseFee))
			return nil
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized proposal content type: %T", c)
		}
	}
}

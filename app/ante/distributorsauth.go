package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DistributorsAuthDecorator keeps track if transaction have correct distributor signer
// NOTE: This decorator does not perform any validation
type DistributorsAuthDecorator struct {
	distrKeeper DistributorsAuthKeeper
}

// NewGasWantedDecorator creates a new NewGasWantedDecorator
func NewDistributorsAuthDecorator(
	distrKeeper DistributorsAuthKeeper,
) DistributorsAuthDecorator {
	fmt.Println("DistributorsAuthDecorator")
	return DistributorsAuthDecorator{
		distrKeeper,
	}
}

func (dg DistributorsAuthDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	fmt.Println("AnteHandle")
	for _, msg := range tx.GetMsgs() {
		// ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
		// if !ok {
		// 	return ctx, sdkerrors.Wrapf(
		// 		sdkerrors.ErrUnknownRequest,
		// 		"invalid message type %T, expected %T",
		// 		msg, (*evmtypes.MsgEthereumTx)(nil),
		// 	)
		// }

		found_correct_signer := false
		signers := msg.GetSigners()
		for _, signer := range signers {
			// fmt.Println("Signer ", signer.String())
			if dg.distrKeeper.ValidateTransaction(ctx, signer.String()) == nil {
				// fmt.Println("Correct distributor signer with address", signer.String())
				found_correct_signer = true
			}
		}

		// fmt.Println("found_correct_signer ", found_correct_signer)
		if !found_correct_signer {
			// fmt.Println("error")
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Transaction have no valid Distributor signers")
		}
	}

	// fmt.Println("success")
	return next(ctx, tx, simulate)
}

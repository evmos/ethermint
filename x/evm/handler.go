package evm

import (
	"fmt"
	"runtime/debug"

	mintlog "github.com/tharsis/ethermint/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/tharsis/ethermint/x/evm/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(server types.MsgServer) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		defer Recover(&err)

		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgEthereumTx:
			// execute state transition
			res, err := server.EthereumTx(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			err := sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, err
		}
	}
}

func Recover(err *error) {
	if r := recover(); r != nil {
		*err = sdkerrors.Wrapf(sdkerrors.ErrPanic, "%v", r)

		if e, ok := r.(error); ok {

			(*mintlog.EthermintLoggerInstance.TendermintLogger).Error(fmt.Sprintf("evm msg handler panicked with an error %v", e))
			(*mintlog.EthermintLoggerInstance.TendermintLogger).Debug(string(debug.Stack()))
		} else {
			(*mintlog.EthermintLoggerInstance.TendermintLogger).Error(fmt.Sprintf("%v", r))
		}
	}
}

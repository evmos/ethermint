package evm

import (
	"fmt"
	"runtime/debug"

	log "github.com/xlab/suplog"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k *keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
		defer Recover(&err)

		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgEthereumTx:
			// execute state transition
			res, err := k.EthereumTx(sdk.WrapSDKContext(ctx), msg)
			if err != nil {
				return sdk.WrapServiceResult(ctx, res, err)
			} else if result, err = sdk.WrapServiceResult(ctx, res, nil); err != nil {
				return nil, err
			}

			// log state transition result
			var recipientLog string
			if res.ContractAddress != "" {
				recipientLog = fmt.Sprintf("contract address %s", res.ContractAddress)
			} else {
				var recipient string
				if to := msg.To(); to != nil {
					recipient = to.Hex()
				}

				recipientLog = fmt.Sprintf("recipient address %s", recipient)
			}

			sender := ethcmn.BytesToAddress(msg.GetFrom().Bytes())

			if res.Reverted {
				result.Log = "transaction reverted"
				log := fmt.Sprintf(
					"reverted EVM state transition; sender address %s; %s", sender.Hex(), recipientLog,
				)
				k.Logger(ctx).Info(log)
			} else {
				log := fmt.Sprintf(
					"executed EVM state transition; sender address %s; %s", sender.Hex(), recipientLog,
				)
				result.Log = log
				k.Logger(ctx).Info(log)
			}

			return result, nil

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
			log.WithError(e).Errorln("evm msg handler panicked with an error")
			log.Debugln(string(debug.Stack()))
		} else {
			log.Errorln(r)
		}
	}
}

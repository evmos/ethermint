package evm

import (
	"fmt"
	"time"

	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	defer telemetry.MeasureSince(time.Now(), "evm", "state_transition")

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		snapshotStateDB := k.CommitStateDB.Copy()

		// The "recover" code here is used to solve the problem of dirty data
		// in CommitStateDB due to insufficient gas.

		// The following is a detailed description:
		// If the gas is insufficient during the execution of the "handler",
		// panic will be thrown from the function "ConsumeGas" and finally
		// caught by the function "runTx" from Cosmos. The function "runTx"
		// will think that the execution of Msg has failed and the modified
		// data in the Store will not take effect.

		// Stacktrace：runTx->runMsgs->handler->...->gaskv.Store.Set->ConsumeGas

		// The problem is that when the modified data in the Store does not take
		// effect, the data in the modified CommitStateDB is not rolled back,
		// they take effect, and dirty data is generated.
		// Therefore, the code here specifically deals with this situation.
		// See https://github.com/cosmos/ethermint/issues/668 for more information.
		defer func() {
			if r := recover(); r != nil {
				// We first used "k.CommitStateDB = snapshotStateDB" to roll back
				// CommitStateDB, but this can only change the CommitStateDB in the
				// current Keeper object, but the Keeper object will be destroyed
				// soon, it is not a global variable, so the content pointed to by
				// the CommitStateDB pointer can be modified to take effect.
				types.CopyCommitStateDB(snapshotStateDB, k.CommitStateDB)
				panic(r)
			}
		}()
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgEthereumTx:
			// execute state transition
			res, err := k.EthereumTx(sdk.WrapSDKContext(ctx), msg)
			if err != nil {
				return nil, err
			}

			result, err := sdk.WrapServiceResult(ctx, res, err)
			if err != nil {
				return nil, err
			}

			// log state transition result
			var recipientLog string
			if res.ContractAddress != "" {
				recipientLog = fmt.Sprintf("contract address %s", res.ContractAddress)
			} else {
				recipientLog = fmt.Sprintf("recipient address %s", msg.Data.Recipient)
			}

			sender := ethcmn.BytesToAddress(msg.GetFrom().Bytes())

			log := fmt.Sprintf(
				"executed EVM state transition; sender address %s; %s", sender, recipientLog,
			)

			k.Logger(ctx).Info(log)
			result.Log = log

			return result, nil

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

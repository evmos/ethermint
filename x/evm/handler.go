package evm

import (
	"github.com/ethereum/go-ethereum/common"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tmtypes "github.com/tendermint/tendermint/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k *Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (result *sdk.Result, err error) {
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
		case types.MsgEthereumTx:
			result, err = handleMsgEthereumTx(ctx, k, msg)
		case types.MsgEthermint:
			result, err = handleMsgEthermint(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
		if err != nil {
			types.CopyCommitStateDB(snapshotStateDB, k.CommitStateDB)
		}
		return result, err
	}
}

// handleMsgEthereumTx handles an Ethereum specific tx
func handleMsgEthereumTx(ctx sdk.Context, k *Keeper, msg types.MsgEthereumTx) (*sdk.Result, error) {
	// execute state transition
	res, err := k.EthereumTx(ctx, msg)
	if err != nil {
		return nil, err
	}

	// log state transition result
	k.Logger(ctx).Info(res.Log)

	return res, nil
}

// handleMsgEthermint handles an sdk.StdTx for an Ethereum state transition
func handleMsgEthermint(ctx sdk.Context, k *Keeper, msg types.MsgEthermint) (*sdk.Result, error) {
	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := common.BytesToHash(txHash)

	st := types.StateTransition{
		AccountNonce: msg.AccountNonce,
		Price:        msg.Price.BigInt(),
		GasLimit:     msg.GasLimit,
		Amount:       msg.Amount.BigInt(),
		Payload:      msg.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       common.BytesToAddress(msg.From.Bytes()),
		Simulate:     ctx.IsCheckTx(),
	}

	if msg.Recipient != nil {
		to := common.BytesToAddress(msg.Recipient.Bytes())
		st.Recipient = &to
	}

	if !st.Simulate {
		// Prepare db for logs
		k.CommitStateDB.Prepare(ethHash, k.TxCount)
		k.TxCount++
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		return nil, err
	}

	// update block bloom filter
	if !st.Simulate {
		k.Bloom.Or(k.Bloom, executionResult.Bloom)

		// update transaction logs in KVStore
		err = k.SetLogs(ctx, common.BytesToHash(txHash), executionResult.Logs)
		if err != nil {
			panic(err)
		}
	}

	// log successful execution
	k.Logger(ctx).Info(executionResult.Result.Log)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthermint,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.From.String()),
		),
	})

	if msg.Recipient != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthermint,
				sdk.NewAttribute(types.AttributeKeyRecipient, msg.Recipient.String()),
			),
		)
	}

	// set the events to the result
	executionResult.Result.Events = ctx.EventManager().Events()
	return executionResult.Result, nil
}

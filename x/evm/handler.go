package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	emint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tmtypes "github.com/tendermint/tendermint/types"
)

// NewHandler returns a handler for Ethermint type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		switch msg := msg.(type) {
		case types.MsgEthereumTx:
			return handleMsgEthereumTx(ctx, k, msg)
		case types.MsgEthermint:
			return handleMsgEthermint(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// handleMsgEthereumTx handles an Ethereum specific tx
func handleMsgEthereumTx(ctx sdk.Context, k Keeper, msg types.MsgEthereumTx) (*sdk.Result, error) {
	// parse the chainID from a string to a base-10 integer
	intChainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return nil, sdkerrors.Wrap(emint.ErrInvalidChainID, ctx.ChainID())
	}

	// Verify signature and retrieve sender address
	sender, err := msg.VerifySig(intChainID)
	if err != nil {
		return nil, err
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := common.BytesToHash(txHash)

	st := types.StateTransition{
		AccountNonce: msg.Data.AccountNonce,
		Price:        msg.Data.Price,
		GasLimit:     msg.Data.GasLimit,
		Recipient:    msg.Data.Recipient,
		Amount:       msg.Data.Amount,
		Payload:      msg.Data.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      intChainID,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		// TODO: block hash
		k.CommitStateDB.Prepare(ethHash, common.Hash{}, k.TxCount)
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

	if !st.Simulate {
		// update block bloom filter
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
			types.EventTypeEthereumTx,
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Data.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	if msg.Data.Recipient != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, msg.Data.Recipient.String()),
			),
		)
	}

	// set the events to the result
	executionResult.Result.Events = ctx.EventManager().Events()
	return executionResult.Result, nil
}

// handleMsgEthermint handles an sdk.StdTx for an Ethereum state transition
func handleMsgEthermint(ctx sdk.Context, k Keeper, msg types.MsgEthermint) (*sdk.Result, error) {
	// parse the chainID from a string to a base-10 integer
	intChainID, ok := new(big.Int).SetString(ctx.ChainID(), 10)
	if !ok {
		return nil, sdkerrors.Wrap(emint.ErrInvalidChainID, ctx.ChainID())
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
		ChainID:      intChainID,
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
		k.CommitStateDB.Prepare(ethHash, common.Hash{}, k.TxCount)
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

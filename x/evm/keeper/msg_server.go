package keeper

import (
	"context"
	"time"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.TypeMsgEthereumTx)

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.WithContext(ctx)

	ethMsg, err := msg.AsMessage()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error())
	}

	var labels []metrics.Label
	if msg.To() == nil {
		labels = []metrics.Label{
			telemetry.NewLabel("execution", "create"),
		}
	} else {
		labels = []metrics.Label{
			telemetry.NewLabel("execution", "call"),
			telemetry.NewLabel("to", msg.Data.To), // recipient address (contract or account)
		}
	}

	sender := ethMsg.From()

	executionResult, err := k.TransitionDb(ethMsg)
	if err != nil {
		return nil, err
	}

	// k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
	// 	Hash:        ethHash.Hex(),
	// 	From:        sender.Hex(),
	// 	Data:        msg.Data,
	// 	Index:       uint64(st.Csdb.TxIndex()),
	// 	BlockHeight: uint64(ctx.BlockHeight()),
	// 	BlockHash:   blockHash.Hex(),
	// 	Result: &types.TxResult{
	// 		ContractAddress: executionResult.Response.ContractAddress,
	// 		Bloom:           executionResult.Response.Bloom,
	// 		TxLogs:          executionResult.Response.TxLogs,
	// 		Ret:             executionResult.Response.Ret,
	// 		Reverted:        executionResult.Response.Reverted,
	// 		GasUsed:         executionResult.GasInfo.GasConsumed,
	// 	},
	// })

	defer func() {
		if ethMsg.Value().IsInt64() {
			telemetry.SetGauge(
				float32(ethMsg.Value().Int64()),
				"tx", "msg", "ethereum_tx",
			)
		}

		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "ethereum_tx"},
			1,
			labels,
		)
	}()

	attrs := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyAmount, ethMsg.Value().String()),
		// sdk.NewAttribute(types.AttributeKeyTxHash, ethcmn.BytesToHash(ethMsg.Hash).Hex()),
	}

	if len(msg.Data.To) > 0 {
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyRecipient, msg.Data.To))
	}

	// emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthereumTx,
			attrs...,
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	return executionResult.Response, nil
}

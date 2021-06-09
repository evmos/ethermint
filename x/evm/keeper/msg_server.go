package keeper

import (
	"context"
	"time"

	"github.com/armon/go-metrics"
	"github.com/palantir/stacktrace"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.TypeMsgEthereumTx)

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.WithContext(ctx)

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

	sender := msg.From
	tx := msg.AsTransaction()

	response, err := k.ApplyTransaction(tx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to apply transaction")
	}

	defer func() {
		if tx.Value().IsInt64() {
			telemetry.SetGauge(
				float32(tx.Value().Int64()),
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
		sdk.NewAttribute(sdk.AttributeKeyAmount, tx.Value().String()),
		// add event for ethereum transaction hash format
		sdk.NewAttribute(types.AttributeKeyEthereumTxHash, response.Hash),
	}

	if len(ctx.TxBytes()) > 0 {
		// add event for tendermint transaction hash format
		hash := tmbytes.HexBytes(tmtypes.Tx(ctx.TxBytes()).Hash())
		attrs = append(attrs, sdk.NewAttribute(types.AttributeKeyTxHash, hash.String()))
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
			sdk.NewAttribute(sdk.AttributeKeySender, sender),
		),
	})

	return response, nil
}

package keeper

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/armon/go-metrics"
	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), types.TypeMsgEthereumTx)

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.CommitStateDB.WithContext(ctx)

	ethMsg, err := msg.AsMessage()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error())
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	var labels []metrics.Label
	if msg.To() == nil {
		labels = []metrics.Label{
			telemetry.NewLabel("execution", "create"),
		}
	} else {
		labels = []metrics.Label{
			telemetry.NewLabel("execution", "call"),
			// add label to the called recipient address (contract or account)
			telemetry.NewLabel("to", msg.Data.To),
		}
	}

	sender := ethMsg.From()

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := ethcmn.BytesToHash(txHash)
	blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())

	st := &types.StateTransition{
		Message:  ethMsg,
		Csdb:     k.CommitStateDB.WithContext(ctx),
		ChainID:  msg.ChainID(),
		TxHash:   &ethHash,
		Simulate: ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		k.CommitStateDB.Prepare(ethHash, blockHash, int(k.GetTxIndexTransient()))
		k.IncreaseTxIndexTransient()
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		if errors.Is(err, vm.ErrExecutionReverted) && executionResult != nil {
			// keep the execution result for revert reason
			executionResult.Response.Reverted = true

			if !st.Simulate {
				k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
					Hash:        ethHash.Hex(),
					From:        sender.Hex(),
					Data:        msg.Data,
					BlockHeight: uint64(ctx.BlockHeight()),
					BlockHash:   blockHash.Hex(),
					Result: &types.TxResult{
						ContractAddress: executionResult.Response.ContractAddress,
						Bloom:           executionResult.Response.Bloom,
						TxLogs:          executionResult.Response.TxLogs,
						Ret:             executionResult.Response.Ret,
						Reverted:        executionResult.Response.Reverted,
						GasUsed:         executionResult.GasInfo.GasConsumed,
					},
				})
			}

			return executionResult.Response, nil
		}

		return nil, err
	}

	if !st.Simulate {
		bloom, found := k.GetBlockBloomTransient()
		if !found {
			bloom = big.NewInt(0)
		}
		// update block bloom filter
		bloom = bloom.Or(bloom, executionResult.Bloom)
		k.SetBlockBloomTransient(bloom)

		// update transaction logs in KVStore
		err = k.CommitStateDB.SetLogs(ethHash, executionResult.Logs)
		if err != nil {
			panic(err)
		}

		blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
		k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
			Hash:        ethHash.Hex(),
			From:        sender.Hex(),
			Data:        msg.Data,
			Index:       uint64(st.Csdb.TxIndex()),
			BlockHeight: uint64(ctx.BlockHeight()),
			BlockHash:   blockHash.Hex(),
			Result: &types.TxResult{
				ContractAddress: executionResult.Response.ContractAddress,
				Bloom:           executionResult.Response.Bloom,
				TxLogs:          executionResult.Response.TxLogs,
				Ret:             executionResult.Response.Ret,
				Reverted:        executionResult.Response.Reverted,
				GasUsed:         executionResult.GasInfo.GasConsumed,
			},
		})

		k.AddTxHashToBlock(ctx, ctx.BlockHeight(), ethHash)
	}

	defer func() {
		if st.Message.Value().IsInt64() {
			telemetry.SetGauge(
				float32(st.Message.Value().Int64()),
				"tx", "msg", "ethereum_tx",
			)
		}

		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "ethereum_tx"},
			1,
			labels,
		)
	}()

	// emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthereumTx,
			sdk.NewAttribute(sdk.AttributeKeyAmount, st.Message.Value().String()),
			sdk.NewAttribute(types.AttributeKeyTxHash, ethcmn.BytesToHash(txHash).Hex()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	if len(msg.Data.To) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, msg.Data.To),
			),
		)
	}

	return executionResult.Response, nil
}

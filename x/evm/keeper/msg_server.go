package keeper

import (
	"context"
	"errors"

	ethcmn "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ethMsg, err := msg.AsMessage()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error())
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
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
		k.Prepare(ctx, ethHash, blockHash, k.TxCount)
		k.TxCount++
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
		// update block bloom filter
		k.Bloom.Or(k.Bloom, executionResult.Bloom)

		// update transaction logs in KVStore
		err = k.SetLogs(ctx, ethHash, executionResult.Logs)
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

		for _, ethLog := range executionResult.Logs {
			k.LogsCache[ethLog.Address] = append(k.LogsCache[ethLog.Address], ethLog)
		}
	}

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

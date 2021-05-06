package keeper

import (
	"context"
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		return nil, err
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		return nil, types.ErrChainConfigNotFound
	}

	ethCfg := config.EthereumConfig(chainIDEpoch)

	blockNum := big.NewInt(ctx.BlockHeight())
	signer := ethtypes.MakeSigner(ethCfg, blockNum)

	ethTx := msg.AsTransaction()

	ethMsg, err := ethTx.AsMessage(signer)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrorInvalidSigner, err.Error())
	}

	sender := ethMsg.From()

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := ethcmn.BytesToHash(txHash)
	blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
	ethBlockHash := ethcmn.BytesToHash(blockHash)

	st := &types.StateTransition{
		Message:  ethMsg,
		Csdb:     k.CommitStateDB.WithContext(ctx),
		ChainID:  chainIDEpoch,
		TxHash:   &ethHash,
		Sender:   sender,
		Simulate: ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		k.Prepare(ctx, ethHash, ethBlockHash, k.TxCount)
		k.TxCount++
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		if err.Error() == "execution reverted" && executionResult != nil {
			// keep the execution result for revert reason
			executionResult.Response.Reverted = true

			if !st.Simulate {
				k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
					Hash:        ethHash.Bytes(),
					From:        sender.Bytes(),
					Data:        msg.Data,
					BlockHeight: uint64(ctx.BlockHeight()),
					BlockHash:   blockHash,
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
			Hash:        ethHash.Bytes(),
			From:        sender.Bytes(),
			Data:        msg.Data,
			Index:       uint64(st.Csdb.TxIndex()),
			BlockHeight: uint64(ctx.BlockHeight()),
			BlockHash:   blockHash,
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
		ethAddr := ethcmn.BytesToAddress(msg.Data.To)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, ethAddr.Hex()),
			),
		)
	}

	return executionResult.Response, nil
}

// func (k *Keeper) SendInternalEthereumTx(
// 	ctx sdk.Context,
// 	payload []byte,
// 	senderAddress common.Address,
// 	recipientAddress common.Address,
// ) (*types.MsgEthereumTxResponse, error) {

// 	accAddress := sdk.AccAddress(senderAddress.Bytes())

// 	acc := k.accountKeeper.GetAccount(ctx, accAddress)
// 	if acc == nil {
// 		acc = k.accountKeeper.NewAccountWithAddress(ctx, accAddress)
// 		k.accountKeeper.SetAccount(ctx, acc)
// 	}

// 	ethAccount, ok := acc.(*ethermint.EthAccount)
// 	if !ok {
// 		return nil, errors.New("could not cast account to EthAccount")
// 	}

// 	if err := ethAccount.SetSequence(ethAccount.GetSequence() + 1); err != nil {
// 		return nil, errors.New("failed to set acc sequence")
// 	}

// 	k.accountKeeper.SetAccount(ctx, ethAccount)

// 	res, err := k.InternalEthereumTx(sdk.WrapSDKContext(ctx), senderAddress, &types.TxData{
// 		Nonce:    ethAccount.GetSequence(),
// 		To:       recipientAddress.Bytes(),
// 		Amount:   big.NewInt(0).Bytes(),
// 		GasPrice: big.NewInt(0).Bytes(),
// 		GasLimit: 10000000, // TODO: don't hardcode, maybe set a limit?
// 		Input:    payload,
// 	})

// 	if err != nil {
// 		err = errors.Wrapf(err, "failed to execute InternalEthereumTx at contract %s", recipientAddress.Hex())
// 		return nil, err
// 	}

// 	return res, nil
// }

// func (k *Keeper) InternalEthereumTx(
// 	goCtx context.Context,
// 	sender common.Address,
// 	tx *types.TxData,
// ) (*types.MsgEthereumTxResponse, error) {
// 	ctx := sdk.UnwrapSDKContext(goCtx)
// 	// parse the chainID from a string to a base-10 integer
// 	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
// 	if err != nil {
// 		return nil, err
// 	}

// 	// true Ethereum tx hash based on its data
// 	ethHash := (&types.MsgEthereumTx{
// 		Data: tx,
// 	}).RLPSignBytes(chainIDEpoch)

// 	var recipient *ethcmn.Address
// 	if len(tx.To) > 0 {
// 		addr := ethcmn.BytesToAddress(tx.To)
// 		recipient = &addr
// 	}

// 	st := &types.StateTransition{
// 		AccountNonce: tx.Nonce,
// 		Price:        new(big.Int).SetBytes(tx.GasPrice),
// 		GasLimit:     tx.GasLimit,
// 		Recipient:    recipient,
// 		Amount:       new(big.Int).SetBytes(tx.Amount),
// 		Payload:      tx.Input,
// 		Csdb:         k.CommitStateDB.WithContext(ctx),
// 		ChainID:      chainIDEpoch,
// 		TxHash:       &ethHash,
// 		Sender:       sender,
// 		Simulate:     ctx.IsCheckTx(),
// 	}

// 	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
// 	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
// 	// other nodes, causing a consensus error
// 	if !st.Simulate {
// 		// Prepare db for logs
// 		hash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
// 		k.Prepare(ctx, ethHash, ethcmn.BytesToHash(hash), k.TxCount)
// 		k.TxCount++
// 	}

// 	config, found := k.GetChainConfig(ctx)
// 	if !found {
// 		return nil, types.ErrChainConfigNotFound
// 	}

// 	executionResult, err := st.TransitionDb(ctx, config)
// 	if err != nil {
// 		if err.Error() == "execution reverted" && executionResult != nil {
// 			// keep the execution result for revert reason
// 			executionResult.Response.Reverted = true

// 			if !st.Simulate {
// 				blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
// 				k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
// 					Hash:        ethHash.Bytes(),
// 					From:        sender.Bytes(),
// 					Data:        tx,
// 					BlockHeight: uint64(ctx.BlockHeight()),
// 					BlockHash:   blockHash,
// 					Result: &types.TxResult{
// 						ContractAddress: executionResult.Response.ContractAddress,
// 						Bloom:           executionResult.Response.Bloom,
// 						TxLogs:          executionResult.Response.TxLogs,
// 						Ret:             executionResult.Response.Ret,
// 						Reverted:        executionResult.Response.Reverted,
// 						GasUsed:         executionResult.GasInfo.GasConsumed,
// 					},
// 				})
// 			}

// 			return executionResult.Response, err
// 		}

// 		return nil, err
// 	}

// 	blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
// 	k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
// 		Hash:        ethHash.Bytes(),
// 		From:        sender.Bytes(),
// 		Data:        tx,
// 		Index:       uint64(st.Csdb.TxIndex()),
// 		BlockHeight: uint64(ctx.BlockHeight()),
// 		BlockHash:   blockHash,
// 		Result: &types.TxResult{
// 			ContractAddress: executionResult.Response.ContractAddress,
// 			Bloom:           executionResult.Response.Bloom,
// 			TxLogs:          executionResult.Response.TxLogs,
// 			Ret:             executionResult.Response.Ret,
// 			Reverted:        executionResult.Response.Reverted,
// 			GasUsed:         executionResult.GasInfo.GasConsumed,
// 		},
// 	})

// 	k.AddTxHashToBlock(ctx, ctx.BlockHeight(), ethHash)

// 	if !st.Simulate {
// 		// update block bloom filter
// 		k.Bloom.Or(k.Bloom, executionResult.Bloom)

// 		// update transaction logs in KVStore
// 		err = k.SetLogs(ctx, ethHash, executionResult.Logs)
// 		if err != nil {
// 			panic(err)
// 		}

// 		for _, ethLog := range executionResult.Logs {
// 			k.LogsCache[ethLog.Address] = append(k.LogsCache[ethLog.Address], ethLog)
// 		}
// 	}

// 	// emit events
// 	ctx.EventManager().EmitEvents(sdk.Events{
// 		sdk.NewEvent(
// 			types.EventTypeEthereumTx,
// 			sdk.NewAttribute(sdk.AttributeKeyAmount, st.Message.Value().String()),
// 			sdk.NewAttribute(types.AttributeKeyTxHash, ethHash.Hex()),
// 		),
// 		sdk.NewEvent(
// 			sdk.EventTypeMessage,
// 			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
// 			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
// 		),
// 	})

// 	if len(tx.To) > 0 {
// 		ethAddr := ethcmn.BytesToAddress(tx.To)
// 		ctx.EventManager().EmitEvent(
// 			sdk.NewEvent(
// 				types.EventTypeEthereumTx,
// 				sdk.NewAttribute(types.AttributeKeyRecipient, ethAddr.Hex()),
// 			),
// 		)
// 	}

// 	return executionResult.Response, nil
// }

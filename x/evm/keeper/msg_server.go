package keeper

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcmn "github.com/ethereum/go-ethereum/common"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/ethermint/metrics"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) EthereumTx(goCtx context.Context, msg *types.MsgEthereumTx) (*types.MsgEthereumTxResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(goCtx)

	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	// Verify signature and retrieve sender address
	var homesteadErr error
	sender, eip155Err := msg.VerifySig(chainIDEpoch)
	if eip155Err != nil {
		sender, homesteadErr = msg.VerifySigHomestead()
		if homesteadErr != nil {
			log.WithFields(log.Fields{
				"eip155_err":    eip155Err.Error(),
				"homestead_err": homesteadErr.Error(),
			}).Warningln("failed to verify signatures with EIP155 and Homestead signers")

			metrics.ReportFuncError(k.svcTags)
			return nil, errors.New("no valid signatures")
		}
	}

	txHash := tmtypes.Tx(ctx.TxBytes()).Hash()
	ethHash := ethcmn.BytesToHash(txHash)

	var recipient *ethcmn.Address
	if len(msg.Data.Recipient) > 0 {
		addr := ethcmn.BytesToAddress(msg.Data.Recipient)
		recipient = &addr
	}

	st := &types.StateTransition{
		AccountNonce: msg.Data.AccountNonce,
		Price:        new(big.Int).SetBytes(msg.Data.Price),
		GasLimit:     msg.Data.GasLimit,
		Recipient:    recipient,
		Amount:       new(big.Int).SetBytes(msg.Data.Amount),
		Payload:      msg.Data.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		hash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
		k.Prepare(ctx, ethHash, ethcmn.BytesToHash(hash), k.TxCount)
		k.TxCount++
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		metrics.ReportFuncError(k.svcTags)
		return nil, types.ErrChainConfigNotFound
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		if err.Error() == "execution reverted" && executionResult != nil {
			// keep the execution result for revert reason
			executionResult.Response.Reverted = true
			metrics.ReportFuncError(k.svcTags)

			if !st.Simulate {
				blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
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

		metrics.ReportFuncError(k.svcTags)
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
			sdk.NewAttribute(sdk.AttributeKeyAmount, st.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyTxHash, ethcmn.BytesToHash(txHash).Hex()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	if len(msg.Data.Recipient) > 0 {
		ethAddr := ethcmn.BytesToAddress(msg.Data.Recipient)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, ethAddr.Hex()),
			),
		)
	}

	metrics.ReportFuncError(k.svcTags)
	return executionResult.Response, nil
}

func (k *Keeper) SendInternalEthereumTx(
	ctx sdk.Context,
	payload []byte,
	senderAddress common.Address,
	recipientAddress common.Address,
) (*types.MsgEthereumTxResponse, error) {

	accAddress := sdk.AccAddress(senderAddress.Bytes())

	acc := k.accountKeeper.GetAccount(ctx, accAddress)
	if acc == nil {
		acc = k.accountKeeper.NewAccountWithAddress(ctx, accAddress)
		k.accountKeeper.SetAccount(ctx, acc)
	}

	ethAccount, ok := acc.(*ethermint.EthAccount)
	if !ok {
		return nil, errors.New("could not cast account to EthAccount")
	}

	if err := ethAccount.SetSequence(ethAccount.GetSequence() + 1); err != nil {
		return nil, errors.New("failed to set acc sequence")
	}

	k.accountKeeper.SetAccount(ctx, ethAccount)

	res, err := k.InternalEthereumTx(sdk.WrapSDKContext(ctx), senderAddress, &types.TxData{
		AccountNonce: ethAccount.GetSequence(),
		Recipient:    recipientAddress.Bytes(),
		Amount:       big.NewInt(0).Bytes(),
		Price:        big.NewInt(0).Bytes(),
		GasLimit:     10000000, // TODO: don't hardcode, maybe set a limit?
		Payload:      payload,
	})

	if err != nil {
		err = errors.Wrapf(err, "failed to execute InternalEthereumTx at contract %s", recipientAddress.Hex())
		return nil, err
	}

	return res, nil
}

func (k *Keeper) InternalEthereumTx(
	goCtx context.Context,
	sender common.Address,
	tx *types.TxData,
) (*types.MsgEthereumTxResponse, error) {
	metrics.ReportFuncCall(k.svcTags)
	doneFn := metrics.ReportFuncTiming(k.svcTags)
	defer doneFn()

	ctx := sdk.UnwrapSDKContext(goCtx)
	// parse the chainID from a string to a base-10 integer
	chainIDEpoch, err := ethermint.ParseChainID(ctx.ChainID())
	if err != nil {
		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	// true Ethereum tx hash based on its data
	ethHash := (&types.MsgEthereumTx{
		Data: tx,
	}).RLPSignBytes(chainIDEpoch)

	var recipient *ethcmn.Address
	if len(tx.Recipient) > 0 {
		addr := ethcmn.BytesToAddress(tx.Recipient)
		recipient = &addr
	}

	st := &types.StateTransition{
		AccountNonce: tx.AccountNonce,
		Price:        new(big.Int).SetBytes(tx.Price),
		GasLimit:     tx.GasLimit,
		Recipient:    recipient,
		Amount:       new(big.Int).SetBytes(tx.Amount),
		Payload:      tx.Payload,
		Csdb:         k.CommitStateDB.WithContext(ctx),
		ChainID:      chainIDEpoch,
		TxHash:       &ethHash,
		Sender:       sender,
		Simulate:     ctx.IsCheckTx(),
	}

	// since the txCount is used by the stateDB, and a simulated tx is run only on the node it's submitted to,
	// then this will cause the txCount/stateDB of the node that ran the simulated tx to be different than the
	// other nodes, causing a consensus error
	if !st.Simulate {
		// Prepare db for logs
		hash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
		k.Prepare(ctx, ethHash, ethcmn.BytesToHash(hash), k.TxCount)
		k.TxCount++
	}

	config, found := k.GetChainConfig(ctx)
	if !found {
		metrics.ReportFuncError(k.svcTags)
		return nil, types.ErrChainConfigNotFound
	}

	executionResult, err := st.TransitionDb(ctx, config)
	if err != nil {
		if err.Error() == "execution reverted" && executionResult != nil {
			// keep the execution result for revert reason
			executionResult.Response.Reverted = true
			metrics.ReportFuncError(k.svcTags)

			if !st.Simulate {
				blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
				k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
					Hash:        ethHash.Bytes(),
					From:        sender.Bytes(),
					Data:        tx,
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

			return executionResult.Response, err
		}

		metrics.ReportFuncError(k.svcTags)
		return nil, err
	}

	blockHash, _ := k.GetBlockHashFromHeight(ctx, ctx.BlockHeight())
	k.SetTxReceiptToHash(ctx, ethHash, &types.TxReceipt{
		Hash:        ethHash.Bytes(),
		From:        sender.Bytes(),
		Data:        tx,
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

	if !st.Simulate {
		// update block bloom filter
		k.Bloom.Or(k.Bloom, executionResult.Bloom)

		// update transaction logs in KVStore
		err = k.SetLogs(ctx, ethHash, executionResult.Logs)
		if err != nil {
			panic(err)
		}

		for _, ethLog := range executionResult.Logs {
			k.LogsCache[ethLog.Address] = append(k.LogsCache[ethLog.Address], ethLog)
		}
	}

	// emit events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEthereumTx,
			sdk.NewAttribute(sdk.AttributeKeyAmount, st.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyTxHash, ethHash.Hex()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
		),
	})

	if len(tx.Recipient) > 0 {
		ethAddr := ethcmn.BytesToAddress(tx.Recipient)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEthereumTx,
				sdk.NewAttribute(types.AttributeKeyRecipient, ethAddr.Hex()),
			),
		)
	}

	return executionResult.Response, nil
}

package keeper_test

import (
	"bytes"
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
)

// LogRecordHook records all the logs
type LogRecordHook struct {
	Logs []*ethtypes.Log
}

func (dh *LogRecordHook) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	dh.Logs = receipt.Logs
	return nil
}

// FailureHook always fail
type FailureHook struct{}

func (dh FailureHook) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return errors.New("post tx processing failed")
}

type ReceiptChangingHook struct{}

func (dh ReceiptChangingHook) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	// Change original receipt
	receipt.BlockHash = common.BytesToHash([]byte("dirtyHash"))
	receipt.Status = ethtypes.ReceiptStatusFailed
	return nil
}

func (suite *KeeperTestSuite) TestEvmHooks() {
	testCases := []struct {
		msg            string
		setupHook      func() types.EvmHooks
		expFunc        func(hook types.EvmHooks, result error)
		receiptChanged bool
	}{
		{
			"receipt can be changed by EVM Hook",
			func() types.EvmHooks {
				return &ReceiptChangingHook{}
			},
			func(hook types.EvmHooks, result error) {
				suite.Require().NoError(result)
			},
			true,
		},
		{
			"log collect hook",
			func() types.EvmHooks {
				return &LogRecordHook{}
			},
			func(hook types.EvmHooks, result error) {
				suite.Require().NoError(result)
				suite.Require().Equal(1, len((hook.(*LogRecordHook).Logs)))
			},
			false,
		},
		{
			"always fail hook",
			func() types.EvmHooks {
				return &FailureHook{}
			},
			func(hook types.EvmHooks, result error) {
				suite.Require().Error(result)
			},
			false,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()
		hook := tc.setupHook()
		suite.app.EvmKeeper.SetHooks(keeper.NewMultiEvmHooks(hook))

		k := suite.app.EvmKeeper
		ctx := suite.ctx
		txHash := common.BigToHash(big.NewInt(1))
		headerHash := common.BytesToHash(ctx.HeaderHash().Bytes())
		originalStatus := ethtypes.ReceiptStatusSuccessful
		vmdb := statedb.New(ctx, k, statedb.NewTxConfig(
			headerHash,
			txHash,
			0,
			0,
		))

		vmdb.AddLog(&ethtypes.Log{
			Topics:  []common.Hash{},
			Address: suite.address,
		})
		logs := vmdb.Logs()
		receipt := &ethtypes.Receipt{
			BlockHash: headerHash,
			Status:    originalStatus,
			TxHash:    txHash,
			Logs:      logs,
		}
		// Deep copy receipt to originalReceipt before PostTxProcessing
		originalReceipt := new(ethtypes.Receipt)
		*originalReceipt = *receipt

		result := k.PostTxProcessing(ctx, ethtypes.Message{}, receipt)
		tc.expFunc(hook, result)
		if !tc.receiptChanged {
			receiptBeforeHook, err := originalReceipt.MarshalBinary()
			suite.Require().NoError(err)
			receiptAfterHook, err := receipt.MarshalBinary()
			suite.Require().NoError(err)
			suite.Require().True(bytes.Equal(receiptBeforeHook, receiptAfterHook))
		}
	}
}

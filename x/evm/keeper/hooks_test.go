package keeper_test

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/ethermint/x/evm/types"
)

// LogRecordHook records all the logs
type LogRecordHook struct {
	Logs []*ethtypes.Log
}

func (dh *LogRecordHook) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	dh.Logs = logs
	return nil
}

// FailureHook always fail
type FailureHook struct{}

func (dh FailureHook) PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error {
	return errors.New("post tx processing failed")
}

func (suite *KeeperTestSuite) TestEvmHooks() {
	suite.SetupTest()
	suite.Commit()

	logRecordHook := LogRecordHook{}
	suite.app.EvmKeeper.SetHooks(keeper.NewMultiEvmHooks(&logRecordHook))

	k := suite.app.EvmKeeper

	txHash := common.BigToHash(big.NewInt(1))

	amt := sdk.Coins{ethermint.NewPhotonCoinInt64(100)}
	err := suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, amt)
	suite.Require().NoError(err)
	err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, suite.address.Bytes(), amt)
	suite.Require().NoError(err)

	k.SetTxHashTransient(txHash)
	k.AddLog(&ethtypes.Log{
		Topics:  []common.Hash{},
		Address: suite.address,
	})

	logs := k.GetTxLogs(txHash)
	suite.Require().Equal(1, len(logs))

	err = k.PostTxProcessing(txHash, logs)
	suite.Require().NoError(err)

	suite.Require().Equal(1, len(logRecordHook.Logs))
}

func (suite *KeeperTestSuite) TestHookFailure() {
	suite.SetupTest()
	k := suite.app.EvmKeeper

	// Test failure hook
	suite.app.EvmKeeper.SetHooks(keeper.NewMultiEvmHooks(FailureHook{}))
	err := k.PostTxProcessing(common.Hash{}, nil)
	suite.Require().Error(err)
}

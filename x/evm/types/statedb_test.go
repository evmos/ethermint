package types_test

import (
	"math/big"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/x/evm/keeper"

	abci "github.com/tendermint/tendermint/abci/types"
)

// nolint: unused
type StateDBTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier sdk.Querier
	app     *app.EthermintApp
}

func (suite *StateDBTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.querier = keeper.NewQuerier(suite.app.EvmKeeper)
}

func (suite *StateDBTestSuite) TestBloomFilter() {
	stateDB := suite.app.EvmKeeper.CommitStateDB

	// Prepare db for logs
	tHash := ethcmn.BytesToHash([]byte{0x1})
	stateDB.Prepare(tHash, ethcmn.Hash{}, 0)

	contractAddress := ethcmn.BigToAddress(big.NewInt(1))

	// Generate and add a log to test
	log := ethtypes.Log{Address: contractAddress}
	stateDB.AddLog(&log)

	// Get log from db
	logs, err := stateDB.GetLogs(tHash)
	suite.Require().NoError(err)
	suite.Require().Equal(len(logs), 1)

	// get logs bloom from the log
	bloomInt := ethtypes.LogsBloom(logs)
	bloomFilter := ethtypes.BytesToBloom(bloomInt.Bytes())

	// Check to make sure bloom filter will succeed on
	suite.Require().True(ethtypes.BloomLookup(bloomFilter, contractAddress))
	suite.Require().False(ethtypes.BloomLookup(bloomFilter, ethcmn.BigToAddress(big.NewInt(2))))
}

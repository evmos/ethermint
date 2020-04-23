package keeper_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/x/evm/keeper"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

const addrHex = "0x756F45E3FA69347A9A973A725E3C98bC4db0b4c1"

var address = ethcmn.HexToAddress(addrHex)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier sdk.Querier
	app     *app.EthermintApp
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, abci.Header{Height: 1, ChainID: "3", Time: time.Now().UTC()})
	suite.querier = keeper.NewQuerier(suite.app.EvmKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestDBStorage() {
	// Perform state transitions
	suite.app.EvmKeeper.CreateAccount(suite.ctx, address)
	suite.app.EvmKeeper.SetBalance(suite.ctx, address, big.NewInt(5))
	suite.app.EvmKeeper.SetNonce(suite.ctx, address, 4)
	suite.app.EvmKeeper.SetState(suite.ctx, address, ethcmn.HexToHash("0x2"), ethcmn.HexToHash("0x3"))
	suite.app.EvmKeeper.SetCode(suite.ctx, address, []byte{0x1})

	// Test block hash mapping functionality
	suite.app.EvmKeeper.SetBlockHashMapping(suite.ctx, ethcmn.FromHex("0x0d87a3a5f73140f46aac1bf419263e4e94e87c292f25007700ab7f2060e2af68"), 7)
	height, err := suite.app.EvmKeeper.GetBlockHashMapping(suite.ctx, ethcmn.FromHex("0x0d87a3a5f73140f46aac1bf419263e4e94e87c292f25007700ab7f2060e2af68"))
	suite.Require().NoError(err)
	suite.Require().Equal(int64(7), height)

	suite.app.EvmKeeper.SetBlockHashMapping(suite.ctx, []byte{0x43, 0x32}, 8)

	// Test block height mapping functionality
	testBloom := ethtypes.BytesToBloom([]byte{0x1, 0x3})
	suite.app.EvmKeeper.SetBlockBloomMapping(suite.ctx, testBloom, 4)

	// Get those state transitions
	suite.Require().Equal(suite.app.EvmKeeper.GetBalance(suite.ctx, address).Cmp(big.NewInt(5)), 0)
	suite.Require().Equal(suite.app.EvmKeeper.GetNonce(suite.ctx, address), uint64(4))
	suite.Require().Equal(suite.app.EvmKeeper.GetState(suite.ctx, address, ethcmn.HexToHash("0x2")), ethcmn.HexToHash("0x3"))
	suite.Require().Equal(suite.app.EvmKeeper.GetCode(suite.ctx, address), []byte{0x1})

	height, err = suite.app.EvmKeeper.GetBlockHashMapping(suite.ctx, ethcmn.FromHex("0x0d87a3a5f73140f46aac1bf419263e4e94e87c292f25007700ab7f2060e2af68"))
	suite.Require().NoError(err)
	suite.Require().Equal(height, int64(7))
	height, err = suite.app.EvmKeeper.GetBlockHashMapping(suite.ctx, []byte{0x43, 0x32})
	suite.Require().NoError(err)
	suite.Require().Equal(height, int64(8))

	bloom, err := suite.app.EvmKeeper.GetBlockBloomMapping(suite.ctx, 4)
	suite.Require().NoError(err)
	suite.Require().Equal(bloom, testBloom)

	// commit stateDB
	_, err = suite.app.EvmKeeper.Commit(suite.ctx, false)
	suite.Require().NoError(err, "failed to commit StateDB")

	// simulate BaseApp EndBlocker commitment
	suite.app.Commit()
}

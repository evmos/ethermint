package keeper_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/ethermint/app"
	ethermint "github.com/cosmos/ethermint/types"
	"github.com/cosmos/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const addrHex = "0x756F45E3FA69347A9A973A725E3C98bC4db0b4c1"
const hex = "0x0d87a3a5f73140f46aac1bf419263e4e94e87c292f25007700ab7f2060e2af68"

var (
	hash = ethcmn.FromHex(hex)
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.EthermintApp
	queryClient types.QueryClient
	address     ethcmn.Address
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: "ethermint-3", Time: time.Now().UTC()})
	suite.address = ethcmn.HexToAddress(addrHex)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.EvmKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)

	balance := ethermint.NewPhotonCoin(sdk.ZeroInt())
	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    ethcrypto.Keccak256(nil),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	suite.app.BankKeeper.SetBalance(suite.ctx, acc.GetAddress(), balance)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestTransactionLogs() {
	ethHash := ethcmn.BytesToHash(hash)
	log := &ethtypes.Log{
		Address:     suite.address,
		Data:        []byte("log"),
		BlockNumber: 10,
	}
	log2 := &ethtypes.Log{
		Address:     suite.address,
		Data:        []byte("log2"),
		BlockNumber: 11,
	}
	expLogs := []*ethtypes.Log{log}

	err := suite.app.EvmKeeper.SetLogs(suite.ctx, ethHash, expLogs)
	suite.Require().NoError(err)

	logs, err := suite.app.EvmKeeper.GetLogs(suite.ctx, ethHash)
	suite.Require().NoError(err)
	suite.Require().Equal(expLogs, logs)

	expLogs = []*ethtypes.Log{log2, log}

	// add another log under the zero hash
	suite.app.EvmKeeper.AddLog(suite.ctx, log2)
	logs = suite.app.EvmKeeper.AllLogs(suite.ctx)
	suite.Require().Equal(expLogs, logs)

	// add another log under the zero hash
	log3 := &ethtypes.Log{
		Address:     suite.address,
		Data:        []byte("log3"),
		BlockNumber: 10,
	}
	suite.app.EvmKeeper.AddLog(suite.ctx, log3)

	txLogs := suite.app.EvmKeeper.GetAllTxLogs(suite.ctx)
	suite.Require().Equal(2, len(txLogs))

	suite.Require().Equal(ethcmn.Hash{}.String(), txLogs[0].Hash)
	suite.Require().Equal([]*ethtypes.Log{log2, log3}, txLogs[0].Logs)

	suite.Require().Equal(ethHash.String(), txLogs[1].Hash)
	suite.Require().Equal([]*ethtypes.Log{log}, txLogs[1].Logs)
}

func (suite *KeeperTestSuite) TestDBStorage() {
	// Perform state transitions
	suite.app.EvmKeeper.CreateAccount(suite.ctx, suite.address)
	suite.app.EvmKeeper.SetBalance(suite.ctx, suite.address, big.NewInt(5))
	suite.app.EvmKeeper.SetNonce(suite.ctx, suite.address, 4)
	suite.app.EvmKeeper.SetState(suite.ctx, suite.address, ethcmn.HexToHash("0x2"), ethcmn.HexToHash("0x3"))
	suite.app.EvmKeeper.SetCode(suite.ctx, suite.address, []byte{0x1})

	// Test block height mapping functionality
	testBloom := ethtypes.BytesToBloom([]byte{0x1, 0x3})
	suite.app.EvmKeeper.SetBlockBloom(suite.ctx, 4, testBloom)

	// Get those state transitions
	suite.Require().Equal(suite.app.EvmKeeper.GetBalance(suite.ctx, suite.address).Cmp(big.NewInt(5)), 0)
	suite.Require().Equal(suite.app.EvmKeeper.GetNonce(suite.ctx, suite.address), uint64(4))
	suite.Require().Equal(suite.app.EvmKeeper.GetState(suite.ctx, suite.address, ethcmn.HexToHash("0x2")), ethcmn.HexToHash("0x3"))
	suite.Require().Equal(suite.app.EvmKeeper.GetCode(suite.ctx, suite.address), []byte{0x1})

	bloom, found := suite.app.EvmKeeper.GetBlockBloom(suite.ctx, 4)
	suite.Require().True(found)
	suite.Require().Equal(bloom, testBloom)

	// commit stateDB
	_, err := suite.app.EvmKeeper.Commit(suite.ctx, false)
	suite.Require().NoError(err, "failed to commit StateDB")

	// simulate BaseApp EndBlocker commitment
	suite.app.Commit()
}

func (suite *KeeperTestSuite) TestChainConfig() {
	config, found := suite.app.EvmKeeper.GetChainConfig(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(types.DefaultChainConfig(), config)

	config.EIP150Block = sdk.NewInt(100)
	suite.app.EvmKeeper.SetChainConfig(suite.ctx, config)
	newConfig, found := suite.app.EvmKeeper.GetChainConfig(suite.ctx)
	suite.Require().True(found)
	suite.Require().Equal(config, newConfig)
}

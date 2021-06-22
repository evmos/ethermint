package keeper_test

import (
	"github.com/tharsis/ethermint/crypto/ethsecp256k1"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/tharsis/ethermint/app"
	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/ethermint/x/evm/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
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
	consAddress sdk.ConsAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false

	suite.app = app.Setup(checkTx)
	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, ChainID: "ethermint-1", Time: time.Now().UTC()})
	suite.app.EvmKeeper.WithContext(suite.ctx)

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

	priv, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)
	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
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

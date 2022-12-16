package keeper_test

import (
	_ "embed"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Entangle-Protocol/entangle-blockchain/app"
	"github.com/Entangle-Protocol/entangle-blockchain/crypto/ethsecp256k1"
	"github.com/Entangle-Protocol/entangle-blockchain/encoding"
	"github.com/Entangle-Protocol/entangle-blockchain/tests"
	ethermint "github.com/Entangle-Protocol/entangle-blockchain/types"
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/keeper"
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	evmtypes "github.com/Entangle-Protocol/entangle-blockchain/x/evm/types"

	// moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	// distributorsauthtypes "github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.EthermintApp
	queryClient types.QueryClient
	address     common.Address
	consAddress sdk.ConsAddress

	// for generate test tx
	clientCtx client.Context
	ethSigner ethtypes.Signer

	msgServer types.MsgServer

	appCodec codec.Codec
	signer   keyring.Signer
	denom    string
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keeper Suite")
}

// SetupTest setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	suite.app = app.Setup(checkTx, nil)
	suite.SetupApp(checkTx)
}

func (suite *KeeperTestSuite) SetupApp(checkTx bool) {
	t := suite.T()
	// Create MsgServiceRouter, but don't populate it before creating the gov
	// keeper.
	// msr := baseapp.NewMsgServiceRouter()

	// encCfg := moduletestutil.MakeTestEncodingConfig()

	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.address = common.BytesToAddress(priv.PubKey().Address().Bytes())
	suite.signer = tests.NewSigner(priv)

	// consensus key
	priv, err = ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	suite.consAddress = sdk.ConsAddress(priv.PubKey().Address())

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:          1,
		ChainID:         "ethermint_9000-1",
		Time:            time.Now().UTC(),
		ProposerAddress: suite.consAddress.Bytes(),
		Version: tmversion.Consensus{
			Block: version.BlockProtocol,
		},
		LastBlockId: tmproto.BlockID{
			Hash: tmhash.Sum([]byte("block_id")),
			PartSetHeader: tmproto.PartSetHeader{
				Total: 11,
				Hash:  tmhash.Sum([]byte("partset_header")),
			},
		},
		AppHash:            tmhash.Sum([]byte("app")),
		DataHash:           tmhash.Sum([]byte("data")),
		EvidenceHash:       tmhash.Sum([]byte("evidence")),
		ValidatorsHash:     tmhash.Sum([]byte("validators")),
		NextValidatorsHash: tmhash.Sum([]byte("next_validators")),
		ConsensusHash:      tmhash.Sum([]byte("consensus")),
		LastResultsHash:    tmhash.Sum([]byte("last_result")),
	})

	// msr.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.DistributorsAuthKeeper)
	// types.RegisterMsgServer(suite.app.MsgServiceRouter(), keeper.NewMsgServerImpl(suite.app.DistributorsAuthKeeper))
	suite.queryClient = types.NewQueryClient(queryHelper)
	suite.msgServer = keeper.NewMsgServerImpl(suite.app.DistributorsAuthKeeper)

	acc := &ethermint.EthAccount{
		BaseAccount: authtypes.NewBaseAccount(sdk.AccAddress(suite.address.Bytes()), nil, 0, 0),
		CodeHash:    common.BytesToHash(crypto.Keccak256(nil)).String(),
	}

	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

	valAddr := sdk.ValAddress(suite.address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, priv.PubKey(), stakingtypes.Description{})
	require.NoError(t, err)
	validator = stakingkeeper.TestingUpdateValidator(suite.app.StakingKeeper, suite.ctx, validator, true)
	err = suite.app.StakingKeeper.AfterValidatorCreated(suite.ctx, validator.GetOperator())
	require.NoError(t, err)

	err = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	require.NoError(t, err)
	suite.app.StakingKeeper.SetValidator(suite.ctx, validator)

	encodingConfig := encoding.MakeConfig(app.ModuleBasics)
	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	suite.ethSigner = ethtypes.LatestSignerForChainID(suite.app.EvmKeeper.ChainID())
	suite.appCodec = encodingConfig.Codec
	suite.denom = evmtypes.DefaultEVMDenom
}

// Commit commits and starts a new block with an updated context.
func (suite *KeeperTestSuite) Commit() {
	suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *KeeperTestSuite) CommitAfter(t time.Duration) {
	header := suite.ctx.BlockHeader()
	suite.app.EndBlock(abci.RequestEndBlock{Height: header.Height})
	_ = suite.app.Commit()

	header.Height += 1
	header.Time = header.Time.Add(t)
	suite.app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// update ctx
	suite.ctx = suite.app.BaseApp.NewContext(false, header)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.DistributorsAuthKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) TestSetGetDistributor() {
	testCases := []struct {
		name     string
		malleate func()
		address  string
		end_date uint64
	}{
		{
			"distributor with end date in future",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EndDate: uint64(1234567890123)})
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			uint64(1234567890123),
		},
		{
			"distributor with end date in past",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: "ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8", EndDate: uint64(1674125669)})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			uint64(1674125669),
		},
		{
			"distributor with end date 1",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: "ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc", EndDate: uint64(1)})
			},
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			uint64(1),
		},
		{
			"distributor with end date 0",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: "ethm1hhcu6yx67x67ykt4cp47g28t3m0jvcps3p3rdk", EndDate: uint64(0)})
			},
			"ethm1hhcu6yx67x67ykt4cp47g28t3m0jvcps3p3rdk",
			uint64(0),
		},
		{
			"distributor with empty address",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: "", EndDate: uint64(0)})
			},
			"",
			uint64(0),
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		distributor, err := suite.app.DistributorsAuthKeeper.GetDistributor(suite.ctx, tc.address)
		// if tc.address != "" {
		suite.Require().NoError(err)
		suite.Require().Equal(tc.address, distributor.Address, tc.name)
		suite.Require().Equal(tc.end_date, distributor.EndDate, tc.name)
		// } else {
		// 	suite.Require().Error(err)
		// }

	}
}

func (suite *KeeperTestSuite) TestSetGetAdmin() {
	testCases := []struct {
		name        string
		malleate    func()
		address     string
		edit_option bool
	}{
		{
			"admin with edit_option true",
			func() {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EditOption: true})
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			true,
		},
		{
			"admin with edit_option false",
			func() {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EditOption: false})
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			false,
		},
		{
			"empty admin address",
			func() {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: "", EditOption: true})
			},
			"",
			true,
		},
		{
			"wrong admin address",
			func() {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: "ethm1hhcu6yx67x", EditOption: true})
			},
			"ethm1hhcu6yx67x",
			true,
		},
	}
	for _, tc := range testCases {
		tc.malleate()

		admin, err := suite.app.DistributorsAuthKeeper.GetAdmin(suite.ctx, tc.address)
		// if tc.address != "" {
		suite.Require().NoError(err)
		suite.Require().Equal(tc.address, admin.Address, tc.name)
		suite.Require().Equal(tc.edit_option, admin.EditOption, tc.name)
		// } else {
		// 	suite.Require().Error(err)
		// }

	}
}

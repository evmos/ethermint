package evm_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/app"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/version"
	"testing"
	"time"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.EthermintApp
}

var s *KeeperTestSuite

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	suite.Run(t, s)

}

// SetupApp setup test environment, it uses`require.TestingT` to support both `testing.T` and `testing.B`.
func (suite *KeeperTestSuite) SetupAppWithT(checkTx bool, t require.TestingT) {
	suite.app = app.Setup(checkTx, func(app *app.EthermintApp, genesis simapp.GenesisState) simapp.GenesisState {
		return genesis
	})

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{
		Height:  1,
		ChainID: "ethermint_9000-1",
		Time:    time.Now().UTC(),
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

}

func (suite *KeeperTestSuite) SetupTestWithT(t require.TestingT) {
	checkTx := false
	suite.app = app.Setup(checkTx, nil)
	suite.SetupAppWithT(checkTx, t)
}

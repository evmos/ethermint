package keeper_test

import (
	"github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	em := suite.ctx.EventManager()
	suite.Require().Equal(0, len(em.Events()))

	res := suite.app.EvmKeeper.EndBlock(suite.ctx, types.RequestEndBlock{})
	suite.Require().Equal([]types.ValidatorUpdate{}, res)

	// should emit 1 EventTypeBlockBloom event on EndBlock
	suite.Require().Equal(1, len(em.Events()))
	suite.Require().Equal("ethermint.evm.v1.EventBlockBloom", em.Events()[0].Type)
}

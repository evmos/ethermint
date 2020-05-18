package evm_test

import (
	"github.com/cosmos/ethermint/x/evm"
	"github.com/cosmos/ethermint/x/evm/types"
)

func (suite *EvmTestSuite) TestExportImport() {
	var genState types.GenesisState
	suite.Require().NotPanics(func() {
		genState = evm.ExportGenesis(suite.ctx, suite.app.EvmKeeper, suite.app.AccountKeeper)
	})

	_ = evm.InitGenesis(suite.ctx, suite.app.EvmKeeper, genState)
}

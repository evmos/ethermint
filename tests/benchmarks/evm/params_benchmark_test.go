package evm_test

import (
	"github.com/evmos/ethermint/x/evm/types"
	"testing"
)

func BenchmarkSetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTestWithT(b)
	for i := 0; i < b.N; i++ {
		params := types.DefaultParams()
		suite.app.EvmKeeper.SetParams(suite.ctx, params)
	}
}

func BenchmarkGetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTestWithT(b)
	for i := 0; i < b.N; i++ {
		suite.app.EvmKeeper.GetParams(suite.ctx)
	}
}

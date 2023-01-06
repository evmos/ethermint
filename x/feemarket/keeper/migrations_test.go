package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	feemarketkeeper "github.com/evmos/ethermint/x/feemarket/keeper"
	v4types "github.com/evmos/ethermint/x/feemarket/migrations/v4/types"
	"github.com/evmos/ethermint/x/feemarket/types"
)

type mockSubspace struct {
	ps v4types.Params
}

func newMockSubspace(ps v4types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSetIfExists(_ sdk.Context, ps types.LegacyParams) {
	*ps.(*v4types.Params) = ms.ps
}

func (suite *KeeperTestSuite) TestMigrations() {
	legacySubspace := newMockSubspace(v4types.DefaultParams())
	migrator := feemarketkeeper.NewMigrator(suite.app.FeeMarketKeeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		{
			"Run Migrate3to4",
			migrator.Migrate3to4,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			err := tc.migrateFunc(suite.ctx)
			suite.Require().NoError(err)
		})
	}
}

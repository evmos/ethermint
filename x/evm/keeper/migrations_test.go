package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmkeeper "github.com/evmos/ethermint/x/evm/keeper"
	v4types "github.com/evmos/ethermint/x/evm/migrations/v4/types"
	"github.com/evmos/ethermint/x/evm/types"
)

type mockSubspace struct {
	ps v4types.Params
}

func newMockSubspace(ps v4types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps types.LegacyParams) {
	*ps.(*v4types.Params) = ms.ps
}

func (suite *KeeperTestSuite) TestMigrations() {
	legacySubspace := newMockSubspace(v4types.DefaultParams())
	migrator := evmkeeper.NewMigrator(*suite.app.EvmKeeper, legacySubspace)

	testCases := []struct {
		name        string
		migrateFunc func(ctx sdk.Context) error
	}{
		// TODO: Figure out if these will be deleted
		//{
		//	"Run Migrate1to2",
		//	migrator.Migrate1to2,
		//},
		//{
		//	"Run Migrate2to3",
		//	migrator.Migrate2to3,
		//},
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

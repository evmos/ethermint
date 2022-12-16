package keeper_test

import (
	"fmt"

	distributorsauthtypes "github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	"github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) TestEndBlock() {
	testCases := []struct {
		name     string
		malleate func()
		found    bool
		address  string
	}{
		{
			"No empty Distributors",
			func() {},
			false,
			"",
		},
		{
			"No correct Distributors",
			func() {},
			false,
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
		},
		{
			"Distributor with timer in future",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, distributorsauthtypes.DistributorInfo{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EndDate: uint64(1234567890123)})
			},
			true,
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
		},
		{
			"Distributor with 0 timer",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, distributorsauthtypes.DistributorInfo{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EndDate: uint64(0)})
			},
			true,
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
		},
		{
			"Distributor remove",
			func() {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, distributorsauthtypes.DistributorInfo{Address: "ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww", EndDate: uint64(100)})
			},
			false,
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.name), func() {
			suite.SetupTest() // reset

			tc.malleate()
			suite.app.DistributorsAuthKeeper.EndBlock(suite.ctx, types.RequestEndBlock{Height: 1})
			distributor, err := suite.app.DistributorsAuthKeeper.GetDistributor(suite.ctx, tc.address)
			if tc.found {
				suite.Require().Equal(tc.address, distributor.Address, tc.name)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

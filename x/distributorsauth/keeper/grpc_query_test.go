package keeper_test

import (
	types "github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
)

// func (suite *KeeperTestSuite) TestQueryParams() {
// 	testCases := []struct {
// 		name    string
// 		expPass bool
// 	}{
// 		{
// 			"pass",
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		params := suite.app.FeeMarketKeeper.GetParams(suite.ctx)
// 		exp := &types.QueryParamsResponse{Params: params}

// 		res, err := suite.queryClient.Params(suite.ctx.Context(), &types.QueryParamsRequest{})
// 		if tc.expPass {
// 			suite.Require().Equal(exp, res, tc.name)
// 			suite.Require().NoError(err)
// 		} else {
// 			suite.Require().Error(err)
// 		}
// 	}
// }

// func (suite *KeeperTestSuite) TestQueryBaseFee() {
// 	var (
// 		aux    sdkmath.Int
// 		expRes *types.QueryBaseFeeResponse
// 	)

// 	testCases := []struct {
// 		name     string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"pass - default Base Fee",
// 			func() {
// 				initialBaseFee := sdkmath.NewInt(ethparams.InitialBaseFee)
// 				expRes = &types.QueryBaseFeeResponse{BaseFee: &initialBaseFee}
// 			},
// 			true,
// 		},
// 		{
// 			"pass - non-nil Base Fee",
// 			func() {
// 				baseFee := sdk.OneInt().BigInt()
// 				suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, baseFee)

// 				aux = sdkmath.NewIntFromBigInt(baseFee)
// 				expRes = &types.QueryBaseFeeResponse{BaseFee: &aux}
// 			},
// 			true,
// 		},
// 	}
// 	for _, tc := range testCases {
// 		tc.malleate()

// 		res, err := suite.queryClient.BaseFee(suite.ctx.Context(), &types.QueryBaseFeeResponse{})
// 		if tc.expPass {
// 			suite.Require().NotNil(res)
// 			suite.Require().Equal(expRes, res, tc.name)
// 			suite.Require().NoError(err)
// 		} else {
// 			suite.Require().Error(err)
// 		}
// 	}
// }

func (suite *KeeperTestSuite) TestQueryDistributor() {
	var (
		expRes *types.QueryDistributorResponse
	)

	testCases := []struct {
		name     string
		malleate func(string, uint64)
		address  string
		end_date uint64
	}{
		{
			"pass",
			func(address string, end_date uint64) {
				distr := types.DistributorInfo{
					Address: address,
					EndDate: end_date,
				}
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, distr)

				expRes = &types.QueryDistributorResponse{Distributor: distr}
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			uint64(123),
		},
	}
	for _, tc := range testCases {
		tc.malleate(tc.address, tc.end_date)

		res, err := suite.queryClient.Distributor(suite.ctx.Context(), &types.QueryDistributorRequest{DistributorAddr: tc.address})
		suite.Require().Equal(expRes, res, tc.name)
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) TestQueryDistributors() {
	var (
		expRes *types.QueryDistributorsResponse
	)

	testCases := []struct {
		name     string
		malleate func(string, uint64)
		address  string
		end_date uint64
	}{
		{
			"pass with 0 distributor",
			func(address string, end_date uint64) {
				expRes = &types.QueryDistributorsResponse{}
			},
			"",
			uint64(0),
		},
		{
			"pass with 1 distributor",
			func(address string, end_date uint64) {
				distr := types.DistributorInfo{
					Address: address,
					EndDate: end_date,
				}
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, distr)

				expRes = &types.QueryDistributorsResponse{Distributors: []types.DistributorInfo{distr}}
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			uint64(123),
		},
	}
	for _, tc := range testCases {
		tc.malleate(tc.address, tc.end_date)

		res, err := suite.queryClient.Distributors(suite.ctx.Context(), &types.QueryDistributorsRequest{})
		suite.Require().Equal(expRes, res, tc.name)
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) TestQueryAdmin() {
	var (
		expRes *types.QueryAdminResponse
	)

	testCases := []struct {
		name        string
		malleate    func(string, bool)
		address     string
		edit_option bool
	}{
		{
			"pass with edit option true",
			func(address string, edit_option bool) {
				admin := types.Admin{
					Address:    address,
					EditOption: edit_option,
				}
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, admin)

				expRes = &types.QueryAdminResponse{Admin: admin}
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			true,
		},
		{
			"pass with edit option false",
			func(address string, edit_option bool) {
				admin := types.Admin{
					Address:    address,
					EditOption: edit_option,
				}
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, admin)

				expRes = &types.QueryAdminResponse{Admin: admin}
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			true,
		},
	}
	for _, tc := range testCases {
		suite.SetupTest() // reset

		tc.malleate(tc.address, tc.edit_option)

		res, err := suite.queryClient.Admin(suite.ctx.Context(), &types.QueryAdminRequest{AdminAddr: tc.address})
		suite.Require().Equal(expRes, res, tc.name)
		suite.Require().NoError(err)
	}
}

func (suite *KeeperTestSuite) TestQueryAdmins() {
	var (
		expRes *types.QueryAdminsResponse
	)

	testCases := []struct {
		name     string
		malleate func(string, bool)
		address  string
		end_date bool
	}{
		{
			"pass with 0 admins",
			func(address string, edit_option bool) {
				expRes = &types.QueryAdminsResponse{}
			},
			"",
			true,
		},
		{
			"pass with 1 admin",
			func(address string, edit_option bool) {
				admin := types.Admin{
					Address:    address,
					EditOption: edit_option,
				}
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, admin)

				expRes = &types.QueryAdminsResponse{Admins: []types.Admin{admin}}
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			true,
		},
		{
			"pass with 1 admin edit_option false",
			func(address string, edit_option bool) {
				admin := types.Admin{
					Address:    address,
					EditOption: edit_option,
				}
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, admin)

				expRes = &types.QueryAdminsResponse{Admins: []types.Admin{admin}}
			},
			"ethm1tjm23pl06ja8zgag08q2vt8smrnyds9yzkx7ww",
			false,
		},
	}
	for _, tc := range testCases {
		tc.malleate(tc.address, tc.end_date)

		res, err := suite.queryClient.Admins(suite.ctx.Context(), &types.QueryAdminsRequest{})
		// if tc.expPass {
		suite.Require().Equal(expRes, res, tc.name)
		suite.Require().NoError(err)
		// } else {
		// 	suite.Require().Error(err)
		// }
	}
}

// Distributors(ctx context.Context, in *QueryDistributorsRequest, opts ...grpc.CallOption) (*QueryDistributorsResponse, error)
// 	// Queries distributor info for given distributor address.
// 	Distributor(ctx context.Context, in *QueryDistributorRequest, opts ...grpc.CallOption) (*QueryDistributorResponse, error)
// 	// Queries admin info for all admins
// 	Admins(ctx context.Context, in *QueryAdminsRequest, opts ...grpc.CallOption) (*QueryAdminsResponse, error)
// 	// Queries admin info for given admin address.
// 	Admin(ctx context.Context, in *QueryAdminRequest, opts ...grpc.CallOption) (*QueryAdminResponse, error)

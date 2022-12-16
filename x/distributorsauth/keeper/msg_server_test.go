package keeper_test

import (
	"github.com/Entangle-Protocol/entangle-blockchain/x/distributorsauth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func (suite *KeeperTestSuite) TestAddDistributor() {

	testCases := []struct {
		name                string
		malleate            func(string)
		sender              string
		distributor_address string
		end_date            uint64
		success             bool
	}{
		{
			"Add distributor success by Admin",
			func(addr string) {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: false})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			uint64(1234),
			true,
		},
		{
			"Add distributor failed by Distributor",
			func(addr string) {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			uint64(1234),
			false,
		},
		{
			"Add distributor failed by guest",
			func(addr string) {},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			uint64(1234),
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate(tc.sender)
			_, add_err := suite.msgServer.AddDistributor(suite.ctx, &types.MsgAddDistributor{Sender: tc.sender, DistributorAddress: tc.distributor_address, EndDate: tc.end_date})

			distr, err := suite.app.DistributorsAuthKeeper.GetDistributor(suite.ctx, tc.distributor_address)
			if !tc.success {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)
			suite.Require().NoError(add_err)
			suite.Require().Equal(distr, types.DistributorInfo{Address: tc.distributor_address, EndDate: tc.end_date})
		})
	}
}

func (suite *KeeperTestSuite) TestAddAdmin() {

	testCases := []struct {
		name          string
		malleate      func(string)
		sender        string
		admin_address string
		edit_option   bool
		success       bool
	}{
		{
			"Add Admin success by Admin with edit wrights",
			func(addr string) {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: true})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			true,
		},
		{
			"Add admin failed by Admin withitout edit wrights",
			func(addr string) {
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: false})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
		{
			"Add admin failed by Distributor",
			func(addr string) {
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
		{
			"Add admin failed by guest",
			func(addr string) {},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate(tc.sender)
			_, add_err := suite.msgServer.AddAdmin(suite.ctx, &types.MsgAddAdmin{Sender: tc.sender, AdminAddress: tc.admin_address, EditOption: tc.edit_option})

			admin, err := suite.app.DistributorsAuthKeeper.GetAdmin(suite.ctx, tc.admin_address)
			if !tc.success {
				suite.Require().Error(err)
				return
			}
			suite.Require().NoError(err)
			suite.Require().NoError(add_err)
			suite.Require().Equal(admin, types.Admin{Address: tc.admin_address, EditOption: tc.edit_option})
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveAdmin() {

	testCases := []struct {
		name                    string
		malleate                func(string, string, bool)
		sender                  string
		admin_to_delete_address string
		edit_option             bool
		success                 bool
	}{
		{
			"Remove Admin success by Admin with edit wrights",
			func(sender string, addr string, edit_option bool) {
				/// Admin sender
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: sender, EditOption: edit_option})
				/// Admin to remove
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: edit_option})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			true,
		},
		{
			"Remove admin failed by Admin withitout edit wrights",
			func(sender string, addr string, edit_option bool) {
				/// Admin sender
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: sender, EditOption: edit_option})
				/// Admin to remove
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: edit_option})

			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			false,
			false,
		},
		{
			"Remove admin failed by Distributor",
			func(sender string, addr string, edit_option bool) {
				/// Distributor sender
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: sender, EndDate: uint64(0)})
				/// Admin to remove
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: edit_option})

			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
		{
			"Remove admin failed by guest",
			func(sender string, addr string, edit_option bool) {
				/// Admin to remove
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: addr, EditOption: edit_option})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate(tc.sender, tc.admin_to_delete_address, tc.edit_option)
			_, add_err := suite.msgServer.RemoveAdmin(suite.ctx, &types.MsgRemoveAdmin{Sender: tc.sender, AdminAddress: tc.admin_to_delete_address})

			admin, err := suite.app.DistributorsAuthKeeper.GetAdmin(suite.ctx, tc.admin_to_delete_address)
			if tc.success {
				suite.Require().NoError(add_err)
				suite.Require().Error(err)
				return
			}

			suite.Require().Error(add_err)
			suite.Require().NoError(err)

			suite.Require().Equal(admin, types.Admin{Address: tc.admin_to_delete_address, EditOption: tc.edit_option})
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveDistributor() {

	testCases := []struct {
		name                string
		malleate            func(string, string, bool)
		sender              string
		distributor_address string
		edit_option         bool
		success             bool
	}{
		{
			"Remove Distributor success by Admin with edit wrights",
			func(sender string, addr string, edit_option bool) {
				/// Admin sender
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: sender, EditOption: edit_option})
				/// Distributor to remove
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			true,
		},
		{
			"Remove Distributor success by Admin withitout edit wrights",
			func(sender string, addr string, edit_option bool) {
				/// Admin sender
				suite.app.DistributorsAuthKeeper.AddAdmin(suite.ctx, types.Admin{Address: sender, EditOption: edit_option})
				/// Distributor to remove
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})

			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			false,
			true,
		},
		{
			"Remove Distributor success by Gov module as a sender",
			func(sender string, addr string, edit_option bool) {
				/// Distributor to remove
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})

			},
			authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			false,
			true,
		},
		{
			"Remove Distributor failed by Distributor",
			func(sender string, addr string, edit_option bool) {
				/// Distributor sender
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: sender, EndDate: uint64(0)})
				/// Distributor to remove
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})

			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
		{
			"Remove Distributor failed by guest",
			func(sender string, addr string, edit_option bool) {
				/// Distributor to remove
				suite.app.DistributorsAuthKeeper.AddDistributor(suite.ctx, types.DistributorInfo{Address: addr, EndDate: uint64(0)})
			},
			"ethm1cdsdkvxydypnhtec5y880qdtdexcu2ehf0lpv8",
			"ethm1trhgn3un9wqlxhxwxspxaaalnynr4539v8vdmc",
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate(tc.sender, tc.distributor_address, tc.edit_option)
			_, add_err := suite.msgServer.RemoveDistributor(suite.ctx, &types.MsgRemoveDistributor{Sender: tc.sender, DistributorAddress: tc.distributor_address})

			distributor, err := suite.app.DistributorsAuthKeeper.GetDistributor(suite.ctx, tc.distributor_address)
			if tc.success {
				suite.Require().NoError(add_err)
				suite.Require().Error(err)
				return
			}

			suite.Require().Error(add_err)
			suite.Require().NoError(err)

			suite.Require().Equal(distributor, types.DistributorInfo{Address: tc.distributor_address, EndDate: uint64(0)})
		})
	}
}

package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/ethermint/tests"
)

type AccessListTestSuite struct {
	suite.Suite

	addr    common.Address
	hexAddr string
}

func (suite *AccessListTestSuite) SetupTest() {
	suite.addr = tests.GenerateAddress()
	suite.hexAddr = suite.addr.Hex()
}

func TestAccessListTestSuite(t *testing.T) {
	suite.Run(t, new(AccessListTestSuite))
}

func (suite *AccessListTestSuite) TestTestNewAccessList() {
	testCases := []struct {
		name          string
		ethAccessList *ethtypes.AccessList
		expAl         AccessList
	}{
		{
			"ethAccessList is nil",
			nil,
			nil,
		},
		{
			"non-empty ethAccessList",
			&ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}},
			AccessList{{Address: suite.hexAddr, StorageKeys: []string{common.Hash{}.Hex()}}},
		},
	}
	for _, tc := range testCases {
		al := NewAccessList(tc.ethAccessList)

		suite.Require().Equal(tc.expAl, al)
	}
}

func (suite *AccessListTestSuite) TestAccessListToEthAccessList() {
	ethAccessList := ethtypes.AccessList{{Address: suite.addr, StorageKeys: []common.Hash{{0}}}}
	al := NewAccessList(&ethAccessList)
	actual := al.ToEthAccessList()

	suite.Require().Equal(&ethAccessList, actual)
}

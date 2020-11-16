package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	ethcmn "github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
)

type AccessListTestSuite struct {
	suite.Suite

	address    ethcmn.Address
	accessList *accessList
}

func (suite *AccessListTestSuite) SetupTest() {
	privkey, err := ethsecp256k1.GenerateKey()
	suite.Require().NoError(err)

	suite.address = ethcmn.BytesToAddress(privkey.PubKey().Address().Bytes())
	suite.accessList = newAccessList()
	suite.accessList.addresses[suite.address] = 1
}

func TestAccessListTestSuite(t *testing.T) {
	suite.Run(t, new(AccessListTestSuite))
}

func (suite *AccessListTestSuite) TestContainsAddress() {
	found := suite.accessList.ContainsAddress(suite.address)
	suite.Require().True(found)
}

func (suite *AccessListTestSuite) TestContains() {
	testCases := []struct {
		name           string
		malleate       func()
		expAddrPresent bool
		expSlotPresent bool
	}{
		{"out of range", func() {}, true, false},
		{
			"address, no slots",
			func() {
				suite.accessList.addresses[suite.address] = -1
			}, true, false,
		},
		{
			"no address, no slots",
			func() {
				delete(suite.accessList.addresses, suite.address)
			}, false, false,
		},
		{
			"address, slot not present",
			func() {
				suite.accessList.addresses[suite.address] = 0
				suite.accessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
			}, true, false,
		},
		{
			"address, slots",
			func() {
				suite.accessList.addresses[suite.address] = 0
				suite.accessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
				suite.accessList.slots[0] = make(map[ethcmn.Hash]struct{})
				suite.accessList.slots[0][ethcmn.Hash{}] = struct{}{}
			}, true, true,
		},
	}

	for _, tc := range testCases {
		tc.malleate()

		addrPresent, slotPresent := suite.accessList.Contains(suite.address, ethcmn.Hash{})

		suite.Require().Equal(tc.expAddrPresent, addrPresent, tc.name)
		suite.Require().Equal(tc.expSlotPresent, slotPresent, tc.name)
	}
}

func (suite *AccessListTestSuite) TestCopy() {
	expAccessList := newAccessList()

	testCases := []struct {
		name     string
		malleate func()
	}{
		{"empty", func() {
			expAccessList.slots = make([]map[ethcmn.Hash]struct{}, 0)
		}},
		{
			"single address", func() {
				expAccessList = newAccessList()
				expAccessList.slots = make([]map[ethcmn.Hash]struct{}, 0)
				expAccessList.addresses[suite.address] = -1
			},
		},
		{
			"single address, single slot",
			func() {
				expAccessList = newAccessList()
				expAccessList.addresses[suite.address] = 0
				expAccessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
				expAccessList.slots[0] = make(map[ethcmn.Hash]struct{})
				expAccessList.slots[0][ethcmn.Hash{}] = struct{}{}
			},
		},
		{
			"multiple addresses, single slot each",
			func() {
				expAccessList = newAccessList()
				expAccessList.slots = make([]map[ethcmn.Hash]struct{}, 10)
				for i := 0; i < 10; i++ {
					expAccessList.addresses[ethcmn.BytesToAddress([]byte(fmt.Sprintf("%d", i)))] = i
					expAccessList.slots[i] = make(map[ethcmn.Hash]struct{})
					expAccessList.slots[i][ethcmn.BytesToHash([]byte(fmt.Sprintf("%d", i)))] = struct{}{}
				}
			},
		},
		{
			"multiple addresses, multiple slots each",
			func() {
				expAccessList = newAccessList()
				expAccessList.slots = make([]map[ethcmn.Hash]struct{}, 10)
				for i := 0; i < 10; i++ {
					expAccessList.addresses[ethcmn.BytesToAddress([]byte(fmt.Sprintf("%d", i)))] = i
					expAccessList.slots[i] = make(map[ethcmn.Hash]struct{})
					for j := 0; j < 10; j++ {
						expAccessList.slots[i][ethcmn.BytesToHash([]byte(fmt.Sprintf("%d-%d", i, j)))] = struct{}{}
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		tc.malleate()

		accessList := expAccessList.Copy()
		suite.Require().EqualValues(expAccessList, accessList, tc.name)
	}
}

func (suite *AccessListTestSuite) TestAddAddress() {
	testCases := []struct {
		name    string
		address ethcmn.Address
		ok      bool
	}{
		{"already present", suite.address, false},
		{"ok", ethcmn.Address{}, true},
	}

	for _, tc := range testCases {
		ok := suite.accessList.AddAddress(tc.address)
		suite.Require().Equal(tc.ok, ok, tc.name)
	}
}

func (suite *AccessListTestSuite) TestAddSlot() {
	testCases := []struct {
		name          string
		malleate      func()
		expAddrChange bool
		expSlotChange bool
	}{
		{"out of range", func() {}, false, false},
		{
			"address not present added, slot added",
			func() {
				delete(suite.accessList.addresses, suite.address)
			}, true, true,
		},
		{
			"address present, slot not present added",
			func() {
				suite.accessList.addresses[suite.address] = 0
				suite.accessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
				suite.accessList.slots[0] = make(map[ethcmn.Hash]struct{})
			}, false, true,
		},
		{
			"address present, slot present",
			func() {
				suite.accessList.addresses[suite.address] = 0
				suite.accessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
				suite.accessList.slots[0] = make(map[ethcmn.Hash]struct{})
				suite.accessList.slots[0][ethcmn.Hash{}] = struct{}{}
			}, false, false,
		},
	}

	for _, tc := range testCases {
		tc.malleate()

		addrChange, slotChange := suite.accessList.AddSlot(suite.address, ethcmn.Hash{})
		suite.Require().Equal(tc.expAddrChange, addrChange, tc.name)
		suite.Require().Equal(tc.expSlotChange, slotChange, tc.name)
	}
}

func (suite *AccessListTestSuite) TestDeleteSlot() {
	testCases := []struct {
		name     string
		malleate func()
		expPanic bool
	}{
		{"panics, out of range", func() {}, true},
		{"panics, address not found", func() {
			delete(suite.accessList.addresses, suite.address)
		}, true},
		{
			"single slot present",
			func() {
				suite.accessList.addresses[suite.address] = 0
				suite.accessList.slots = make([]map[ethcmn.Hash]struct{}, 1)
				suite.accessList.slots[0] = make(map[ethcmn.Hash]struct{})
				suite.accessList.slots[0][ethcmn.Hash{}] = struct{}{}
			}, false,
		},
	}

	for _, tc := range testCases {
		tc.malleate()

		if tc.expPanic {
			suite.Require().Panics(func() {
				suite.accessList.DeleteSlot(suite.address, ethcmn.Hash{})
			}, tc.name)
		} else {
			suite.Require().NotPanics(func() {
				suite.accessList.DeleteSlot(suite.address, ethcmn.Hash{})
			}, tc.name)
		}
	}
}

package types_test

import (
	"math/big"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

func (suite *StateDBTestSuite) TestStateObject_State() {
	testCase := []struct {
		name     string
		key      ethcmn.Hash
		expValue ethcmn.Hash
		malleate func()
	}{
		{
			"no set value, load from KVStore",
			ethcmn.BytesToHash([]byte("key")),
			ethcmn.Hash{},
			func() {},
		},
		{
			"no-op SetState",
			ethcmn.BytesToHash([]byte("key")),
			ethcmn.Hash{},
			func() {
				suite.stateObject.SetState(nil, ethcmn.BytesToHash([]byte("key")), ethcmn.Hash{})
			},
		},
		{
			"cached value",
			ethcmn.BytesToHash([]byte("key1")),
			ethcmn.BytesToHash([]byte("value1")),
			func() {
				suite.stateObject.SetState(nil, ethcmn.BytesToHash([]byte("key1")), ethcmn.BytesToHash([]byte("value1")))
			},
		},
		{
			"update value",
			ethcmn.BytesToHash([]byte("key1")),
			ethcmn.BytesToHash([]byte("value2")),
			func() {
				suite.stateObject.SetState(nil, ethcmn.BytesToHash([]byte("key1")), ethcmn.BytesToHash([]byte("value2")))
			},
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		value := suite.stateObject.GetState(nil, tc.key)
		suite.Require().Equal(tc.expValue, value, tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateObject_AddBalance() {
	testCase := []struct {
		name       string
		amount     *big.Int
		expBalance *big.Int
	}{
		{"zero amount", big.NewInt(0), big.NewInt(0)},
		{"positive amount", big.NewInt(10), big.NewInt(10)},
		{"negative amount", big.NewInt(-1), big.NewInt(9)},
	}

	for _, tc := range testCase {
		suite.stateObject.AddBalance(tc.amount)
		suite.Require().Equal(tc.expBalance, suite.stateObject.Balance(), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateObject_SubBalance() {
	testCase := []struct {
		name       string
		amount     *big.Int
		expBalance *big.Int
	}{
		{"zero amount", big.NewInt(0), big.NewInt(0)},
		{"negative amount", big.NewInt(-10), big.NewInt(10)},
		{"positive amount", big.NewInt(1), big.NewInt(9)},
	}

	for _, tc := range testCase {
		suite.stateObject.SubBalance(tc.amount)
		suite.Require().Equal(tc.expBalance, suite.stateObject.Balance(), tc.name)
	}
}

func (suite *StateDBTestSuite) TestStateObject_Code() {
	testCase := []struct {
		name     string
		expCode  []byte
		malleate func()
	}{
		{
			"cached code",
			[]byte("code"),
			func() {
				suite.stateObject.SetCode(ethcmn.BytesToHash([]byte("code_hash")), []byte("code"))
			},
		},
		{
			"empty code hash",
			nil,
			func() {
				suite.stateObject.SetCode(ethcmn.Hash{}, nil)
			},
		},
		{
			"empty code",
			nil,
			func() {
				suite.stateObject.SetCode(ethcmn.BytesToHash([]byte("code_hash")), nil)
			},
		},
	}

	for _, tc := range testCase {
		tc.malleate()

		code := suite.stateObject.Code(nil)
		suite.Require().Equal(tc.expCode, code, tc.name)
	}
}

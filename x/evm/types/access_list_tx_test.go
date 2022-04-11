package types

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"
	"github.com/tharsis/ethermint/tests"
)

type AccessListTxTestSuite struct {
	suite.Suite

	sdkInt         sdk.Int
	uint64         uint64
	bigInt         *big.Int
	overflowBigInt *big.Int
	sdkZeroInt     sdk.Int
	sdkMinusOneInt sdk.Int
	invalidAddr    string
	addr           common.Address
	hexAddr        string
}

func (suite *AccessListTxTestSuite) SetupTest() {
	suite.sdkInt = sdk.NewInt(100)
	suite.uint64 = suite.sdkInt.Uint64()
	suite.bigInt = big.NewInt(1)
	suite.overflowBigInt = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(256), nil)
	suite.sdkZeroInt = sdk.ZeroInt()
	suite.sdkMinusOneInt = sdk.NewInt(-1)
	suite.invalidAddr = "123456"
	suite.addr = tests.GenerateAddress()
	suite.hexAddr = suite.addr.Hex()
}

func TestAccessListTxTestSuite(t *testing.T) {
	suite.Run(t, new(AccessListTxTestSuite))
}

func (suite *AccessListTxTestSuite) TestAccessListTxCopy() {
	tx := &AccessListTx{}
	txCopy := tx.Copy()

	suite.Require().Equal(&AccessListTx{}, txCopy)
	// TODO: Test for different pointers
}

func (suite *AccessListTxTestSuite) TestAccessListTxGetGasTipCap() {
	testCases := []struct {
		name string
		tx   AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			AccessListTx{
				GasPrice: &suite.sdkInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasTipCap()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *AccessListTxTestSuite) TestAccessListTxGetGasFeeCap() {
	testCases := []struct {
		name string
		tx   AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			AccessListTx{
				GasPrice: &suite.sdkInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasFeeCap()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *AccessListTxTestSuite) TestAccessListTxCost() {
	testCases := []struct {
		name string
		tx   AccessListTx
		exp  *big.Int
	}{
		{
			"non-empty access list tx",
			AccessListTx{
				GasPrice: &suite.sdkInt,
				GasLimit: uint64(1),
				Amount:   &suite.sdkZeroInt,
			},
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.Cost()

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *AccessListTxTestSuite) TestAccessListTxEffectiveCost() {
	testCases := []struct {
		name    string
		tx      AccessListTx
		baseFee *big.Int
		exp     *big.Int
	}{
		{
			"non-empty access list tx",
			AccessListTx{
				GasPrice: &suite.sdkInt,
				GasLimit: uint64(1),
				Amount:   &suite.sdkZeroInt,
			},
			(&suite.sdkInt).BigInt(),
			(&suite.sdkInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.EffectiveCost(tc.baseFee)

		suite.Require().Equal(tc.exp, actual, tc.name)
	}
}

func (suite *AccessListTxTestSuite) TestAccessListTxType() {
	testCases := []struct {
		name string
		tx   AccessListTx
	}{
		{
			"non-empty access list tx",
			AccessListTx{},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.TxType()

		suite.Require().Equal(uint8(ethtypes.AccessListTxType), actual, tc.name)
	}
}

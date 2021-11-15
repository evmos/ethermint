package types

import (
	"testing"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/suite"
)

type ParamsTestSuite struct {
	suite.Suite
}

func (suite *ParamsTestSuite) SetupTest() {

}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}

func (suite *ParamsTestSuite) TestParamKeyTable() {
	suite.Require().IsType(paramtypes.KeyTable{}, ParamKeyTable())
}

func (suite *ParamsTestSuite) TestParamsValidate() {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid",
			NewParams(true, 7, 3, 2000000000, int64(544435345345435345)),
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"base fee change denominator is 0 ",
			NewParams(true, 0, 3, 2000000000, int64(544435345345435345)),
			true,
		},
		{
			"initial base fee cannot is negative",
			NewParams(true, 7, 3, -2000000000, int64(544435345345435345)),
			true,
		},
		{
			"initial base fee cannot is negative",
			NewParams(true, 7, 3, 2000000000, int64(-544435345345435345)),
			true,
		},
		// {
		// 	"invalid eip",
		// 	Params{
		// 		EvmDenom:  "stake",
		// 		ExtraEIPs: []int64{1},
		// 	},
		// 	true,
		// },
		// {
		// 	"invalid chain config",
		// 	NewParams("ara", true, true, ChainConfig{}, 2929, 1884, 1344),
		// 	false,
		// },
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

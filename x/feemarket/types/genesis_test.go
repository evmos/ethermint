package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type GenesisTestSuite struct {
	suite.Suite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestValidateGenesis() {
	testCases := []struct {
		name     string
		genState *GenesisState
		expPass  bool
	}{
		{
			"default",
			DefaultGenesisState(),
			true,
		},
		{
			"valid genesis",
			&GenesisState{
				DefaultParams(),
				sdk.ZeroInt(),
				uint64(1),
			},
			true,
		},
		{
			"valid New genesis",
			NewGenesisState(
				DefaultParams(),
				sdk.ZeroInt(),
				uint64(1),
			),
			true,
		},
		{
			"empty genesis",
			&GenesisState{
				Params:   Params{},
				BaseFee:  sdk.ZeroInt(),
				BlockGas: 0,
			},
			false,
		},
		{
			"base fee is negative",
			&GenesisState{
				Params:   Params{},
				BaseFee:  sdk.OneInt().Neg(),
				BlockGas: 0,
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.genState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

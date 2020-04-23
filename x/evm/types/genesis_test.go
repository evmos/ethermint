package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		msg      string
		genstate GenesisState
		expPass  bool
	}{
		{
			msg:      "pass with defaultState ",
			genstate: DefaultGenesisState(),
			expPass:  true,
		},
		{
			msg: "empty address",
			genstate: GenesisState{
				Accounts: []GenesisAccount{{}},
			},
			expPass: false,
		},
		{
			msg: "empty balance",
			genstate: GenesisState{
				Accounts: []GenesisAccount{{Balance: nil}},
			},
			expPass: false,
		},
	}
	for i, tc := range testCases {

		err := ValidateGenesis(tc.genstate)
		if tc.expPass {
			require.NoError(t, err, "test (%d) %s", i, tc.msg)
		} else {
			require.Error(t, err, "test (%d): %s", i, tc.msg)
		}
	}

}

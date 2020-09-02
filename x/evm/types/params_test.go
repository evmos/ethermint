package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		params   Params
		expError bool
	}{
		{"default", DefaultParams(), false},
		{
			"valid",
			NewParams("ara"),
			false,
		},
		{
			"empty",
			Params{},
			true,
		},
		{
			"invalid evm denom",
			Params{
				EvmDenom: "@!#!@$!@5^32",
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.params.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
		} else {
			require.NoError(t, err, tc.name)
		}
	}
}

func TestParamsValidatePriv(t *testing.T) {
	require.Error(t, validateEVMDenom(false))
}

func TestParams_String(t *testing.T) {
	require.Equal(t, "evm_denom: aphoton\n", DefaultParams().String())
}

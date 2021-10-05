package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyTxValidate(t *testing.T) {
	testCases := []struct {
		name     string
		tx       LegacyTx
		expError bool
	}{
		{
			"empty",
			LegacyTx{},
			true,
		},
		{
			"gas price is nil",
			LegacyTx{
				GasPrice: nil,
			},
			true,
		},
		{
			"gas price is negative",
			LegacyTx{
				GasPrice: &minusOneInt,
			},
			true,
		},
		{
			"amount is negative",
			LegacyTx{
				GasPrice: &hundredInt,
				Amount:   &minusOneInt,
			},
			true,
		},
	}

	for _, tc := range testCases {
		err := tc.tx.Validate()

		if tc.expError {
			require.Error(t, err, tc.name)
			continue
		}

		require.NoError(t, err, tc.name)
	}
}

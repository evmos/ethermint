package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/x/evm/types"
)

func TestLegacyTxValidate(t *testing.T) {
	testCases := []struct {
		name     string
		tx       types.LegacyTx
		expError bool
	}{
		{
			"empty",
			types.LegacyTx{},
			true,
		},
		{
			"gas price is nil",
			types.LegacyTx{
				GasPrice: nil,
			},
			true,
		},
		{
			"gas price is negative",
			types.LegacyTx{
				GasPrice: &minusOneInt,
			},
			true,
		},
		{
			"amount is negative",
			types.LegacyTx{
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

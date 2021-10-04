package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/x/evm/types"
)

func TestDynamicFeeTxValidate(t *testing.T) {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	minusOneInt := sdk.NewInt(-1)

	testCases := []struct {
		name     string
		tx       types.DynamicFeeTx
		expError bool
	}{
		{
			"empty",
			types.DynamicFeeTx{},
			true,
		},
		{
			"gas tip cap is nil",
			types.DynamicFeeTx{
				GasTipCap: nil,
			},
			true,
		},
		{
			"gas fee cap is nil",
			types.DynamicFeeTx{
				GasTipCap: &zeroInt,
			},
			true,
		},
		{
			"gas tip cap is negative",
			types.DynamicFeeTx{
				GasTipCap: &minusOneInt,
				GasFeeCap: &zeroInt,
			},
			true,
		},
		{
			"gas tip cap is negative",
			types.DynamicFeeTx{
				GasTipCap: &zeroInt,
				GasFeeCap: &minusOneInt,
			},
			true,
		},
		{
			"gas fee cap < gas tip cap",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &zeroInt,
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

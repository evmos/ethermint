package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyTxSetSignatureValues(t *testing.T) {
	testCases := []struct {
		name string
		v    *big.Int
		r    *big.Int
		s    *big.Int
	}{
		{
			"non-empty values",
			hundredbigInt,
			hundredbigInt,
			hundredbigInt,
		},
	}
	for _, tc := range testCases {
		tx := &LegacyTx{}
		tx.SetSignatureValues(nil, tc.v, tc.r, tc.s)

		v, r, s := tx.GetRawSignatureValues()

		require.Equal(t, tc.v, v, tc.name)
		require.Equal(t, tc.r, r, tc.name)
		require.Equal(t, tc.s, s, tc.name)
	}
}

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
		{
			"to address is invalid",
			LegacyTx{
				GasPrice: &hundredInt,
				Amount:   &hundredInt,
				To:       invalidAddr,
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

func TestLegacyTxFeeCost(t *testing.T) {
	tx := &LegacyTx{}

	require.Panics(t, func() { tx.Fee() }, "should panice")
	require.Panics(t, func() { tx.Cost() }, "should panice")
}

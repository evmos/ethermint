package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/types"
)

func TestGetRawSignatureValues(t *testing.T) {
	hundredInt := sdk.NewInt(100)
	validAddr := tests.GenerateAddress().Hex()

	testCases := []struct {
		name    string
		tx      types.DynamicFeeTx
		expPass bool
	}{
		{
			"success",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &hundredInt,
				To:        validAddr,
				ChainID:   &hundredInt,
			},
			true,
		},
	}

	for _, tc := range testCases {
		values := tc.tx.GetRawSignatureValues()

		if tc.expPass {
			require.NoError(t, values, tc.name)
			continue
		}

		require.Error(t, values, tc.name)
	}
}

// func TestSetSignatureValues(t *testing.T) {
// 	hundredInt := sdk.NewInt(100)
// 	validAddr := tests.GenerateAddress().Hex()
// 	chainID := big.NewInt(0)

// 	testCases := []struct {
// 		name    string
// 		tx      types.DynamicFeeTx
// 		chainID *big.Int
// 		r       *big.Int
// 		v       *big.Int
// 		s       *big.Int
// 		expPass bool
// 	}{
// 		{
// 			"chainID",
// 			types.DynamicFeeTx{
// 				GasTipCap: &hundredInt,
// 				GasFeeCap: &hundredInt,
// 				Amount:    &hundredInt,
// 				To:        validAddr,
// 				ChainID:   chainID,
// 			},
// 			chainID,
// 			big.NewInt(0),
// 			big.NewInt(1),
// 			big.NewInt(2),
// 			true,
// 		},
// 	}

// for _, tc := range testCases {
// 	err := tc.tx.SetSignatureValues(tc.chainID, tc.v, tc.r, tc.s)

// 	if tc.expPass {
// 		require.NoError(t, err, tc.name)
// 		continue
// 	}

// 	require.Error(t, err, tc.name)
// }
// }

func TestDynamicFeeTxValidate(t *testing.T) {
	hundredInt := sdk.NewInt(100)
	zeroInt := sdk.ZeroInt()
	minusOneInt := sdk.NewInt(-1)
	invalidAddr := "123456"
	validAddr := tests.GenerateAddress().Hex()

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
		{
			"amount is negative",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &minusOneInt,
			},
			true,
		},
		{
			"to address is invalid",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &hundredInt,
				To:        invalidAddr,
			},
			true,
		},
		{
			"chain ID not present on AccessList txs",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &hundredInt,
				To:        validAddr,
				ChainID:   nil,
			},
			true,
		},
		{
			"no errors",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &hundredInt,
				To:        validAddr,
				ChainID:   &hundredInt,
			},
			false,
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

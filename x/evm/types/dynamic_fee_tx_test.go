package types_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/types"
)

var hundredInt sdk.Int = sdk.NewInt(100)
var hundredUInt64 uint64 = hundredInt.Uint64()
var zeroInt sdk.Int = sdk.ZeroInt()
var minusOneInt sdk.Int = sdk.NewInt(-1)
var invalidAddr string = "123456"
var addr common.Address = tests.GenerateAddress()
var hexAddr string = addr.Hex()

func TestGetValue(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty amount",
			types.DynamicFeeTx{
				Amount: nil,
			},
			nil,
		},
		{
			"non-empty amount",
			types.DynamicFeeTx{
				Amount: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetValue()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetNonce(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  uint64
	}{
		{
			"non-empty nonce",
			types.DynamicFeeTx{
				Nonce: hundredUInt64,
			},
			hundredUInt64,
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetNonce()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetTo(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *common.Address
	}{
		{
			"empty address",
			types.DynamicFeeTx{
				To: "",
			},
			nil,
		},
		{
			"non-empty address",
			types.DynamicFeeTx{
				To: hexAddr,
			},
			&addr,
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetTo()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestSetSignatureValues(t *testing.T) {
	testCases := []struct {
		name    string
		chainID *big.Int
		r       *big.Int
		v       *big.Int
		s       *big.Int
	}{
		{
			"non-empty values",
			big.NewInt(9000),
			big.NewInt(1),
			big.NewInt(1),
			big.NewInt(1),
		},
		{
			"empty values",
			nil,
			nil,
			nil,
			nil,
		},
	}

	for _, tc := range testCases {
		tx := &types.DynamicFeeTx{}
		tx.SetSignatureValues(tc.chainID, tc.v, tc.r, tc.s)

		v, r, s := tx.GetRawSignatureValues()
		chainID := tx.GetChainID()

		require.Equal(t, tc.v, v, tc.name)
		require.Equal(t, tc.r, r, tc.name)
		require.Equal(t, tc.s, s, tc.name)
		require.Equal(t, tc.chainID, chainID, tc.name)
	}
}

func TestDynamicFeeTxValidate(t *testing.T) {
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
				To:        hexAddr,
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
				To:        hexAddr,
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

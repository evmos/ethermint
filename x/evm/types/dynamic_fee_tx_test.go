package types_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
	"github.com/tharsis/ethermint/x/evm/types"
)

var hundredInt sdk.Int = sdk.NewInt(100)
var hundredUInt64 uint64 = hundredInt.Uint64()
var hundredbigInt *big.Int = big.NewInt(1)
var zeroInt sdk.Int = sdk.ZeroInt()
var minusOneInt sdk.Int = sdk.NewInt(-1)
var invalidAddr string = "123456"
var addr common.Address = tests.GenerateAddress()
var hexAddr string = addr.Hex()

// TODO: Change big.Int to sdk.Int
func TestCopy(t *testing.T) {
	testCases := []struct {
		name      string
		chainID   sdk.Int
		nonce     sdk.Int
		gasTipCap sdk.Int
		gasFeeCap sdk.Int
		gasLimit  uint64
		to        *common.Address
		amount    sdk.Int
		data      []byte
		accesses  ethtypes.AccessList
		v         sdk.Int
		r         sdk.Int
		s         sdk.Int
	}{
		{
			"empty values",
			nil,
			nil,
			nil,
			nil,
			0,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{
			"non-empty values",
			hundredInt,
			hundredInt,
			hundredInt,
			hundredInt,
			hundredUInt64,
			&addr,
			hundredInt,
			nil,
			ethtypes.AccessList{
				{
					Address:     addr,
					StorageKeys: []common.Hash{},
				},
			},
			hundredInt,
			hundredInt,
			hundredInt,
		},
	}

	for _, tc := range testCases {
		tx := &types.DynamicFeeTx{
			ChainID:   &tc.chainID,
			Nonce:     &tc.nonce,
			GasTipCap: &tc.gasTipCap,
			GasFeeCap: &tc.gasFeeCap,
			GasLimit:  tc.gasLimit,
			To:        tc.to,
			Amount:    tc.amount,
			Data:      common.CopyBytes(tc.data),
			Accesses:  tc.Accesses,
			V:         common.CopyBytes(tc.v),
			R:         common.CopyBytes(tc.r),
			S:         common.CopyBytes(tc.s),
		}
		copy := tx.Copy()
		require.Equal(t, tx, copy, tc.name)
	}
}

func TestGetChainID(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty chainID",
			types.DynamicFeeTx{
				ChainID: nil,
			},
			nil,
		},
		{
			"non-empty chainID",
			types.DynamicFeeTx{
				ChainID: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetChainID()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetAccessList(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  ethtypes.AccessList
	}{
		{
			"empty accesses",
			types.DynamicFeeTx{
				Accesses: nil,
			},
			nil,
		},
		{
			"nil",
			types.DynamicFeeTx{
				Accesses: types.NewAccessList(nil),
			},
			nil,
		},
		{
			"non-empty accesses",
			types.DynamicFeeTx{
				Accesses: types.AccessList{
					{
						Address:     hexAddr,
						StorageKeys: []string{},
					},
				},
			},
			ethtypes.AccessList{
				{
					Address:     addr,
					StorageKeys: []common.Hash{},
				},
			},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetAccessList()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetData(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
	}{
		{
			"non-empty transaction",
			types.DynamicFeeTx{
				Data: nil,
			},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetData()

		require.Equal(t, tc.tx.Data, actual, tc.name)
	}
}

func TestGetGas(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  uint64
	}{
		{
			"non-empty gas",
			types.DynamicFeeTx{
				GasLimit: hundredUInt64,
			},
			hundredUInt64,
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGas()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetGasPrice(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *big.Int
	}{
		{
			"non-empty gasFeeCap",
			types.DynamicFeeTx{
				GasFeeCap: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasPrice()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetGasTipCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty gasTipCap",
			types.DynamicFeeTx{
				GasTipCap: nil,
			},
			nil,
		},
		{
			"non-empty gasTipCap",
			types.DynamicFeeTx{
				GasTipCap: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasTipCap()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestGetGasFeeCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   types.DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty gasFeeCap",
			types.DynamicFeeTx{
				GasFeeCap: nil,
			},
			nil,
		},
		{
			"non-empty gasFeeCap",
			types.DynamicFeeTx{
				GasFeeCap: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasFeeCap()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

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
			"empty values",
			nil,
			nil,
			nil,
			nil,
		},
		{
			"non-empty values",
			hundredbigInt,
			hundredbigInt,
			hundredbigInt,
			hundredbigInt,
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

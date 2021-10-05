package types

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"github.com/tharsis/ethermint/tests"
)

var (
	hundredInt    sdk.Int        = sdk.NewInt(100)
	hundredUInt64 uint64         = hundredInt.Uint64()
	hundredbigInt *big.Int       = big.NewInt(1)
	zeroInt       sdk.Int        = sdk.ZeroInt()
	minusOneInt   sdk.Int        = sdk.NewInt(-1)
	invalidAddr   string         = "123456"
	addr          common.Address = tests.GenerateAddress()
	hexAddr       string         = addr.Hex()
)

func TestNewDynamicFeeTx(t *testing.T) {
	testCases := []struct {
		name string
		tx   *ethtypes.Transaction
	}{
		{
			"non-empty tx",
			ethtypes.NewTx(&ethtypes.AccessListTx{ // TODO: change to DynamicFeeTx on Geth
				Nonce:      1,
				Data:       []byte("data"),
				Gas:        100,
				Value:      big.NewInt(1),
				AccessList: ethtypes.AccessList{},
				To:         &addr,
				V:          big.NewInt(1),
				R:          big.NewInt(27),
				S:          big.NewInt(10),
			}),
		},
	}
	for _, tc := range testCases {
		tx := newDynamicFeeTx(tc.tx)

		require.NotEmpty(t, tx)
		require.Equal(t, uint8(2), tx.TxType())
	}
}

func TestDynamicFeeTxCopy(t *testing.T) {
	tx := &DynamicFeeTx{}
	txCopy := tx.Copy()

	require.Equal(t, &DynamicFeeTx{}, txCopy)
	// TODO: Test for different pointers
}

func TestDynamicFeeTxGetChainID(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty chainID",
			DynamicFeeTx{
				ChainID: nil,
			},
			nil,
		},
		{
			"non-empty chainID",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetAccessList(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  ethtypes.AccessList
	}{
		{
			"empty accesses",
			DynamicFeeTx{
				Accesses: nil,
			},
			nil,
		},
		{
			"nil",
			DynamicFeeTx{
				Accesses: NewAccessList(nil),
			},
			nil,
		},
		{
			"non-empty accesses",
			DynamicFeeTx{
				Accesses: AccessList{
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

func TestDynamicFeeTxGetData(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
	}{
		{
			"non-empty transaction",
			DynamicFeeTx{
				Data: nil,
			},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetData()

		require.Equal(t, tc.tx.Data, actual, tc.name)
	}
}

func TestDynamicFeeTxGetGas(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  uint64
	}{
		{
			"non-empty gas",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetGasPrice(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *big.Int
	}{
		{
			"non-empty gasFeeCap",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetGasTipCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty gasTipCap",
			DynamicFeeTx{
				GasTipCap: nil,
			},
			nil,
		},
		{
			"non-empty gasTipCap",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetGasFeeCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty gasFeeCap",
			DynamicFeeTx{
				GasFeeCap: nil,
			},
			nil,
		},
		{
			"non-empty gasFeeCap",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetValue(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *big.Int
	}{
		{
			"empty amount",
			DynamicFeeTx{
				Amount: nil,
			},
			nil,
		},
		{
			"non-empty amount",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetNonce(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  uint64
	}{
		{
			"non-empty nonce",
			DynamicFeeTx{
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

func TestDynamicFeeTxGetTo(t *testing.T) {
	testCases := []struct {
		name string
		tx   DynamicFeeTx
		exp  *common.Address
	}{
		{
			"empty address",
			DynamicFeeTx{
				To: "",
			},
			nil,
		},
		{
			"non-empty address",
			DynamicFeeTx{
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

func TestDynamicFeeTxSetSignatureValues(t *testing.T) {
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
		tx := &DynamicFeeTx{}
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
		tx       DynamicFeeTx
		expError bool
	}{
		{
			"empty",
			DynamicFeeTx{},
			true,
		},
		{
			"gas tip cap is nil",
			DynamicFeeTx{
				GasTipCap: nil,
			},
			true,
		},
		{
			"gas fee cap is nil",
			DynamicFeeTx{
				GasTipCap: &zeroInt,
			},
			true,
		},
		{
			"gas tip cap is negative",
			DynamicFeeTx{
				GasTipCap: &minusOneInt,
				GasFeeCap: &zeroInt,
			},
			true,
		},
		{
			"gas tip cap is negative",
			DynamicFeeTx{
				GasTipCap: &zeroInt,
				GasFeeCap: &minusOneInt,
			},
			true,
		},
		{
			"gas fee cap < gas tip cap",
			DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &zeroInt,
			},
			true,
		},
		{
			"amount is negative",
			DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &minusOneInt,
			},
			true,
		},
		{
			"to address is invalid",
			DynamicFeeTx{
				GasTipCap: &hundredInt,
				GasFeeCap: &hundredInt,
				Amount:    &hundredInt,
				To:        invalidAddr,
			},
			true,
		},
		{
			"chain ID not present on AccessList txs",
			DynamicFeeTx{
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
			DynamicFeeTx{
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

func TestDynamicFeeTxFeeCost(t *testing.T) {
	tx := &DynamicFeeTx{}
	require.Panics(t, func() { tx.Fee() }, "should panic")
	require.Panics(t, func() { tx.Cost() }, "should panic")
}

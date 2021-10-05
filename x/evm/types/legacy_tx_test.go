package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestNewLegacyTx(t *testing.T) {
	testCases := []struct {
		name string
		tx   *ethtypes.Transaction
	}{
		{
			"non-empty Transaction",
			ethtypes.NewTx(&ethtypes.AccessListTx{
				Nonce:      1,
				Data:       []byte("data"),
				Gas:        100,
				Value:      big.NewInt(1),
				AccessList: ethtypes.AccessList{},
				To:         &addr,
				V:          big.NewInt(1),
				R:          big.NewInt(1),
				S:          big.NewInt(1),
			}),
		},
	}

	for _, tc := range testCases {
		tx := newLegacyTx(tc.tx)

		require.NotEmpty(t, tc.tx)
		require.Equal(t, uint8(0), tx.TxType())
	}
}

func TestLegacyTxTxType(t *testing.T) {
	tx := LegacyTx{}
	actual := tx.TxType()

	require.Equal(t, uint8(0), actual)
}

func TestLegacyTxCopy(t *testing.T) {
	tx := &LegacyTx{}
	txData := tx.Copy()

	require.Equal(t, &LegacyTx{}, txData)
	// TODO: Test for different pointers
}

func TestLegacyTxGetChainID(t *testing.T) {
	tx := LegacyTx{}
	actual := tx.GetChainID()

	require.Nil(t, actual)
}

func TestLegacyTxGetAccessList(t *testing.T) {
	tx := LegacyTx{}
	actual := tx.GetAccessList()

	require.Nil(t, actual)
}

func TestLegacyTxGetData(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
	}{
		{
			"non-empty transaction",
			LegacyTx{
				Data: nil,
			},
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetData()

		require.Equal(t, tc.tx.Data, actual, tc.name)
	}
}

func TestLegacyTxGetGas(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  uint64
	}{
		{
			"non-empty gas",
			LegacyTx{
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

func TestLegacyTxGetGasPrice(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  *big.Int
	}{
		{
			"empty gasPrice",
			LegacyTx{
				GasPrice: nil,
			},
			nil,
		},
		{
			"non-empty gasPrice",
			LegacyTx{
				GasPrice: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasFeeCap()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestLegacyTxGetGasTipCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			LegacyTx{
				GasPrice: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasTipCap()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestLegacyTxGetGasFeeCap(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  *big.Int
	}{
		{
			"non-empty gasPrice",
			LegacyTx{
				GasPrice: &hundredInt,
			},
			(&hundredInt).BigInt(),
		},
	}

	for _, tc := range testCases {
		actual := tc.tx.GetGasFeeCap()

		require.Equal(t, tc.exp, actual, tc.name)
	}
}

func TestLegacyTxGetValue(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  *big.Int
	}{
		{
			"empty amount",
			LegacyTx{
				Amount: nil,
			},
			nil,
		},
		{
			"non-empty amount",
			LegacyTx{
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

func TestLegacyTxGetNonce(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  uint64
	}{
		{
			"none-empty nonce",
			LegacyTx{
				Nonce: hundredUInt64,
			},
			hundredUInt64,
		},
	}
	for _, tc := range testCases {
		actual := tc.tx.GetNonce()

		require.Equal(t, tc.exp, actual)
	}
}

func TestLegacyTxGetTo(t *testing.T) {
	testCases := []struct {
		name string
		tx   LegacyTx
		exp  *common.Address
	}{
		{
			"empty address",
			LegacyTx{
				To: "",
			},
			nil,
		},
		{
			"non-empty address",
			LegacyTx{
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

func TestLegacyTxAsEthereumData(t *testing.T) {
	tx := &LegacyTx{}
	txData := tx.AsEthereumData()

	require.Equal(t, &ethtypes.LegacyTx{}, txData)
}

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

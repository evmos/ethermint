package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/tharsis/ethermint/tests"
)

func TestTxData_chainID(t *testing.T) {
	chainID := sdk.NewInt(1)

	testCases := []struct {
		msg        string
		data       TxData
		expChainID *big.Int
	}{
		{
			"access list tx", &AccessListTx{Accesses: AccessList{}, ChainID: &chainID}, big.NewInt(1),
		},
		{
			"access list tx, nil chain ID", &AccessListTx{Accesses: AccessList{}}, nil,
		},
		{
			"legacy tx, derived", &LegacyTx{}, nil,
		},
	}

	for _, tc := range testCases {
		chainID := tc.data.GetChainID()
		require.Equal(t, chainID, tc.expChainID, tc.msg)
	}
}

func TestTxData_DeriveChainID(t *testing.T) {
	testCases := []struct {
		msg        string
		data       TxData
		expChainID *big.Int
		from       common.Address
	}{
		{
			"v = 0", &AccessListTx{V: big.NewInt(0).Bytes()}, nil, tests.GenerateAddress(),
		},
		{
			"v = 1", &AccessListTx{V: big.NewInt(1).Bytes()}, big.NewInt(9223372036854775791), tests.GenerateAddress(),
		},
		{
			"v = 27", &AccessListTx{V: big.NewInt(27).Bytes()}, new(big.Int), tests.GenerateAddress(),
		},
		{
			"v = 28", &AccessListTx{V: big.NewInt(28).Bytes()}, new(big.Int), tests.GenerateAddress(),
		},
		{
			"v = nil ", &AccessListTx{V: nil}, nil, tests.GenerateAddress(),
		},
	}

	for _, tc := range testCases {
		v, _, _ := tc.data.GetRawSignatureValues()

		chainID := DeriveChainID(v)
		require.Equal(t, tc.expChainID, chainID, tc.msg)
	}
}

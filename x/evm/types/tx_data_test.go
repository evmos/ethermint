package types

import (
	"math/big"
	"testing"

	"github.com/cosmos/ethermint/tests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestTxData_chainID(t *testing.T) {
	testCases := []struct {
		msg        string
		data       TxData
		expChainID *big.Int
	}{
		{
			"access list tx", TxData{Accesses: AccessList{}, ChainID: big.NewInt(1).Bytes()}, big.NewInt(1),
		},
		{
			"access list tx, nil chain ID", TxData{Accesses: AccessList{}}, nil,
		},
		{
			"legacy tx, derived", TxData{}, nil,
		},
	}

	for _, tc := range testCases {
		chainID := tc.data.chainID()
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
			"v = 0", TxData{V: big.NewInt(0).Bytes()}, nil, tests.GenerateAddress(),
		},
		{
			"v = 1", TxData{V: big.NewInt(1).Bytes()}, big.NewInt(9223372036854775791), tests.GenerateAddress(),
		},
		{
			"v = 27", TxData{V: big.NewInt(27).Bytes()}, new(big.Int), tests.GenerateAddress(),
		},
		{
			"v = 28", TxData{V: big.NewInt(28).Bytes()}, new(big.Int), tests.GenerateAddress(),
		},
		{
			"v = nil ", TxData{V: nil}, nil, tests.GenerateAddress(),
		},
	}

	for _, tc := range testCases {
		v, _, _ := tc.data.rawSignatureValues()

		chainID := DeriveChainID(v)
		require.Equal(t, tc.expChainID, chainID, tc.msg)
	}
}

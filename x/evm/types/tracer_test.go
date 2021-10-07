package types

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestFormatLogs(t *testing.T) {
	zeroUint256 := []uint256.Int{*uint256.NewInt(0)}
	zeroByte := []byte{5}
	zeroStorage := make(map[string]string)

	testCases := []struct {
		name string
		logs []vm.StructLog
		exp  []StructLogRes
	}{
		{
			"empty logs",
			[]vm.StructLog{},
			[]StructLogRes{},
		},
		{
			"non-empty stack",
			[]vm.StructLog{
				{
					Stack: zeroUint256,
				},
			},
			[]StructLogRes{
				{
					Pc:    uint64(0),
					Op:    "STOP",
					Stack: &[]string{fmt.Sprintf("%x", zeroUint256[0])},
				},
			},
		},
		{
			"non-empty memory",
			[]vm.StructLog{
				{
					Memory: zeroByte,
				},
			},
			[]StructLogRes{
				{
					Pc:     uint64(0),
					Op:     "STOP",
					Memory: &[]string{},
				},
			},
		},
		{
			"non-empty storage",
			[]vm.StructLog{
				{
					Storage: make(map[common.Hash]common.Hash),
				},
			},
			[]StructLogRes{
				{
					Pc:      uint64(0),
					Op:      "STOP",
					Storage: &zeroStorage,
				},
			},
		},
	}
	for _, tc := range testCases {
		actual := FormatLogs(tc.logs)

		require.Equal(t, tc.exp, actual)
	}
}

func TestNewNoOpTracer(t *testing.T) {
	require.Equal(t, &NoOpTracer{}, NewNoOpTracer())
}

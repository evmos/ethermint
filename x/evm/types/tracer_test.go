package types

import (
	"math/big"
	"testing"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/require"
)

func TestNewTracer(t *testing.T) {
	testCases := []struct {
		tracer    string
		expTracer vm.Tracer
	}{
		{
			"access_list",
			&vm.AccessListTracer{},
		},
	}
	for _, tc := range testCases {
		chainID := big.NewInt(9000)
		msgEth := NewTxContract(
			chainID,
			hundredUInt64,
			hundredbigInt,
			hundredUInt64,
			hundredbigInt,
			hundredbigInt,
			hundredbigInt,
			[]byte("test"),
			&ethtypes.AccessList{},
		)
		cfg := DefaultChainConfig().EthereumConfig(chainID)
		msg, _ := msgEth.AsMessage(ethtypes.LatestSignerForChainID(chainID), big.NewInt(1))

		actual := NewTracer(tc.tracer, msg, cfg, 100, false)

		require.Equal(t, tc.expTracer, actual, tc.tracer)
	}
}

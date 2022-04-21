package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

func TestParseTxResult(t *testing.T) {
	testCases := []struct {
		name     string
		response abci.ResponseDeliverTx
		expTxs   map[common.Hash]*ParsedTx
	}{
		{"format 1 events",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: "coin_received", Attributes: []abci.EventAttribute{
						{Key: []byte("receiver"), Value: []byte("ethm12luku6uxehhak02py4rcz65zu0swh7wjun6msa")},
						{Key: []byte("amount"), Value: []byte("1252860basetcro")},
					}},
					{Type: "coin_spent", Attributes: []abci.EventAttribute{
						{Key: []byte("spender"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("amount"), Value: []byte("1252860basetcro")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2")},
						{Key: []byte("txIndex"), Value: []byte("0")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
					{Type: "message", Attributes: []abci.EventAttribute{
						{Key: []byte("action"), Value: []byte("/ethermint.evm.v1.MsgEthereumTx")},
						{Key: []byte("key"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("module"), Value: []byte("evm")},
						{Key: []byte("sender"), Value: []byte("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2")},
					}},
				},
			},
			map[common.Hash]*ParsedTx{
				common.HexToHash("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"): &ParsedTx{
					MsgIndex:   0,
					Hash:       common.HexToHash("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"),
					EthTxIndex: 0,
					GasUsed:    21000,
					Failed:     false,
					RawLogs:    nil,
				},
			},
		},
		{"format 2 events",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: "coin_received", Attributes: []abci.EventAttribute{
						{Key: []byte("receiver"), Value: []byte("ethm12luku6uxehhak02py4rcz65zu0swh7wjun6msa")},
						{Key: []byte("amount"), Value: []byte("1252860basetcro")},
					}},
					{Type: "coin_spent", Attributes: []abci.EventAttribute{
						{Key: []byte("spender"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("amount"), Value: []byte("1252860basetcro")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2")},
						{Key: []byte("txIndex"), Value: []byte("0")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
					{Type: "message", Attributes: []abci.EventAttribute{
						{Key: []byte("action"), Value: []byte("/ethermint.evm.v1.MsgEthereumTx")},
						{Key: []byte("key"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("module"), Value: []byte("evm")},
						{Key: []byte("sender"), Value: []byte("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2")},
					}},
				},
			},
			map[common.Hash]*ParsedTx{
				common.HexToHash("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"): &ParsedTx{
					MsgIndex:   0,
					Hash:       common.HexToHash("0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"),
					EthTxIndex: 0,
					GasUsed:    21000,
					Failed:     false,
					RawLogs:    nil,
				},
			},
		},
		// TODO negative cases
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTxResult(&tc.response)
			require.NoError(t, err)
			for hash, expTx := range tc.expTxs {
				require.Equal(t, expTx, parsed.GetTxByHash(hash))
			}
		})
	}
}

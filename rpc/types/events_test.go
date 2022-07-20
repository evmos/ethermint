package types

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestParseTxResult(t *testing.T) {
	rawLogs := [][]byte{
		[]byte("{\"address\":\"0xdcC261c03cD2f33eBea404318Cdc1D9f8b78e1AD\",\"topics\":[\"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef\",\"0x000000000000000000000000569608516a81c0b1247310a3e0cd001046da0663\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANB/GezJGOI=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":5}"),
		[]byte("{\"address\":\"0x569608516A81C0B1247310A3E0CD001046dA0663\",\"topics\":[\"0xe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e0486\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANB/GezJGOI=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":6}"),
		[]byte("{\"address\":\"0x569608516A81C0B1247310A3E0CD001046dA0663\",\"topics\":[\"0xf279e6a1f5e320cca91135676d9cb6e44ca8a08c0b88342bcdb1144f6511b568\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\",\"0x0000000000000000000000000000000000000000000000000000000000000001\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":7}"),
	}
	address := "0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"
	txHash := common.BigToHash(big.NewInt(1))
	txHash2 := common.BigToHash(big.NewInt(2))

	testCases := []struct {
		name     string
		response abci.ResponseDeliverTx
		expTxs   []*ParsedTx // expected parse result, nil means expect error.
	}{
		{
			"format 1 events",
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
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("10")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[0]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[1]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[2]},
					}},
					{Type: "message", Attributes: []abci.EventAttribute{
						{Key: []byte("action"), Value: []byte("/ethermint.evm.v1.MsgEthereumTx")},
						{Key: []byte("key"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("module"), Value: []byte("evm")},
						{Key: []byte("sender"), Value: []byte(address)},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash2.Hex())},
						{Key: []byte("txIndex"), Value: []byte("11")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						{Key: []byte("ethereumTxFailed"), Value: []byte("contract reverted")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
				},
			},
			[]*ParsedTx{
				{
					MsgIndex:   0,
					Hash:       txHash,
					EthTxIndex: 10,
					GasUsed:    21000,
					Failed:     false,
					RawLogs:    rawLogs,
				},
				{
					MsgIndex:   1,
					Hash:       txHash2,
					EthTxIndex: 11,
					GasUsed:    21000,
					Failed:     true,
					RawLogs:    nil,
				},
			},
		},
		{
			"format 2 events",
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
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("0")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("0")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[0]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[1]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[2]},
					}},
					{Type: "message", Attributes: []abci.EventAttribute{
						{Key: []byte("action"), Value: []byte("/ethermint.evm.v1.MsgEthereumTx")},
						{Key: []byte("key"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
						{Key: []byte("module"), Value: []byte("evm")},
						{Key: []byte("sender"), Value: []byte(address)},
					}},
				},
			},
			[]*ParsedTx{
				{
					MsgIndex:   0,
					Hash:       txHash,
					EthTxIndex: 0,
					GasUsed:    21000,
					Failed:     false,
					RawLogs:    rawLogs,
				},
			},
		},
		{
			"format 1 events, failed",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("10")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[0]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[1]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[2]},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash2.Hex())},
						{Key: []byte("txIndex"), Value: []byte("0x01")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						{Key: []byte("ethereumTxFailed"), Value: []byte("contract reverted")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
				},
			},
			nil,
		},
		{
			"format 1 events, failed",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("10")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[0]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[1]},
						{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[2]},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash2.Hex())},
						{Key: []byte("txIndex"), Value: []byte("10")},
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("0x01")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
						{Key: []byte("ethereumTxFailed"), Value: []byte("contract reverted")},
					}},
					{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
				},
			},
			nil,
		},
		{
			"format 2 events failed",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("0x01")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("21000")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
				},
			},
			nil,
		},
		{
			"format 2 events failed",
			abci.ResponseDeliverTx{
				GasUsed: 21000,
				Events: []abci.Event{
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
						{Key: []byte("txIndex"), Value: []byte("10")},
					}},
					{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
						{Key: []byte("amount"), Value: []byte("1000")},
						{Key: []byte("txGasUsed"), Value: []byte("0x01")},
						{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
						{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
					}},
				},
			},
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := ParseTxResult(&tc.response)
			if tc.expTxs == nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				for msgIndex, expTx := range tc.expTxs {
					require.Equal(t, expTx, parsed.GetTxByMsgIndex(msgIndex))
					require.Equal(t, expTx, parsed.GetTxByHash(expTx.Hash))
					require.Equal(t, expTx, parsed.GetTxByTxIndex(int(expTx.EthTxIndex)))
					_, err := expTx.ParseTxLogs()
					require.NoError(t, err)
				}
				// non-exists tx hash
				require.Nil(t, parsed.GetTxByHash(common.Hash{}))
				// out of range
				require.Nil(t, parsed.GetTxByMsgIndex(len(tc.expTxs)))
				require.Nil(t, parsed.GetTxByTxIndex(99999999))
			}
		})
	}
}

func TestParseTxLogs(t *testing.T) {
	rawLogs := [][]byte{
		[]byte("{\"address\":\"0xdcC261c03cD2f33eBea404318Cdc1D9f8b78e1AD\",\"topics\":[\"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef\",\"0x000000000000000000000000569608516a81c0b1247310a3e0cd001046da0663\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANB/GezJGOI=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":5}"),
		[]byte("{\"address\":\"0x569608516A81C0B1247310A3E0CD001046dA0663\",\"topics\":[\"0xe2403640ba68fed3a2f88b7557551d1993f84b99bb10ff833f0cf8db0c5e0486\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANB/GezJGOI=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":6}"),
		[]byte("{\"address\":\"0x569608516A81C0B1247310A3E0CD001046dA0663\",\"topics\":[\"0xf279e6a1f5e320cca91135676d9cb6e44ca8a08c0b88342bcdb1144f6511b568\",\"0x0000000000000000000000002eea2c1ae0cdd2622381c2f9201b2a07c037b1f6\",\"0x0000000000000000000000000000000000000000000000000000000000000001\"],\"data\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"blockNumber\":1803258,\"transactionHash\":\"0xcf4354b55b9ac77436cf8b2f5c229ad3b3119b5196cd79ac5c6c382d9f7b0a71\",\"transactionIndex\":1,\"blockHash\":\"0xa69a510b0848180a094904ea9ae3f0ca2216029470c8e03e6941b402aba610d8\",\"logIndex\":7}"),
	}
	address := "0x57f96e6B86CdeFdB3d412547816a82E3E0EbF9D2"
	txHash := common.BigToHash(big.NewInt(1))
	txHash2 := common.BigToHash(big.NewInt(2))
	response := abci.ResponseDeliverTx{
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
				{Key: []byte("ethereumTxHash"), Value: []byte(txHash.Hex())},
				{Key: []byte("txIndex"), Value: []byte("10")},
				{Key: []byte("amount"), Value: []byte("1000")},
				{Key: []byte("txGasUsed"), Value: []byte("21000")},
				{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
				{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
			}},
			{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{
				{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[0]},
				{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[1]},
				{Key: []byte(evmtypes.AttributeKeyTxLog), Value: rawLogs[2]},
			}},
			{Type: "message", Attributes: []abci.EventAttribute{
				{Key: []byte("action"), Value: []byte("/ethermint.evm.v1.MsgEthereumTx")},
				{Key: []byte("key"), Value: []byte("ethm17xpfvakm2amg962yls6f84z3kell8c5lthdzgl")},
				{Key: []byte("module"), Value: []byte("evm")},
				{Key: []byte("sender"), Value: []byte(address)},
			}},
			{Type: evmtypes.EventTypeEthereumTx, Attributes: []abci.EventAttribute{
				{Key: []byte("ethereumTxHash"), Value: []byte(txHash2.Hex())},
				{Key: []byte("txIndex"), Value: []byte("11")},
				{Key: []byte("amount"), Value: []byte("1000")},
				{Key: []byte("txGasUsed"), Value: []byte("21000")},
				{Key: []byte("txHash"), Value: []byte("14A84ED06282645EFBF080E0B7ED80D8D8D6A36337668A12B5F229F81CDD3F57")},
				{Key: []byte("recipient"), Value: []byte("0x775b87ef5D82ca211811C1a02CE0fE0CA3a455d7")},
				{Key: []byte("ethereumTxFailed"), Value: []byte("contract reverted")},
			}},
			{Type: evmtypes.EventTypeTxLog, Attributes: []abci.EventAttribute{}},
		},
	}
	parsed, err := ParseTxResult(&response)
	require.NoError(t, err)
	tx1 := parsed.GetTxByMsgIndex(0)
	txLogs1, err := tx1.ParseTxLogs()
	require.NoError(t, err)
	require.NotEmpty(t, txLogs1)

	tx2 := parsed.GetTxByMsgIndex(1)
	txLogs2, err := tx2.ParseTxLogs()
	require.Empty(t, txLogs2)
}

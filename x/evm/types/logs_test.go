package types

import (
	"testing"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTransactionLogsValidate(t *testing.T) {
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	testCases := []struct {
		name    string
		txLogs  TransactionLogs
		expPass bool
	}{
		{
			"valid log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")),
				Logs: []*ethtypes.Log{
					{
						Address:     addr,
						Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("tx_hash")),
						TxIndex:     1,
						BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
						Index:       1,
						Removed:     false,
					},
				},
			},
			true,
		},
		{
			"empty hash",
			TransactionLogs{
				Hash: ethcmn.Hash{},
			},
			false,
		},
		{
			"invalid log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")),
				Logs: []*ethtypes.Log{nil},
			},
			false,
		},
		{
			"hash mismatch log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")),
				Logs: []*ethtypes.Log{
					{
						Address:     addr,
						Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("other_hash")),
						TxIndex:     1,
						BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
						Index:       1,
						Removed:     false,
					},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.txLogs.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestValidateLog(t *testing.T) {
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)

	testCases := []struct {
		name    string
		log     *ethtypes.Log
		expPass bool
	}{
		{
			"valid log",
			&ethtypes.Log{
				Address:     addr,
				Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      ethcmn.BytesToHash([]byte("tx_hash")),
				TxIndex:     1,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
				Index:       1,
				Removed:     false,
			},
			true,
		},
		{
			"nil log", nil, false,
		},
		{
			"zero address",
			&ethtypes.Log{
				Address: ethcmn.Address{},
			},
			false,
		},
		{
			"empty block hash",
			&ethtypes.Log{
				Address:   addr,
				BlockHash: ethcmn.Hash{},
			},
			false,
		},
		{
			"zero block number",
			&ethtypes.Log{
				Address:     addr,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
				BlockNumber: 0,
			},
			false,
		},
		{
			"empty tx hash",
			&ethtypes.Log{
				Address:     addr,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")),
				BlockNumber: 1,
				TxHash:      ethcmn.Hash{},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := ValidateLog(tc.log)
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func TestTransactionLogsValidate(t *testing.T) {
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(t, err)
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey).String()

	testCases := []struct {
		name    string
		txLogs  TransactionLogs
		expPass bool
	}{
		{
			"valid log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*Log{
					{
						Address:     addr,
						Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("tx_hash")).String(),
						TxIndex:     1,
						BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
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
				Hash: ethcmn.Hash{}.String(),
			},
			false,
		},
		{
			"invalid log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*Log{{}},
			},
			false,
		},
		{
			"hash mismatch log",
			TransactionLogs{
				Hash: ethcmn.BytesToHash([]byte("tx_hash")).String(),
				Logs: []*Log{
					{
						Address:     addr,
						Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("other_hash")).String(),
						TxIndex:     1,
						BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
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
	addr := ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey).String()

	testCases := []struct {
		name    string
		log     *Log
		expPass bool
	}{
		{
			"valid log",
			&Log{
				Address:     addr,
				Topics:      []string{ethcmn.BytesToHash([]byte("topic")).String()},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      ethcmn.BytesToHash([]byte("tx_hash")).String(),
				TxIndex:     1,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
				Index:       1,
				Removed:     false,
			},
			true,
		},
		{
			"empty log", &Log{}, false,
		},
		{
			"zero address",
			&Log{
				Address: ethcmn.Address{}.String(),
			},
			false,
		},
		{
			"empty block hash",
			&Log{
				Address:   addr,
				BlockHash: ethcmn.Hash{}.String(),
			},
			false,
		},
		{
			"zero block number",
			&Log{
				Address:     addr,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
				BlockNumber: 0,
			},
			false,
		},
		{
			"empty tx hash",
			&Log{
				Address:     addr,
				BlockHash:   ethcmn.BytesToHash([]byte("block_hash")).String(),
				BlockNumber: 1,
				TxHash:      ethcmn.Hash{}.String(),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.log.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

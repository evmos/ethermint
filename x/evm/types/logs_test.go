package types

import (
	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func (suite *GenesisTestSuite) TestTransactionLogsValidate() {
	testCases := []struct {
		name    string
		txLogs  TransactionLogs
		expPass bool
	}{
		{
			"valid log",
			TransactionLogs{
				Hash: suite.hash.String(),
				Logs: []*ethtypes.Log{
					{
						Address:     suite.address,
						Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      suite.hash,
						TxIndex:     1,
						BlockHash:   suite.hash,
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
				Hash: suite.hash.String(),
				Logs: []*ethtypes.Log{nil},
			},
			false,
		},
		{
			"hash mismatch log",
			TransactionLogs{
				Hash: suite.hash.String(),
				Logs: []*ethtypes.Log{
					{
						Address:     suite.address,
						Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("other_hash")),
						TxIndex:     1,
						BlockHash:   suite.hash,
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

func (suite *GenesisTestSuite) TestValidateLog() {
	testCases := []struct {
		name    string
		log     *ethtypes.Log
		expPass bool
	}{
		{
			"valid log",
			&ethtypes.Log{
				Address:     suite.address,
				Topics:      []ethcmn.Hash{ethcmn.BytesToHash([]byte("topic"))},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      suite.hash,
				TxIndex:     1,
				BlockHash:   suite.hash,
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
				Address:   suite.address,
				BlockHash: ethcmn.Hash{},
			},
			false,
		},
		{
			"zero block number",
			&ethtypes.Log{
				Address:     suite.address,
				BlockHash:   suite.hash,
				BlockNumber: 0,
			},
			false,
		},
		{
			"empty tx hash",
			&ethtypes.Log{
				Address:     suite.address,
				BlockHash:   suite.hash,
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

package types

import (
	ethcmn "github.com/ethereum/go-ethereum/common"
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
				Logs: []*Log{
					{
						Address:     suite.address,
						Topics:      []string{suite.hash.String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      suite.hash.String(),
						TxIndex:     1,
						BlockHash:   suite.hash.String(),
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
				Logs: []*Log{nil},
			},
			false,
		},
		{
			"hash mismatch log",
			TransactionLogs{
				Hash: suite.hash.String(),
				Logs: []*Log{
					{
						Address:     suite.address,
						Topics:      []string{suite.hash.String()},
						Data:        []byte("data"),
						BlockNumber: 1,
						TxHash:      ethcmn.BytesToHash([]byte("other_hash")).String(),
						TxIndex:     1,
						BlockHash:   suite.hash.String(),
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
		log     *Log
		expPass bool
	}{
		{
			"valid log",
			&Log{
				Address:     suite.address,
				Topics:      []string{suite.hash.String()},
				Data:        []byte("data"),
				BlockNumber: 1,
				TxHash:      suite.hash.String(),
				TxIndex:     1,
				BlockHash:   suite.hash.String(),
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
				Address:   suite.address,
				BlockHash: ethcmn.Hash{}.String(),
			},
			false,
		},
		{
			"zero block number",
			&Log{
				Address:     suite.address,
				BlockHash:   suite.hash.String(),
				BlockNumber: 0,
			},
			false,
		},
		{
			"empty tx hash",
			&Log{
				Address:     suite.address,
				BlockHash:   suite.hash.String(),
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
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

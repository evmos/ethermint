package state

import (
	"fmt"
	"testing"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/require"
)

func TestDatabaseInterface(t *testing.T) {
	require.Implements(t, (*ethstate.Database)(nil), new(Database))
}

func TestDatabaseLatestVersion(t *testing.T) {
	var version int64

	testDB := newTestDatabase()

	version = testDB.LatestVersion()
	require.Equal(t, int64(0), version)

	testDB.Commit()
	version = testDB.LatestVersion()
	require.Equal(t, int64(1), version)
}

func TestDatabaseCopyTrie(t *testing.T) {
	// TODO: Implement once CopyTrie is implemented
	t.SkipNow()
}

func TestDatabaseContractCode(t *testing.T) {
	testDB := newTestDatabase()

	testCases := []struct {
		db           *Database
		data         *code
		codeHash     ethcmn.Hash
		expectedCode []byte
	}{
		{
			db:           testDB,
			codeHash:     ethcmn.BytesToHash([]byte("code hash")),
			expectedCode: nil,
		},
		{
			db:           testDB,
			data:         &code{ethcmn.BytesToHash([]byte("code hash")), []byte("some awesome code")},
			codeHash:     ethcmn.BytesToHash([]byte("code hash")),
			expectedCode: []byte("some awesome code"),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.db.codeDB.Set(tc.data.hash[:], tc.data.blob)
		}

		code, err := tc.db.ContractCode(ethcmn.Hash{}, tc.codeHash)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedCode, code, fmt.Sprintf("unexpected result: test case #%d", i))
	}
}

func TestDatabaseContractCodeSize(t *testing.T) {
	testDB := newTestDatabase()

	testCases := []struct {
		db              *Database
		data            *code
		codeHash        ethcmn.Hash
		expectedCodeLen int
	}{
		{
			db:              testDB,
			codeHash:        ethcmn.BytesToHash([]byte("code hash")),
			expectedCodeLen: 0,
		},
		{
			db:              testDB,
			data:            &code{ethcmn.BytesToHash([]byte("code hash")), []byte("some awesome code")},
			codeHash:        ethcmn.BytesToHash([]byte("code hash")),
			expectedCodeLen: 17,
		},
		{
			db:              testDB,
			codeHash:        ethcmn.BytesToHash([]byte("code hash")),
			expectedCodeLen: 17,
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.db.codeDB.Set(tc.data.hash[:], tc.data.blob)
		}

		codeLen, err := tc.db.ContractCodeSize(ethcmn.Hash{}, tc.codeHash)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedCodeLen, codeLen, fmt.Sprintf("unexpected result: test case #%d", i))
	}
}

func TestDatabaseTrieDB(t *testing.T) {
	testDB := newTestDatabase()

	db := testDB.TrieDB()
	require.Equal(t, testDB.ethTrieDB, db)
}

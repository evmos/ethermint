package state

// NOTE: A bulk of these unit tests will change and evolve as the context and
// implementation of ChainConext evolves.

import (
	"fmt"
	"testing"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tendermint/libs/db"
)

type (
	kvPair struct {
		key, value []byte
	}
)

func newEthereumDB() *EthereumDB {
	memDB := dbm.NewMemDB()
	return &EthereumDB{CodeDB: memDB}
}

func TestEthereumDBInterface(t *testing.T) {
	require.Implements(t, (*ethdb.Database)(nil), new(EthereumDB))
	require.Implements(t, (*ethdb.Batch)(nil), new(EthereumDB))
}

func TestEthereumDBGet(t *testing.T) {
	testEDB := newEthereumDB()

	testCases := []struct {
		edb           *EthereumDB
		data          *kvPair
		key           []byte
		expectedValue []byte
	}{
		{
			edb:           testEDB,
			key:           []byte("foo"),
			expectedValue: nil,
		},
		{
			edb:           testEDB,
			data:          &kvPair{[]byte("foo"), []byte("bar")},
			key:           []byte("foo"),
			expectedValue: []byte("bar"),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.edb.Put(tc.data.key, tc.data.value)
		}

		value, err := tc.edb.Get(tc.key)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedValue, value, fmt.Sprintf("unexpected result: test case #%d", i))
	}
}

func TestEthereumDBHas(t *testing.T) {
	testEDB := newEthereumDB()

	testCases := []struct {
		edb           *EthereumDB
		data          *kvPair
		key           []byte
		expectedValue bool
	}{
		{
			edb:           testEDB,
			key:           []byte("foo"),
			expectedValue: false,
		},
		{
			edb:           testEDB,
			data:          &kvPair{[]byte("foo"), []byte("bar")},
			key:           []byte("foo"),
			expectedValue: true,
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.edb.Put(tc.data.key, tc.data.value)
		}

		ok, err := tc.edb.Has(tc.key)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedValue, ok, fmt.Sprintf("unexpected result: test case #%d", i))
	}
}

func TestEthereumDBDelete(t *testing.T) {
	testEDB := newEthereumDB()

	testCases := []struct {
		edb  *EthereumDB
		data *kvPair
		key  []byte
	}{
		{
			edb: testEDB,
			key: []byte("foo"),
		},
		{
			edb:  testEDB,
			data: &kvPair{[]byte("foo"), []byte("bar")},
			key:  []byte("foo"),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.edb.Put(tc.data.key, tc.data.value)
		}

		err := tc.edb.Delete(tc.key)
		ok, _ := tc.edb.Has(tc.key)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.False(t, ok, fmt.Sprintf("unexpected existence of key: test case #%d", i))
	}
}

func TestEthereumDBNewBatch(t *testing.T) {
	edb := newEthereumDB()

	batch := edb.NewBatch()
	require.Equal(t, edb, batch)
}

func TestEthereumDBValueSize(t *testing.T) {
	edb := newEthereumDB()

	size := edb.ValueSize()
	require.Equal(t, 0, size)
}

func TestEthereumDBWrite(t *testing.T) {
	edb := newEthereumDB()

	err := edb.Write()
	require.Nil(t, err)
}

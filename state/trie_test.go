package state

import (
	"fmt"
	"math/rand"
	"testing"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	"github.com/stretchr/testify/require"
)

func newTestTrie() *Trie {
	testDB := newTestDatabase()
	testTrie, _ := testDB.OpenTrie(rootHashFromVersion(0))

	return testTrie.(*Trie)
}

func newTestPrefixTrie() *Trie {
	testDB := newTestDatabase()

	prefix := make([]byte, ethcmn.HashLength)
	rand.Read(prefix)

	testDB.OpenTrie(rootHashFromVersion(0))
	testTrie, _ := testDB.OpenStorageTrie(ethcmn.BytesToHash(prefix), rootHashFromVersion(0))

	return testTrie.(*Trie)
}

func TestTrieInterface(t *testing.T) {
	require.Implements(t, (*ethstate.Trie)(nil), new(Trie))
}

func TestTrieTryGet(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()

	testCases := []struct {
		trie          *Trie
		data          *kvPair
		key           []byte
		expectedValue []byte
	}{
		{
			trie:          testTrie,
			data:          &kvPair{[]byte("foo"), []byte("bar")},
			key:           []byte("foo"),
			expectedValue: []byte("bar"),
		},
		{
			trie:          testTrie,
			key:           []byte("baz"),
			expectedValue: nil,
		},
		{
			trie:          testPrefixTrie,
			data:          &kvPair{[]byte("foo"), []byte("bar")},
			key:           []byte("foo"),
			expectedValue: []byte("bar"),
		},
		{
			trie:          testPrefixTrie,
			key:           []byte("baz"),
			expectedValue: nil,
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.trie.TryUpdate(tc.data.key, tc.data.value)
		}

		value, err := tc.trie.TryGet(tc.key)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedValue, value, fmt.Sprintf("unexpected value: test case #%d", i))
	}
}

func TestTrieTryUpdate(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()
	kv := &kvPair{[]byte("foo"), []byte("bar")}

	var err error

	err = testTrie.TryUpdate(kv.key, kv.value)
	require.Nil(t, err)

	err = testPrefixTrie.TryUpdate(kv.key, kv.value)
	require.Nil(t, err)
}

func TestTrieTryDelete(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()

	testCases := []struct {
		trie *Trie
		data *kvPair
		key  []byte
	}{
		{
			trie: testTrie,
			data: &kvPair{[]byte("foo"), []byte("bar")},
			key:  []byte("foo"),
		},
		{
			trie: testTrie,
			key:  []byte("baz"),
		},
		{
			trie: testPrefixTrie,
			data: &kvPair{[]byte("foo"), []byte("bar")},
			key:  []byte("foo"),
		},
		{
			trie: testPrefixTrie,
			key:  []byte("baz"),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.trie.TryUpdate(tc.data.key, tc.data.value)
		}

		err := tc.trie.TryDelete(tc.key)
		value, _ := tc.trie.TryGet(tc.key)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Nil(t, value, fmt.Sprintf("unexpected value: test case #%d", i))
	}
}

func TestTrieCommit(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()

	testCases := []struct {
		trie         *Trie
		data         *kvPair
		code         *code
		expectedRoot ethcmn.Hash
	}{
		{
			trie:         &Trie{empty: true},
			expectedRoot: ethcmn.Hash{},
		},
		{
			trie:         testTrie,
			data:         &kvPair{[]byte("foo"), []byte("bar")},
			expectedRoot: rootHashFromVersion(1),
		},
		{
			trie:         testTrie,
			data:         &kvPair{[]byte("baz"), []byte("cat")},
			code:         &code{ethcmn.BytesToHash([]byte("code hash")), []byte("code hash")},
			expectedRoot: rootHashFromVersion(2),
		},
		{
			trie:         testTrie,
			expectedRoot: rootHashFromVersion(3),
		},
		{
			trie:         testPrefixTrie,
			expectedRoot: rootHashFromVersion(0),
		},
		{
			trie:         testPrefixTrie,
			data:         &kvPair{[]byte("foo"), []byte("bar")},
			expectedRoot: rootHashFromVersion(1),
		},
		{
			trie:         testPrefixTrie,
			expectedRoot: rootHashFromVersion(2),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.trie.TryUpdate(tc.data.key, tc.data.value)
		}
		if tc.code != nil {
			tc.trie.ethTrieDB.Insert(tc.code.hash, tc.code.blob)
		}

		root, err := tc.trie.Commit(nil)
		require.Nil(t, err, fmt.Sprintf("unexpected error: test case #%d", i))
		require.Equal(t, tc.expectedRoot, root, fmt.Sprintf("unexpected root: test case #%d", i))
	}
}

func TestTrieHash(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()

	testCases := []struct {
		trie         *Trie
		data         *kvPair
		expectedRoot ethcmn.Hash
	}{
		{
			trie:         testTrie,
			expectedRoot: rootHashFromVersion(0),
		},
		{
			trie:         testTrie,
			data:         &kvPair{[]byte("foo"), []byte("bar")},
			expectedRoot: rootHashFromVersion(1),
		},
		{
			trie:         testPrefixTrie,
			expectedRoot: rootHashFromVersion(0),
		},
		{
			trie:         testPrefixTrie,
			data:         &kvPair{[]byte("foo"), []byte("bar")},
			expectedRoot: rootHashFromVersion(1),
		},
	}

	for i, tc := range testCases {
		if tc.data != nil {
			tc.trie.TryUpdate(tc.data.key, tc.data.value)
			tc.trie.Commit(nil)
		}

		root := tc.trie.Hash()
		require.Equal(t, tc.expectedRoot, root, fmt.Sprintf("unexpected root: test case #%d", i))
	}
}

func TestTrieNodeIterator(t *testing.T) {
	// TODO: Implement once NodeIterator is implemented
	t.SkipNow()
}

func TestTrieGetKey(t *testing.T) {
	testTrie := newTestTrie()
	testPrefixTrie := newTestPrefixTrie()

	var key []byte
	expectedKey := []byte("foo")

	key = testTrie.GetKey(expectedKey)
	require.Equal(t, expectedKey, key)

	key = testPrefixTrie.GetKey(expectedKey)
	require.Equal(t, expectedKey, key)
}

func TestTrieProve(t *testing.T) {
	// TODO: Implement once Prove is implemented
	t.SkipNow()
}

package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
)

func TestStorageValidate(t *testing.T) {
	testCases := []struct {
		name    string
		storage Storage
		expPass bool
	}{
		{
			"valid storage",
			Storage{
				NewState(ethcmn.BytesToHash([]byte{1, 2, 3}), ethcmn.BytesToHash([]byte{1, 2, 3})),
			},
			true,
		},
		{
			"empty storage key bytes",
			Storage{
				{Key: ethcmn.Hash{}.String()},
			},
			false,
		},
		{
			"duplicated storage key",
			Storage{
				{Key: ethcmn.BytesToHash([]byte{1, 2, 3}).String()},
				{Key: ethcmn.BytesToHash([]byte{1, 2, 3}).String()},
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		err := tc.storage.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}

func TestStorageCopy(t *testing.T) {
	testCases := []struct {
		name    string
		storage Storage
	}{
		{
			"single storage",
			Storage{
				NewState(ethcmn.BytesToHash([]byte{1, 2, 3}), ethcmn.BytesToHash([]byte{1, 2, 3})),
			},
		},
		{
			"empty storage key value bytes",
			Storage{
				{Key: ethcmn.Hash{}.String(), Value: ethcmn.Hash{}.String()},
			},
		},
		{
			"empty storage",
			Storage{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		require.Equal(t, tc.storage, tc.storage.Copy(), tc.name)
	}
}

func TestStorageString(t *testing.T) {
	storage := Storage{NewState(ethcmn.BytesToHash([]byte("key")), ethcmn.BytesToHash([]byte("value")))}
	str := `key:"0x00000000000000000000000000000000000000000000000000000000006b6579" value:"0x00000000000000000000000000000000000000000000000000000076616c7565" 
`
	require.Equal(t, str, storage.String())
}

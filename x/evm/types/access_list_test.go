package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestNewAccessList(t *testing.T) {
	testCases := []struct {
		name          string
		ethAccessList *ethtypes.AccessList
		expAl         AccessList
	}{
		{
			"ethAccessList is nil",
			nil,
			nil,
		},
		{
			"non-empty ethAccessList",
			&ethtypes.AccessList{{Address: addr, StorageKeys: []common.Hash{{0}}}},
			AccessList{{Address: addr.Hex(), StorageKeys: []string{common.Hash{}.Hex()}}},
		},
	}
	for _, tc := range testCases {
		al := NewAccessList(tc.ethAccessList)

		require.Equal(t, tc.expAl, al)
	}
}

func TestAccessListToEthAccessList(t *testing.T) {
	ethAccessList := ethtypes.AccessList{{Address: addr, StorageKeys: []common.Hash{{0}}}}
	al := NewAccessList(&ethAccessList)
	actual := al.ToEthAccessList()

	require.Equal(t, &ethAccessList, actual)
}

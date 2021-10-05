package types

import (
	"testing"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestAccessListToEthAccessList(t *testing.T) {
	al := AccessList{}
	actual := al.ToEthAccessList()

	require.Equal(t, ethtypes.AccessList{}, actual)
}

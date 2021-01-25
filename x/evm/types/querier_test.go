package types

import (
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestQueryETHLogs_String(t *testing.T) {
	const expectedQueryETHLogsStr = `{0x0000000000000000000000000000000000000000 [] [1 2 3 4] 9 0x0000000000000000000000000000000000000000000000000000000000000000 0 0x0000000000000000000000000000000000000000000000000000000000000000 0 false}
{0x0000000000000000000000000000000000000000 [] [5 6 7 8] 10 0x0000000000000000000000000000000000000000000000000000000000000000 0 0x0000000000000000000000000000000000000000000000000000000000000000 0 false}
`
	logs := []*ethtypes.Log{
		{
			Data:        []byte{1, 2, 3, 4},
			BlockNumber: 9,
		},
		{
			Data:        []byte{5, 6, 7, 8},
			BlockNumber: 10,
		},
	}

	require.True(t, strings.EqualFold(expectedQueryETHLogsStr, QueryETHLogs{logs}.String()))
}

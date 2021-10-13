package types

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsLondon(t *testing.T) {
	testCases := []struct {
		name         string
		height       int64
		result       bool
	}{
		{
			"Before london block",
			5,
			false,
		},
		{
			"After london block",
			12_965_001,
			true,
		},
		{
			"london block",
			12_965_000,
			true,
		},
	}

	for _, tc := range testCases {
		ethConfig := params.MainnetChainConfig
		require.Equal(t, IsLondon(ethConfig, tc.height), tc.result)
	}
}

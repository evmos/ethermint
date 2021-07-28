package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultEVMConfig()
	require.True(t, cfg.Enable)
	require.Equal(t, cfg.RPCAddress, DefaultEVMAddress)
	require.Equal(t, cfg.WsAddress, DefaultEVMWSAddress)
}

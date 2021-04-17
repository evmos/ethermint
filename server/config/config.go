package config

import (
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/server/config"
)

const (
	// DefaultJSONRPCAddress is the default address the JSON-RPC server binds to.
	DefaultJSONRPCAddress = "tcp://0.0.0.0:8545"
	// DefaultEthereumWebsocketAddress is the default address the Ethereum websocket server binds to.
	DefaultEthereumWebsocketAddress = "tcp://0.0.0.0:8546"
)

// Config defines the server's top level configuration
type Config struct {
	*config.Config

	JSONRPC           JSONRPCConfig   `mapstructure:"json-rpc"`
	EthereumWebsocket WebsocketConfig `mapstructure:"ethereum-websocket"`
}

// JSONRPCConfig defines the Ethereum API listener configuration.
type JSONRPCConfig struct {
	// Enable defines if the JSON-RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// Address defines the JSON-RPC server address to listen on
	Address string `mapstructure:"address"`
}

// WebsocketConfig defines the Ethereum API listener configuration.
type WebsocketConfig struct {
	// Enable defines if the Ethereum websocker server should be enabled.
	Enable bool `mapstructure:"enable"`

	// Address defines the Websocket server address to listen on
	Address string `mapstructure:"address"`
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config: config.DefaultConfig(),
		JSONRPC: JSONRPCConfig{
			Enable:  true,
			Address: DefaultJSONRPCAddress,
		},
		EthereumWebsocket: WebsocketConfig{
			Enable:  true,
			Address: DefaultEthereumWebsocketAddress,
		},
	}
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) Config {
	sdkConfig := config.GetConfig(v)
	return Config{
		Config: &sdkConfig,
		JSONRPC: JSONRPCConfig{
			Enable:  v.GetBool("json-rpc.enable"),
			Address: v.GetString("json-rpc.address"),
		},
		EthereumWebsocket: WebsocketConfig{
			Enable:  v.GetBool("ethereum-websocket.enable"),
			Address: v.GetString("ethereum-websocket.address"),
		},
	}
}

package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultEVMAddress is the default address the EVM JSON-RPC server binds to.
	DefaultEVMAddress = "0.0.0.0:1317"

	// DefaultEVMWSAddress is the default address the EVM WebSocket server binds to.
	DefaultEVMWSAddress = "0.0.0.0:1318"
)

// EVMRPCConfig defines configuration for the EVM RPC server.
type EVMRPCConfig struct {
	// Enable defines if the EVM RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// Address defines the HTTP server to listen on
	RPCAddress string `mapstructure:"address"`
	// Address defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
}

// Config defines the server's top level configuration
type Config struct {
	config.BaseConfig `mapstructure:",squash"`

	// Telemetry defines the application telemetry configuration
	Telemetry telemetry.Config       `mapstructure:"telemetry"`
	API       config.APIConfig       `mapstructure:"api"`
	GRPC      config.GRPCConfig      `mapstructure:"grpc"`
	EVMRPC    EVMRPCConfig           `mapstructure:"evm-rpc"`
	StateSync config.StateSyncConfig `mapstructure:"state-sync"`
}

func (c *Config) ToSDKConfig() *config.Config {
	return &config.Config{
		BaseConfig: c.BaseConfig,
		Telemetry:  c.Telemetry,
		API:        c.API,
		GRPC:       c.GRPC,
		StateSync:  c.StateSync,
	}
}

// SetMinGasPrices sets the validator's minimum gas prices.
func (c *Config) SetMinGasPrices(gasPrices sdk.DecCoins) {
	c.MinGasPrices = gasPrices.String()
}

// GetMinGasPrices returns the validator's minimum gas prices based on the set
// configuration.
func (c *Config) GetMinGasPrices() sdk.DecCoins {
	if c.MinGasPrices == "" {
		return sdk.DecCoins{}
	}

	gasPricesStr := strings.Split(c.MinGasPrices, ";")
	gasPrices := make(sdk.DecCoins, len(gasPricesStr))

	for i, s := range gasPricesStr {
		gasPrice, err := sdk.ParseDecCoin(s)
		if err != nil {
			panic(fmt.Errorf("failed to parse minimum gas price coin (%s): %s", s, err))
		}

		gasPrices[i] = gasPrice
	}

	return gasPrices
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	cfg := config.DefaultConfig()
	return &Config{
		BaseConfig: cfg.BaseConfig,
		Telemetry:  cfg.Telemetry,
		API:        cfg.API,
		GRPC:       cfg.GRPC,
		EVMRPC: EVMRPCConfig{
			Enable:     true,
			RPCAddress: DefaultEVMAddress,
			WsAddress:  DefaultEVMWSAddress,
		},
		StateSync: cfg.StateSync,
	}
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) Config {

	cfg := config.GetConfig(v)

	return Config{
		BaseConfig: cfg.BaseConfig,
		Telemetry:  cfg.Telemetry,
		API:        cfg.API,
		GRPC:       cfg.GRPC,
		EVMRPC: EVMRPCConfig{
			Enable:     v.GetBool("evm-rpc.enable"),
			RPCAddress: v.GetString("evm-rpc.address"),
			WsAddress:  v.GetString("evm-rpc.ws-address"),
		},
		StateSync: cfg.StateSync,
	}
}

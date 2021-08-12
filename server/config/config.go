package config

import (
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/viper"
)

const (
	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultEVMAddress is the default address the EVM JSON-RPC server binds to.
	DefaultEVMAddress = "0.0.0.0:8545"

	// DefaultEVMWSAddress is the default address the EVM WebSocket server binds to.
	DefaultEVMWSAddress = "0.0.0.0:8546"
)

// GetDefaultAPINamespaces returns the default list of JSON-RPC namespaces that should be enabled
func GetDefaultAPINamespaces() []string {
	return []string{"eth", "net", "web3"}
}

// AppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func AppConfig(denom string) (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := config.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In ethermint, we set the min gas prices to 0.
	if denom != "" {
		srvCfg.MinGasPrices = "0" + denom
	}

	customAppConfig := Config{
		Config:  *srvCfg,
		JSONRPC: *DefaultEVMConfig(),
	}

	customAppTemplate := config.DefaultConfigTemplate + DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config:  *config.DefaultConfig(),
		JSONRPC: *DefaultEVMConfig(),
	}
}

// EVMConfig defines the application configuration values for the EVM.
type EVMConfig struct {
	// Tracer defines vm.Tracer type that the EVM will use if the node is run in
	// trace mode. Default 'json'
	Tracer string `mapstructure:"tracer"`
}

// JSONRPCConfig defines configuration for the EVM RPC server.
type JSONRPCConfig struct {
	// Address defines the HTTP server to listen on
	Address string `mapstructure:"address"`
	// WsAddress defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
	// API defines a list of JSON-RPC namespaces that should be enabled
	API []string `mapstructure:"api"`
	// Enable defines if the EVM RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk)
	EnableUnsafeCORS bool `mapstructure:"enable-unsafe-cors"`
}

// DefaultEVMConfig returns an EVM config with the JSON-RPC API enabled by default
func DefaultEVMConfig() *JSONRPCConfig {
	return &JSONRPCConfig{
		Enable:           true,
		API:              GetDefaultAPINamespaces(),
		Address:          DefaultEVMAddress,
		WsAddress:        DefaultEVMWSAddress,
		EnableUnsafeCORS: false,
	}
}

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config

	JSONRPC JSONRPCConfig `mapstructure:"json-rpc"`
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) Config {
	cfg := config.GetConfig(v)

	return Config{
		Config: cfg,
		JSONRPC: JSONRPCConfig{
			Enable:           v.GetBool("json-rpc.enable"),
			API:              v.GetStringSlice("json-rpc.api"),
			Address:          v.GetString("json-rpc.address"),
			WsAddress:        v.GetString("json-rpc.ws-address"),
			EnableUnsafeCORS: v.GetBool("json-rpc.enable-unsafe-cors"),
		},
	}
}

package config

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/libs/strings"

	"github.com/cosmos/cosmos-sdk/server/config"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultJSONRPCAddress is the default address the JSON-RPC server binds to.
	DefaultJSONRPCAddress = "0.0.0.0:8545"

	// DefaultJSONRPCWsAddress is the default address the JSON-RPC WebSocket server binds to.
	DefaultJSONRPCWsAddress = "0.0.0.0:8546"

	// DefaultEVMTracer is the default vm.Tracer type
	DefaultEVMTracer = "json"

	DefaultGasCap uint64 = 25000000
)

var (
	evmTracers = []string{DefaultEVMTracer, "markdown", "struct", "access_list"}
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
		EVM:     *DefaultEVMConfig(),
		JSONRPC: *DefaultJSONRPCConfig(),
	}

	customAppTemplate := config.DefaultConfigTemplate + DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// DefaultConfig returns server's default configuration.
func DefaultConfig() *Config {
	return &Config{
		Config:  *config.DefaultConfig(),
		EVM:     *DefaultEVMConfig(),
		JSONRPC: *DefaultJSONRPCConfig(),
	}
}

// EVMConfig defines the application configuration values for the EVM.
type EVMConfig struct {
	// Tracer defines vm.Tracer type that the EVM will use if the node is run in
	// trace mode. Default: 'json'.
	Tracer string `mapstructure:"tracer"`
}

// DefaultEVMConfig returns the default EVM configuration
func DefaultEVMConfig() *EVMConfig {
	return &EVMConfig{
		Tracer: DefaultEVMTracer,
	}
}

// Validate returns an error if the tracer type is invalid.
func (c EVMConfig) Validate() error {
	if !strings.StringInSlice(c.Tracer, evmTracers) {
		return fmt.Errorf("invalid tracer type %s, available types: %v", c.Tracer, evmTracers)
	}

	return nil
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
	// GasCap is the global gas cap for eth-call variants.
	GasCap uint64 `mapstructure:"gas-cap"`
}

// Validate returns an error if the JSON-RPC configuration fields are invalid.
func (c JSONRPCConfig) Validate() error {
	if c.Enable && len(c.API) == 0 {
		return errors.New("cannot enable JSON-RPC without defining any API namespace")
	}

	// TODO: validate APIs
	seenAPIs := make(map[string]bool)
	for _, api := range c.API {
		if seenAPIs[api] {
			return fmt.Errorf("repeated API namespace '%s'", api)
		}

		seenAPIs[api] = true
	}

	return nil
}

// DefaultJSONRPCConfig returns an EVM config with the JSON-RPC API enabled by default
func DefaultJSONRPCConfig() *JSONRPCConfig {
	return &JSONRPCConfig{
		Enable:    true,
		API:       GetDefaultAPINamespaces(),
		Address:   DefaultJSONRPCAddress,
		WsAddress: DefaultJSONRPCWsAddress,
		GasCap:    DefaultGasCap,
	}
}

// Config defines the server's top level configuration. It includes the default app config
// from the SDK as well as the EVM configuration to enable the JSON-RPC APIs.
type Config struct {
	config.Config

	EVM     EVMConfig     `mapstructure:"evm"`
	JSONRPC JSONRPCConfig `mapstructure:"json-rpc"`
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) Config {
	cfg := config.GetConfig(v)

	return Config{
		Config: cfg,
		EVM: EVMConfig{
			Tracer: v.GetString("evm.tracer"),
		},
		JSONRPC: JSONRPCConfig{
			Enable:    v.GetBool("json-rpc.enable"),
			API:       v.GetStringSlice("json-rpc.api"),
			Address:   v.GetString("json-rpc.address"),
			WsAddress: v.GetString("json-rpc.ws-address"),
			GasCap:    v.GetUint64("json-rpc.gas-cap"),
		},
	}
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig(v *viper.Viper) (*Config, error) {
	conf := DefaultConfig()
	err := v.Unmarshal(conf)

	return conf, err
}

// ValidateBasic returns an error any of the application configuration fields are invalid
func (c Config) ValidateBasic() error {
	if err := c.EVM.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrAppConfig, "invalid evm config value: %s", err.Error())
	}

	if err := c.JSONRPC.Validate(); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrAppConfig, "invalid json-rpc config value: %s", err.Error())
	}

	return c.Config.ValidateBasic()
}

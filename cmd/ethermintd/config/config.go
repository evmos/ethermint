package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/telemetry"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	defaultMinGasPrices = ""

	// DefaultGRPCAddress is the default address the gRPC server binds to.
	DefaultGRPCAddress = "0.0.0.0:9900"

	// DefaultEVMAddress is the default address the EVM JSON-RPC server binds to.
	DefaultEVMAddress = "0.0.0.0:1317"

	// DefaultEVMWSAddress is the default address the EVM WebSocket server binds to.
	DefaultEVMWSAddress = "0.0.0.0:1318"
)

// BaseConfig defines the server's basic configuration
type BaseConfig struct {
	// The minimum gas prices a validator is willing to accept for processing a
	// transaction. A transaction's fees must meet the minimum of any denomination
	// specified in this config (e.g. 0.25token1;0.0001token2).
	MinGasPrices string `mapstructure:"minimum-gas-prices"`

	Pruning           string `mapstructure:"pruning"`
	PruningKeepRecent string `mapstructure:"pruning-keep-recent"`
	PruningKeepEvery  string `mapstructure:"pruning-keep-every"`
	PruningInterval   string `mapstructure:"pruning-interval"`

	// HaltHeight contains a non-zero block height at which a node will gracefully
	// halt and shutdown that can be used to assist upgrades and testing.
	//
	// Note: Commitment of state will be attempted on the corresponding block.
	HaltHeight uint64 `mapstructure:"halt-height"`

	// HaltTime contains a non-zero minimum block time (in Unix seconds) at which
	// a node will gracefully halt and shutdown that can be used to assist
	// upgrades and testing.
	//
	// Note: Commitment of state will be attempted on the corresponding block.
	HaltTime uint64 `mapstructure:"halt-time"`

	// MinRetainBlocks defines the minimum block height offset from the current
	// block being committed, such that blocks past this offset may be pruned
	// from Tendermint. It is used as part of the process of determining the
	// ResponseCommit.RetainHeight value during ABCI Commit. A value of 0 indicates
	// that no blocks should be pruned.
	//
	// This configuration value is only responsible for pruning Tendermint blocks.
	// It has no bearing on application state pruning which is determined by the
	// "pruning-*" configurations.
	//
	// Note: Tendermint block pruning is dependant on this parameter in conunction
	// with the unbonding (safety threshold) period, state pruning and state sync
	// snapshot parameters to determine the correct minimum value of
	// ResponseCommit.RetainHeight.
	MinRetainBlocks uint64 `mapstructure:"min-retain-blocks"`

	// InterBlockCache enables inter-block caching.
	InterBlockCache bool `mapstructure:"inter-block-cache"`

	// IndexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	IndexEvents []string `mapstructure:"index-events"`
}

// APIConfig defines the API listener configuration.
type APIConfig = serverconfig.APIConfig

// GRPCConfig defines configuration for the gRPC server.
type GRPCConfig struct {
	// Enable defines if the gRPC server should be enabled.
	Enable bool `mapstructure:"enable"`

	// Address defines the API server to listen on
	Address string `mapstructure:"address"`
}

// EVMRPCConfig defines configuration for the EVM RPC server.
type EVMRPCConfig struct {
	// Enable defines if the EVM RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// Address defines the HTTP server to listen on
	RPCAddress string `mapstructure:"address"`
	// Address defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
}

// StateSyncConfig defines the state sync snapshot configuration.
type StateSyncConfig struct {
	// SnapshotInterval sets the interval at which state sync snapshots are taken.
	// 0 disables snapshots. Must be a multiple of PruningKeepEvery.
	SnapshotInterval uint64 `mapstructure:"snapshot-interval"`

	// SnapshotKeepRecent sets the number of recent state sync snapshots to keep.
	// 0 keeps all snapshots.
	SnapshotKeepRecent uint32 `mapstructure:"snapshot-keep-recent"`
}

// Config defines the server's top level configuration
type Config struct {
	BaseConfig `mapstructure:",squash"`

	// Telemetry defines the application telemetry configuration
	Telemetry telemetry.Config `mapstructure:"telemetry"`
	API       APIConfig        `mapstructure:"api"`
	GRPC      GRPCConfig       `mapstructure:"grpc"`
	EVMRPC    EVMRPCConfig     `mapstructure:"evm-rpc"`
	StateSync StateSyncConfig  `mapstructure:"state-sync"`
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

	return &Config{
		BaseConfig: BaseConfig{
			MinGasPrices:      defaultMinGasPrices,
			InterBlockCache:   true,
			Pruning:           storetypes.PruningOptionNothing,
			PruningKeepRecent: "0",
			PruningKeepEvery:  "0",
			PruningInterval:   "0",
			MinRetainBlocks:   0,
			IndexEvents:       make([]string, 0),
		},
		Telemetry: telemetry.Config{
			Enabled:      false,
			GlobalLabels: [][]string{},
		},
		API: APIConfig{
			Enable:             true,
			Swagger:            true,
			Address:            "tcp://0.0.0.0:10337",
			MaxOpenConnections: 1000,
			RPCReadTimeout:     10,
			RPCMaxBodyBytes:    1000000,
		},
		GRPC: GRPCConfig{
			Enable:  true,
			Address: DefaultGRPCAddress,
		},
		EVMRPC: EVMRPCConfig{
			Enable:     true,
			RPCAddress: DefaultEVMAddress,
			WsAddress:  DefaultEVMWSAddress,
		},
		StateSync: StateSyncConfig{
			SnapshotInterval:   0,
			SnapshotKeepRecent: 2,
		},
	}
}

// GetConfig returns a fully parsed Config object.
func GetConfig(v *viper.Viper) Config {
	globalLabelsRaw := v.Get("telemetry.global-labels").([]interface{})
	globalLabels := make([][]string, 0, len(globalLabelsRaw))

	return Config{
		BaseConfig: BaseConfig{
			MinGasPrices:      v.GetString("minimum-gas-prices"),
			InterBlockCache:   v.GetBool("inter-block-cache"),
			Pruning:           v.GetString("pruning"),
			PruningKeepRecent: v.GetString("pruning-keep-recent"),
			PruningKeepEvery:  v.GetString("pruning-keep-every"),
			PruningInterval:   v.GetString("pruning-interval"),
			HaltHeight:        v.GetUint64("halt-height"),
			HaltTime:          v.GetUint64("halt-time"),
			IndexEvents:       v.GetStringSlice("index-events"),
			MinRetainBlocks:   v.GetUint64("min-retain-blocks"),
		},
		Telemetry: telemetry.Config{
			ServiceName:             v.GetString("telemetry.service-name"),
			Enabled:                 v.GetBool("telemetry.enabled"),
			EnableHostname:          v.GetBool("telemetry.enable-hostname"),
			EnableHostnameLabel:     v.GetBool("telemetry.enable-hostname-label"),
			EnableServiceLabel:      v.GetBool("telemetry.enable-service-label"),
			PrometheusRetentionTime: v.GetInt64("telemetry.prometheus-retention-time"),
			GlobalLabels:            globalLabels,
		},
		API: APIConfig{
			Enable:             v.GetBool("api.enable"),
			Swagger:            v.GetBool("api.swagger"),
			Address:            v.GetString("api.address"),
			MaxOpenConnections: v.GetUint("api.max-open-connections"),
			RPCReadTimeout:     v.GetUint("api.rpc-read-timeout"),
			RPCWriteTimeout:    v.GetUint("api.rpc-write-timeout"),
			RPCMaxBodyBytes:    v.GetUint("api.rpc-max-body-bytes"),
			EnableUnsafeCORS:   v.GetBool("api.enabled-unsafe-cors"),
		},
		GRPC: GRPCConfig{
			Enable:  v.GetBool("grpc.enable"),
			Address: v.GetString("grpc.address"),
		},
		EVMRPC: EVMRPCConfig{
			Enable:     v.GetBool("evm-rpc.enable"),
			RPCAddress: v.GetString("evm-rpc.address"),
			WsAddress:  v.GetString("evm-rpc.ws-address"),
		},
		StateSync: StateSyncConfig{
			SnapshotInterval:   v.GetUint64("state-sync.snapshot-interval"),
			SnapshotKeepRecent: v.GetUint32("state-sync.snapshot-keep-recent"),
		},
	}
}

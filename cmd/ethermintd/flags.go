package main

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tharsis/ethermint/version"
)

const (
	flagLong = "long"
)

func init() {
	infoCmd.Flags().Bool(flagLong, false, "Print full information")
}

var (
	infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Print version info",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(version.Version())
			return nil
		},
	}
)

var (
	fromPassphrase  string
	ethNodeWS       string
	ethNodeHTTP     string
	statsdEnabled   bool
	statsdPrefix    string
	statsdAddress   string
	statsdStuckFunc string
	evmDebug        bool
	logJSON         bool
	logLevel        string
)

// addTxFlags adds common flags for commands to post tx
func addTxFlags(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().String(flags.FlagChainID, "testnet", "Specify Chain ID for sending Tx")
	cmd.PersistentFlags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.PersistentFlags().StringVar(&fromPassphrase, "from-passphrase", "12345678", "Passphrase for private key specified with 'from'")
	cmd.PersistentFlags().StringVar(&ethNodeWS, "eth-node-ws", "ws://localhost:1317", "WebSocket endpoint for an Ethereum node.")
	cmd.PersistentFlags().StringVar(&ethNodeHTTP, "eth-node-http", "http://localhost:1317", "HTTP endpoint for an Ethereum node.")
	cmd.PersistentFlags().BoolVar(&statsdEnabled, "statsd-enabled", false, "Enabled StatsD reporting.")
	cmd.PersistentFlags().StringVar(&statsdPrefix, "statsd-prefix", "ethermintd", "Specify StatsD compatible metrics prefix.")
	cmd.PersistentFlags().StringVar(&statsdAddress, "statsd-address", "localhost:8125", "UDP address of a StatsD compatible metrics aggregator.")
	cmd.PersistentFlags().StringVar(&statsdStuckFunc, "statsd-stuck-func", "5m", "Sets a duration to consider a function to be stuck (e.g. in deadlock).")
	cmd.PersistentFlags().String(flags.FlagFees, "", "Fees to pay along with transaction; eg: 10aphoton")
	cmd.PersistentFlags().String(flags.FlagGasPrices, "", "Gas prices to determine the transaction fee (e.g. 10aphoton)")
	cmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.PersistentFlags().Float64(flags.FlagGasAdjustment, flags.DefaultGasAdjustment, "adjustment factor to be multiplied against the estimate returned by the tx simulation; if the gas limit is set manually this flag is ignored ")
	cmd.PersistentFlags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
	cmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "Use JSON as the output format of the own logger (default: text)")
	cmd.PersistentFlags().BoolVar(&evmDebug, "evm-debug", false, "Enable EVM debug traces")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Sets the level of the own logger (error, warn, info, debug)")
	//cmd.PersistentFlags().Bool(flags.FlagTrustNode, true, "Trust connected full node (don't verify proofs for responses)")
	cmd.PersistentFlags().String(flags.FlagKeyringBackend, keyring.BackendFile, "Select keyring's backend")

	// --gas can accept integers and "simulate"
	//cmd.PersistentFlags().Var(&flags.GasFlagVar, "gas", fmt.Sprintf(
	//	"gas limit to set per-transaction; set to %q to calculate required gas automatically (default %d)",
	//	flags.GasFlagAuto, flags.DefaultGasLimit,
	//))

	//viper.BindPFlag(flags.FlagTrustNode, cmd.Flags().Lookup(flags.FlagTrustNode))

	// TODO: we need to handle the errors for these, decide if we should return error upward and handle
	// nolint: errcheck
	viper.BindPFlag(flags.FlagNode, cmd.Flags().Lookup(flags.FlagNode))
	// nolint: errcheck
	viper.BindPFlag(flags.FlagKeyringBackend, cmd.Flags().Lookup(flags.FlagKeyringBackend))
	// nolint: errcheck
	cmd.MarkFlagRequired(flags.FlagChainID)
	return cmd
}

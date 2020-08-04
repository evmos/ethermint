package client

import (
	"fmt"
	"math/big"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/tendermint/tendermint/libs/cli"
)

// InitConfig adds the chain-id, encoding and output flags to the persistent flag set.
func InitConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	configFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}

	if err := viper.BindPFlag(flags.FlagChainID, cmd.PersistentFlags().Lookup(flags.FlagChainID)); err != nil {
		return err
	}

	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}

	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}

// ValidateChainID wraps a cobra command with a RunE function with base 10 integer chain-id verification.
func ValidateChainID(baseCmd *cobra.Command) *cobra.Command {
	// Copy base run command to be used after chain verification
	baseRunE := baseCmd.RunE

	// Function to replace command's RunE function
	validateFn := func(cmd *cobra.Command, args []string) error {
		chainIDFlag := viper.GetString(flags.FlagChainID)

		// Verify that the chain-id entered is a base 10 integer
		_, ok := new(big.Int).SetString(chainIDFlag, 10)
		if !ok {
			return fmt.Errorf("invalid chainID: %s, must be base-10 integer format", chainIDFlag)
		}

		return baseRunE(cmd, args)
	}

	baseCmd.RunE = validateFn
	return baseCmd
}

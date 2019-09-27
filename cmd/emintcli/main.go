package main

import (
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/ethermint/rpc"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client"
	sdkrpc "github.com/cosmos/cosmos-sdk/client/rpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	emintkeys "github.com/cosmos/ethermint/keys"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	emintapp "github.com/cosmos/ethermint/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := emintapp.MakeCodec()

	authtypes.ModuleCdc = cdc

	// Read in the configuration file for the sdk
	config := sdk.GetConfig()
	// TODO: Remove or change prefix if usable to generate Ethereum address
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "emintcli",
		Short: "Ethermint Client",
	}

	// Add --chain-id to persistent flags and mark it required
	rootCmd.PersistentFlags().String(client.FlagChainID, "", "Chain ID of tendermint node")
	rootCmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		return initConfig(rootCmd)
	}

	// Construct Root Command
	rootCmd.AddCommand(
		sdkrpc.StatusCommand(),
		client.ConfigCmd(emintapp.DefaultCLIHome),
		queryCmd(cdc),
		txCmd(cdc),
		// TODO: Set up rest routes (if included, different from web3 api)
		rpc.Web3RpcCmd(cdc),
		client.LineBreak,
		// TODO: Remove these commands once ethermint keys and genesis set up
		keys.Commands(),
		emintkeys.Commands(),
		client.LineBreak,
	)

	executor := cli.PrepareMainCmd(rootCmd, "EM", emintapp.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func queryCmd(cdc *amino.Codec) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:     "query",
		Aliases: []string{"q"},
		Short:   "Querying subcommands",
	}

	// TODO: Possibly add these query commands from other modules
	queryCmd.AddCommand(
		authcmd.GetAccountCmd(cdc),
		client.LineBreak,
		authcmd.QueryTxsByEventsCmd(cdc),
		authcmd.QueryTxCmd(cdc),
		client.LineBreak,
	)

	// add modules' query commands
	emintapp.ModuleBasics.AddQueryCommands(queryCmd, cdc)

	return queryCmd
}

func txCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "tx",
		Short: "Transactions subcommands",
	}

	txCmd.AddCommand(
		bankcmd.SendTxCmd(cdc),
		client.LineBreak,
		authcmd.GetSignCommand(cdc),
		client.LineBreak,
		authcmd.GetBroadcastCommand(cdc),
		authcmd.GetEncodeCommand(cdc),
		client.LineBreak,
	)

	// add modules' tx commands
	emintapp.ModuleBasics.AddTxCommands(txCmd, cdc)

	return txCmd
}

func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(cli.HomeFlag)
	if err != nil {
		return err
	}

	cfgFile := path.Join(home, "config", "config.toml")
	if _, err := os.Stat(cfgFile); err == nil {
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	}
	if err := viper.BindPFlag(client.FlagChainID, cmd.PersistentFlags().Lookup(client.FlagChainID)); err != nil {
		return err
	}
	if err := viper.BindPFlag(cli.EncodingFlag, cmd.PersistentFlags().Lookup(cli.EncodingFlag)); err != nil {
		return err
	}
	return viper.BindPFlag(cli.OutputFlag, cmd.PersistentFlags().Lookup(cli.OutputFlag))
}

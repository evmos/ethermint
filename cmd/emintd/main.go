package main

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccscli "github.com/cosmos/cosmos-sdk/x/genaccounts/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	sdk "github.com/cosmos/cosmos-sdk/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	emintapp "github.com/cosmos/ethermint/app"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := emintapp.MakeCodec()

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "emintd",
		Short:             "Ethermint App Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}
	// CLI commands to initialize the chain
	rootCmd.AddCommand(
		genutilcli.InitCmd(ctx, cdc, emintapp.ModuleBasics, emintapp.DefaultNodeHome),
		genutilcli.CollectGenTxsCmd(ctx, cdc, genaccounts.AppModuleBasic{}, emintapp.DefaultNodeHome),
		genutilcli.GenTxCmd(ctx, cdc, emintapp.ModuleBasics, staking.AppModuleBasic{}, genaccounts.AppModuleBasic{}, emintapp.DefaultNodeHome, emintapp.DefaultCLIHome),
		genutilcli.ValidateGenesisCmd(ctx, cdc, emintapp.ModuleBasics),

		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccscli.AddGenesisAccountCmd(ctx, cdc, emintapp.DefaultNodeHome, emintapp.DefaultCLIHome),
	)

	// TODO: Add export app state and TM validators commands
	// server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "EM", emintapp.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

// func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
// 	return emintapp.NewEthermintApp(logger, db)
// }

// func exportAppStateAndTMValidators(
// 	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
// ) (json.RawMessage, []tmtypes.GenesisValidator, error) {

// 	// if height != -1 {
// 	// 	emintApp := emintapp.NewEthermintApp(logger, db)
// 	// 	err := emintApp.LoadHeight(height)
// 	// 	if err != nil {
// 	// 		return nil, nil, err
// 	// 	}
// 	// 	return emintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
// 	// }

// 	// emintApp := emintapp.NewEthermintApp(logger, db)

// 	// return emintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
// 	// TODO: Unstub method
// 	return json.RawMessage{}, []tmtypes.GenesisValidator{}, nil
// }

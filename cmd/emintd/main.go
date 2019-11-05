package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/ethermint/app"
	emintapp "github.com/cosmos/ethermint/app"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := emintapp.MakeCodec()

	genutil.ModuleCdc = cdc
	genutiltypes.ModuleCdc = cdc
	authtypes.ModuleCdc = cdc

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
		withChainIDValidation(genutilcli.InitCmd(ctx, cdc, emintapp.ModuleBasics, emintapp.DefaultNodeHome)),
		genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, emintapp.DefaultNodeHome),
		GenTxCmd(
			ctx, cdc, emintapp.ModuleBasics, staking.AppModuleBasic{}, auth.GenesisAccountIterator{}, emintapp.DefaultNodeHome, emintapp.DefaultCLIHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, emintapp.ModuleBasics),

		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
	)

	// Tendermint node base commands
	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "EM", emintapp.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger tmlog.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return emintapp.NewEthermintApp(logger, db, true, 0,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))))
}

func exportAppStateAndTMValidators(
	logger tmlog.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		emintApp := emintapp.NewEthermintApp(logger, db, true, 0)
		err := emintApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return emintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	emintApp := emintapp.NewEthermintApp(logger, db, true, 0)

	return emintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

// Wraps cobra command with a RunE function with integer chain-id verification
func withChainIDValidation(baseCmd *cobra.Command) *cobra.Command {
	// Copy base run command to be used after chain verification
	baseRunE := baseCmd.RunE

	// Function to replace command's RunE function
	chainIDVerify := func(cmd *cobra.Command, args []string) error {
		chainIDFlag := viper.GetString(client.FlagChainID)

		// Verify that the chain-id entered is a base 10 integer
		_, ok := new(big.Int).SetString(chainIDFlag, 10)
		if !ok {
			return fmt.Errorf(
				fmt.Sprintf("Invalid chainID: %s, must be base-10 integer format", chainIDFlag))
		}

		return baseRunE(cmd, args)
	}

	baseCmd.RunE = chainIDVerify
	return baseCmd
}

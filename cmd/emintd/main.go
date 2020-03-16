package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/client/genaccounts"
	emintcrypto "github.com/cosmos/ethermint/crypto"

	abci "github.com/tendermint/tendermint/abci/types"
	tmamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	"github.com/tendermint/tendermint/libs/cli"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func main() {
	cobra.EnableCommandSorting = false

	cdc := app.MakeCodec()

	tmamino.RegisterKeyType(emintcrypto.PubKeySecp256k1{}, emintcrypto.PubKeyAminoName)
	tmamino.RegisterKeyType(emintcrypto.PrivKeySecp256k1{}, emintcrypto.PrivKeyAminoName)

	cryptokeys.CryptoCdc = cdc
	genutil.ModuleCdc = cdc
	genutiltypes.ModuleCdc = cdc
	clientkeys.KeysCdc = cdc

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
		withChainIDValidation(genutilcli.InitCmd(ctx, cdc, app.ModuleBasics, app.DefaultNodeHome)),
		genutilcli.CollectGenTxsCmd(ctx, cdc, auth.GenesisAccountIterator{}, app.DefaultNodeHome),
		genutilcli.GenTxCmd(
			ctx, cdc, app.ModuleBasics, staking.AppModuleBasic{}, auth.GenesisAccountIterator{}, app.DefaultNodeHome, app.DefaultCLIHome,
		),
		genutilcli.ValidateGenesisCmd(ctx, cdc, app.ModuleBasics),

		// AddGenesisAccountCmd allows users to add accounts to the genesis file
		genaccounts.AddGenesisAccountCmd(ctx, cdc, app.DefaultNodeHome, app.DefaultCLIHome),
	)

	// Tendermint node base commands
	server.AddCommands(ctx, cdc, rootCmd, newApp, exportAppStateAndTMValidators)

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "EM", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

func newApp(logger tmlog.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewEthermintApp(logger, db, true, 0,
		baseapp.SetPruning(store.NewPruningOptionsFromString(viper.GetString("pruning"))))
}

func exportAppStateAndTMValidators(
	logger tmlog.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string,
) (json.RawMessage, []tmtypes.GenesisValidator, error) {

	if height != -1 {
		emintApp := app.NewEthermintApp(logger, db, true, 0)
		err := emintApp.LoadHeight(height)
		if err != nil {
			return nil, nil, err
		}
		return emintApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	emintApp := app.NewEthermintApp(logger, db, true, 0)

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
				fmt.Sprintf("invalid chainID: %s, must be base-10 integer format", chainIDFlag))
		}

		return baseRunE(cmd, args)
	}

	baseCmd.RunE = chainIDVerify
	return baseCmd
}

package main

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"

	emintCrypto "github.com/cosmos/ethermint/crypto"
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagDryRun = "dry-run"
)

// keyCommands registers a sub-tree of commands to interact with
// local private key storage.
func keyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}
	addCmd := clientkeys.AddKeyCommand()
	addCmd.RunE = runAddCmd
	cmd.AddCommand(
		clientkeys.MnemonicKeyCommand(),
		addCmd,
		clientkeys.ExportKeyCommand(),
		clientkeys.ImportKeyCommand(),
		clientkeys.ListKeysCmd(),
		clientkeys.ShowKeysCmd(),
		flags.LineBreak,
		clientkeys.DeleteKeyCommand(),
		clientkeys.UpdateKeyCommand(),
		clientkeys.ParseKeyStringCommand(),
		clientkeys.MigrateCommand(),
	)
	return cmd
}

func getKeybase(dryrun bool) (keys.Keybase, error) {
	if dryrun {
		return keys.NewInMemory(keys.WithKeygenFunc(ethermintKeygenFunc)), nil
	}

	return clientkeys.NewKeyBaseFromHomeFlag(keys.WithKeygenFunc(ethermintKeygenFunc))
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	kb, err := getKeybase(viper.GetBool(flagDryRun))
	if err != nil {
		return err
	}

	return clientkeys.RunAddCmd(cmd, args, kb)
}

func ethermintKeygenFunc(bz [32]byte) tmcrypto.PrivKey {
	return emintCrypto.PrivKeySecp256k1(bz[:])
}

package client

import (
	"bufio"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ethermint/crypto"
)

const (
	flagDryRun = "dry-run"
)

// KeyCommands registers a sub-tree of commands to interact with
// local private key storage.
func KeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
		Long: `Keys allows you to manage your local keystore for tendermint.

    These keys may be in any format supported by go-crypto and can be
    used by light-clients, full nodes, or any other application that
    needs to sign with a private key.`,
	}

	// support adding Ethereum supported keys
	addCmd := clientkeys.AddKeyCommand()

	// update the default signing algorithm value to "eth_secp256k1"
	algoFlag := addCmd.Flag("algo")
	algoFlag.DefValue = string(crypto.EthSecp256k1)
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
		clientkeys.ParseKeyStringCommand(),
		clientkeys.MigrateCommand(),
		flags.LineBreak,
		UnsafeExportEthKeyCommand(),
	)
	return cmd
}

func runAddCmd(cmd *cobra.Command, args []string) error {
	inBuf := bufio.NewReader(cmd.InOrStdin())
	kb, err := getKeybase(viper.GetBool(flagDryRun), inBuf)
	if err != nil {
		return err
	}

	return clientkeys.RunAddCmd(cmd, args, kb, inBuf)
}

func getKeybase(transient bool, buf io.Reader) (keys.Keybase, error) {
	if transient {
		return keys.NewInMemory(
			crypto.EthSecp256k1Options()...,
		), nil
	}

	return keys.NewKeyring(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		buf,
		crypto.EthSecp256k1Options()...,
	)
}

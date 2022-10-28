package client

import (
	"bufio"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"

	"github.com/evmos/ethermint/crypto/hd"
)

// UnsafeImportKeyCommand imports private keys from a keyfile.
func UnsafeImportKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unsafe-import-eth-key <name> <pk>",
		Short: "**UNSAFE** Import Ethereum private keys into the local keybase",
		Long:  "**UNSAFE** Import a hex-encoded Ethereum private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	clientCtx := client.GetClientContextFromCmd(cmd).WithKeyringOptions(hd.EthSecp256k1Option())
	clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
	if err != nil {
		return err
	}

	inBuf := bufio.NewReader(cmd.InOrStdin())
	passphrase, err := input.GetPassword("Enter passphrase to encrypt your key:", inBuf)
	if err != nil {
		return err
	}

	privKey := &ethsecp256k1.PrivKey{
		Key: common.FromHex(args[1]),
	}

	armor := crypto.EncryptArmorPrivKey(privKey, passphrase, "eth_secp256k1")

	return clientCtx.Keyring.ImportPrivKey(args[0], armor, passphrase)
}

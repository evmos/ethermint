package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/client/input"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	emintcrypto "github.com/cosmos/ethermint/crypto"
)

func exportEthKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-eth-key <name>",
		Short: "Export an Ethereum private key",
		Long:  `Export an Ethereum private key unencrypted to use in dev tooling **UNSAFE**`,
		Args:  cobra.ExactArgs(1),
		RunE:  runExportCmd,
	}
	return cmd
}

func runExportCmd(cmd *cobra.Command, args []string) error {
	kb, err := clientkeys.NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	// Get password from input or standard input
	buf := bufio.NewReader(cmd.InOrStdin())
	decryptPassword, err := input.GetPassword(
		"**WARNING this is an unsafe way to export your unencrypted private key**\nEnter key password:",
		buf)
	if err != nil {
		return err
	}

	// Exports private key from keybase using password
	privKey, err := kb.ExportPrivateKeyObject(args[0], decryptPassword)
	if err != nil {
		return err
	}

	// Converts key to Ethermint secp256 implementation
	emintKey, ok := privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		return fmt.Errorf("invalid private key type, must be Ethereum key: %T", privKey)
	}

	// Formats key for output
	privB := ethcrypto.FromECDSA(emintKey.ToECDSA())
	keyS := strings.ToUpper(hexutil.Encode(privB)[2:])

	fmt.Println(keyS)

	return nil
}

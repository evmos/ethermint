package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	emintcrypto "github.com/cosmos/ethermint/crypto"
)

func unsafeExportEthKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsafe-export-eth-key [name]",
		Short: "**UNSAFE** Export an Ethereum private key",
		Long:  `**UNSAFE** Export an Ethereum private key unencrypted to use in dev tooling`,
		Args:  cobra.ExactArgs(1),
		RunE:  runExportCmd,
	}
	return cmd
}

func runExportCmd(cmd *cobra.Command, args []string) error {
	inBuf := bufio.NewReader(cmd.InOrStdin())

	keystore, err := keyring.New(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		inBuf,
	)
	if err != nil {
		return err
	}

	decryptPassword := ""
	conf := true
	keyringBackend := viper.GetString(flags.FlagKeyringBackend)
	switch keyringBackend {
	case keyring.BackendFile:
		decryptPassword, err = input.GetPassword(
			"**WARNING this is an unsafe way to export your unencrypted private key**\nEnter key password:",
			inBuf)
	case keyring.BackendOS:
		conf, err = input.GetConfirmation(
			"**WARNING** this is an unsafe way to export your unencrypted private key, are you sure?",
			inBuf, cmd.ErrOrStderr())
	}
	if err != nil || !conf {
		return err
	}

	// Exports private key from keyring using password
	armored, err := keystore.ExportPrivKeyArmor(args[0], decryptPassword)
	if err != nil {
		return err
	}

	privKey, _, err := crypto.UnarmorDecryptPrivKey(armored, decryptPassword)
	if err != nil {
		return err
	}

	// Converts key to Ethermint secp256 implementation
	emintKey, ok := privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		return fmt.Errorf("invalid private key type %T, must be Ethereum key PrivKeySecp256k1", privKey)
	}

	// Formats key for output
	privB := ethcrypto.FromECDSA(emintKey.ToECDSA())
	keyS := strings.ToUpper(hexutil.Encode(privB)[2:])

	fmt.Println(keyS)

	return nil
}

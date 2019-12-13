package rpc

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	emintkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/codec"
	emintcrypto "github.com/cosmos/ethermint/crypto"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagUnlockKey = "unlock-key"
)

// Config contains configuration fields that determine the behavior of the RPC HTTP server.
// TODO: These may become irrelevant if HTTP config is handled by the SDK
type Config struct {
	// EnableRPC defines whether or not to enable the RPC server
	EnableRPC bool
	// RPCAddr defines the IP address to listen on
	RPCAddr string
	// RPCPort defines the port to listen on
	RPCPort int
	// RPCCORSDomains defines list of domains to enable CORS headers for (used by browsers)
	RPCCORSDomains []string
	// RPCVhosts defines list of domains to listen on (useful if Tendermint is addressable via DNS)
	RPCVHosts []string
}

// Web3RpcCmd creates a CLI command to start RPC server
func Web3RpcCmd(cdc *codec.Codec) *cobra.Command {
	cmd := lcd.ServeCommand(cdc, registerRoutes)
	cmd.Flags().String(flagUnlockKey, "", "Select a key to unlock on the RPC server")
	cmd.Flags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
	return cmd
}

// registerRoutes creates a new server and registers the `/rpc` endpoint.
// Rpc calls are enabled based on their associated module (eg. "eth").
func registerRoutes(rs *lcd.RestServer) {
	s := rpc.NewServer()
	accountName := viper.GetString(flagUnlockKey)

	var emintKey emintcrypto.PrivKeySecp256k1
	if len(accountName) > 0 {
		var err error
		buf := bufio.NewReader(os.Stdin)
		keyringBackend := viper.GetString(flags.FlagKeyringBackend)
		passphrase := ""
		switch keyringBackend {
		case flags.KeyringBackendOS:
			break
		case flags.KeyringBackendFile:
			passphrase, err = input.GetPassword("Enter password to unlock key for RPC API: ", buf)
			if err != nil {
				panic(err)
			}
		}

		emintKey, err = unlockKeyFromNameAndPassphrase(accountName, passphrase)
		if err != nil {
			panic(err)
		}
	}

	apis := GetRPCAPIs(rs.CliCtx, emintKey)

	// TODO: Allow cli to configure modules https://github.com/ChainSafe/ethermint/issues/74
	whitelist := make(map[string]bool)

	// Register all the APIs exposed by the services
	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := s.RegisterName(api.Namespace, api.Service); err != nil {
				panic(err)
			}
		}
	}

	rs.Mux.HandleFunc("/", s.ServeHTTP).Methods("POST", "OPTIONS")
}

func unlockKeyFromNameAndPassphrase(accountName, passphrase string) (emintKey emintcrypto.PrivKeySecp256k1, err error) {
	keybase, err := emintkeys.NewKeyringFromHomeFlag(os.Stdin)
	if err != nil {
		return
	}

	// With keyring keybase, password is not required as it is pulled from the OS prompt
	privKey, err := keybase.ExportPrivateKeyObject(accountName, passphrase)
	if err != nil {
		return
	}

	var ok bool
	emintKey, ok = privKey.(emintcrypto.PrivKeySecp256k1)
	if !ok {
		panic(fmt.Sprintf("invalid private key type: %T", privKey))
	}

	return
}

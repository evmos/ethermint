package rpc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	flagUnlockKey = "unlock-key"
	flagWebsocket = "wsport"
)

// EmintServeCmd creates a CLI command to start Cosmos REST server with web3 RPC API and
// Cosmos rest-server endpoints
func EmintServeCmd(cdc *codec.Codec) *cobra.Command {
	cmd := lcd.ServeCommand(cdc, registerRoutes)
	cmd.Flags().String(flagUnlockKey, "", "Select a key to unlock on the RPC server")
	cmd.Flags().String(flagWebsocket, "8546", "websocket port to listen to")
	cmd.Flags().StringP(flags.FlagBroadcastMode, "b", flags.BroadcastSync, "Transaction broadcasting mode (sync|async|block)")
	return cmd
}

// registerRoutes creates a new server and registers the `/rpc` endpoint.
// Rpc calls are enabled based on their associated module (eg. "eth").
func registerRoutes(rs *lcd.RestServer) {
	s := rpc.NewServer()
	accountName := viper.GetString(flagUnlockKey)
	accountNames := strings.Split(accountName, ",")

	var privkeys []crypto.PrivKeySecp256k1
	if len(accountName) > 0 {
		var err error
		inBuf := bufio.NewReader(os.Stdin)

		keyringBackend := viper.GetString(flags.FlagKeyringBackend)
		passphrase := ""
		switch keyringBackend {
		case keys.BackendOS:
			break
		case keys.BackendFile:
			passphrase, err = input.GetPassword(
				"Enter password to unlock key for RPC API: ",
				inBuf)
			if err != nil {
				panic(err)
			}
		}

		privkeys, err = unlockKeyFromNameAndPassphrase(accountNames, passphrase)
		if err != nil {
			panic(err)
		}
	}

	apis := GetRPCAPIs(rs.CliCtx, privkeys)

	// TODO: Allow cli to configure modules https://github.com/ChainSafe/ethermint/issues/74
	whitelist := make(map[string]bool)

	// Register all the APIs exposed by the services
	for _, api := range apis {
		if whitelist[api.Namespace] || (len(whitelist) == 0 && api.Public) {
			if err := s.RegisterName(api.Namespace, api.Service); err != nil {
				panic(err)
			}
		} else if !api.Public { // TODO: how to handle private apis? should only accept local calls
			if err := s.RegisterName(api.Namespace, api.Service); err != nil {
				panic(err)
			}
		}
	}

	// Web3 RPC API route
	rs.Mux.HandleFunc("/", s.ServeHTTP).Methods("POST", "OPTIONS")

	// Register all other Cosmos routes
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)

	// start websockets server
	websocketAddr := viper.GetString(flagWebsocket)
	ws := newWebsocketsServer(rs.CliCtx, websocketAddr)
	ws.start()
}

func unlockKeyFromNameAndPassphrase(accountNames []string, passphrase string) ([]crypto.PrivKeySecp256k1, error) {
	keybase, err := keys.NewKeyring(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		os.Stdin,
		crypto.EthSecp256k1Options()...,
	)
	if err != nil {
		return []crypto.PrivKeySecp256k1{}, err
	}

	// try the for loop with array []string accountNames
	// run through the bottom code inside the for loop

	keys := make([]crypto.PrivKeySecp256k1, len(accountNames))
	for i, acc := range accountNames {
		// With keyring keybase, password is not required as it is pulled from the OS prompt
		privKey, err := keybase.ExportPrivateKeyObject(acc, passphrase)
		if err != nil {
			return []crypto.PrivKeySecp256k1{}, err
		}

		var ok bool
		keys[i], ok = privKey.(crypto.PrivKeySecp256k1)
		if !ok {
			panic(fmt.Sprintf("invalid private key type %T at index %d", privKey, i))
		}
	}

	return keys, nil
}

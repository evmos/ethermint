package rpc

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"

	"github.com/cosmos/ethermint/app"
	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/crypto/hd"
	"github.com/cosmos/ethermint/rpc/websockets"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	flagUnlockKey = "unlock-key"
	flagWebsocket = "wsport"
)

// RegisterRoutes creates a new server and registers the `/rpc` endpoint.
// Rpc calls are enabled based on their associated module (eg. "eth").
func RegisterRoutes(rs *lcd.RestServer) {
	server := rpc.NewServer()
	accountName := viper.GetString(flagUnlockKey)
	accountNames := strings.Split(accountName, ",")

	var privkeys []ethsecp256k1.PrivKey
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

	apis := GetAPIs(rs.CliCtx, privkeys...)

	// Register all the APIs exposed by the namespace services
	// TODO: handle allowlist and private APIs
	for _, api := range apis {
		if err := server.RegisterName(api.Namespace, api.Service); err != nil {
			panic(err)
		}
	}

	// Web3 RPC API route
	rs.Mux.HandleFunc("/", server.ServeHTTP).Methods("POST", "OPTIONS")

	// Register all other Cosmos routes
	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	authrest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	app.ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)

	// start websockets server
	websocketAddr := viper.GetString(flagWebsocket)
	ws := websockets.NewServer(rs.CliCtx, websocketAddr)
	ws.Start()
}

func unlockKeyFromNameAndPassphrase(accountNames []string, passphrase string) ([]ethsecp256k1.PrivKey, error) {
	keybase, err := keys.NewKeyring(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		viper.GetString(flags.FlagHome),
		os.Stdin,
		hd.EthSecp256k1Options()...,
	)
	if err != nil {
		return []ethsecp256k1.PrivKey{}, err
	}

	// try the for loop with array []string accountNames
	// run through the bottom code inside the for loop

	keys := make([]ethsecp256k1.PrivKey, len(accountNames))
	for i, acc := range accountNames {
		// With keyring keybase, password is not required as it is pulled from the OS prompt
		privKey, err := keybase.ExportPrivateKeyObject(acc, passphrase)
		if err != nil {
			return []ethsecp256k1.PrivKey{}, err
		}

		var ok bool
		keys[i], ok = privKey.(ethsecp256k1.PrivKey)
		if !ok {
			panic(fmt.Sprintf("invalid private key type %T at index %d", privKey, i))
		}
	}

	return keys, nil
}

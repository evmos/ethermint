package keys

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"gopkg.in/yaml.v2"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"

	"github.com/cosmos/cosmos-sdk/client/flags"
	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	emintKeys "github.com/cosmos/ethermint/crypto/keys"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"

	defaultKeyDBName = "emintkeys"
)

type bechKeyOutFn func(keyInfo cosmosKeys.Info) (emintKeys.KeyOutput, error)

// GetKeyInfo returns key info for a given name. An error is returned if the
// keybase cannot be retrieved or getting the info fails.
func GetKeyInfo(name string) (cosmosKeys.Info, error) {
	keybase, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return nil, err
	}

	return keybase.Get(name)
}

// NewKeyBaseFromHomeFlag initializes a Keybase based on the configuration.
func NewKeyBaseFromHomeFlag() (cosmosKeys.Keybase, error) {
	rootDir := viper.GetString(flags.FlagHome)
	return NewKeyBaseFromDir(rootDir)
}

// NewKeyBaseFromDir initializes a keybase at a particular dir.
func NewKeyBaseFromDir(rootDir string) (cosmosKeys.Keybase, error) {
	return getLazyKeyBaseFromDir(rootDir)
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase() cosmosKeys.Keybase { return emintKeys.NewInMemory() }

func getLazyKeyBaseFromDir(rootDir string) (cosmosKeys.Keybase, error) {
	return emintKeys.New(defaultKeyDBName, filepath.Join(rootDir, defaultKeyDBName)), nil
}

// GetPassphrase returns a passphrase for a given name. It will first retrieve
// the key info for that name if the type is local, it'll fetch input from
// STDIN. Otherwise, an empty passphrase is returned. An error is returned if
// the key info cannot be fetched or reading from STDIN fails.
func GetPassphrase(name string) (string, error) {
	var passphrase string

	keyInfo, err := GetKeyInfo(name)
	if err != nil {
		return passphrase, err
	}

	// we only need a passphrase for locally stored keys
	if keyInfo.GetType().String() == emintKeys.TypeLocal.String() {
		passphrase, err = clientkeys.ReadPassphraseFromStdin(name)
		if err != nil {
			return passphrase, err
		}
	}

	return passphrase, nil
}

func printKeyInfo(keyInfo cosmosKeys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos([]emintKeys.KeyOutput{ko})

	case OutputFormatJSON:
		var out []byte
		var err error
		if viper.GetBool(flags.FlagIndentResponse) {
			out, err = cdc.MarshalJSONIndent(ko, "", "  ")
		} else {
			out, err = cdc.MarshalJSON(ko)
		}
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	}
}

// func printInfos(infos []keys.Info) {
// 	kos, err := keys.Bech32KeysOutput(infos)
// 	if err != nil {
// 		panic(err)
// 	}

// 	switch viper.Get(cli.OutputFlag) {
// 	case OutputFormatText:
// 		printTextInfos(kos)

// 	case OutputFormatJSON:
// 		var out []byte
// 		var err error

// 		if viper.GetBool(flags.FlagIndentResponse) {
// 			out, err = cdc.MarshalJSONIndent(kos, "", "  ")
// 		} else {
// 			out, err = cdc.MarshalJSON(kos)
// 		}

// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Printf("%s", out)
// 	}
// }

func printTextInfos(kos []emintKeys.KeyOutput) {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

func printKeyAddress(info cosmosKeys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.Address)
}

func printPubKey(info cosmosKeys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.PubKey)
}

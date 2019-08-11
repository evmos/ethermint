package keys

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
)

// available output formats.
const (
	OutputFormatText = "text"
	OutputFormatJSON = "json"
)

type bechKeyOutFn func(keyInfo cosmosKeys.Info) (cosmosKeys.KeyOutput, error)

// GetKeyInfo returns key info for a given name. An error is returned if the
// keybase cannot be retrieved or getting the info fails.
func GetKeyInfo(name string) (cosmosKeys.Info, error) {
	keybase, err := clientkeys.NewKeyBaseFromHomeFlag()
	if err != nil {
		return nil, err
	}

	return keybase.Get(name)
}

func printKeyInfo(keyInfo cosmosKeys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
	if err != nil {
		panic(err)
	}

	switch viper.Get(cli.OutputFlag) {
	case OutputFormatText:
		printTextInfos([]cosmosKeys.KeyOutput{ko})

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

func printTextInfos(kos []cosmosKeys.KeyOutput) {
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

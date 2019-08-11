package keys

import (
	"encoding/hex"

	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
)

// KeyOutput defines a structure wrapping around an Info object used for output
// functionality.
type KeyOutput struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Address   string `json:"address"`
	PubKey    string `json:"pubkey"`
	Mnemonic  string `json:"mnemonic,omitempty"`
	Threshold uint   `json:"threshold,omitempty"`
}

// Bech32KeysOutput returns a slice of KeyOutput objects, each with the "acc"
// Bech32 prefixes, given a slice of Info objects. It returns an error if any
// call to Bech32KeyOutput fails.
func Bech32KeysOutput(infos []cosmosKeys.Info) ([]cosmosKeys.KeyOutput, error) {
	kos := make([]cosmosKeys.KeyOutput, len(infos))
	for i, info := range infos {
		ko, err := Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
		kos[i] = ko
	}

	return kos, nil
}

// Bech32ConsKeyOutput create a KeyOutput in with "cons" Bech32 prefixes.
func Bech32ConsKeyOutput(keyInfo cosmosKeys.Info) (cosmosKeys.KeyOutput, error) {
	consAddr := keyInfo.GetPubKey().Address()
	bytes := keyInfo.GetPubKey().Bytes()

	// bechPubKey, err := sdk.Bech32ifyConsPub(keyInfo.GetPubKey())
	// if err != nil {
	// 	return KeyOutput{}, err
	// }

	return cosmosKeys.KeyOutput{
		Name:    keyInfo.GetName(),
		Type:    keyInfo.GetType().String(),
		Address: consAddr.String(),
		PubKey:  hex.EncodeToString(bytes),
	}, nil
}

// Bech32ValKeyOutput create a KeyOutput in with "val" Bech32 prefixes.
func Bech32ValKeyOutput(keyInfo cosmosKeys.Info) (cosmosKeys.KeyOutput, error) {
	valAddr := keyInfo.GetPubKey().Address()
	bytes := keyInfo.GetPubKey().Bytes()

	// bechPubKey, err := sdk.Bech32ifyValPub(keyInfo.GetPubKey())
	// if err != nil {
	// 	return KeyOutput{}, err
	// }

	return cosmosKeys.KeyOutput{
		Name:    keyInfo.GetName(),
		Type:    keyInfo.GetType().String(),
		Address: valAddr.String(),
		PubKey:  hex.EncodeToString(bytes),
	}, nil
}

// Bech32KeyOutput create a KeyOutput in with "acc" Bech32 prefixes.
func Bech32KeyOutput(info cosmosKeys.Info) (cosmosKeys.KeyOutput, error) {
	accAddr := info.GetPubKey().Address()
	bytes := info.GetPubKey().Bytes()

	ko := cosmosKeys.KeyOutput{
		Name:    info.GetName(),
		Type:    info.GetType().String(),
		Address: accAddr.String(),
		PubKey:  hex.EncodeToString(bytes),
	}

	return ko, nil
}

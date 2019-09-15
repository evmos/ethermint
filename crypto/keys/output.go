package keys

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
)

// KeyOutput defines a structure wrapping around an Info object used for output
// functionality.
type KeyOutput struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Address    string `json:"address"`
	ETHAddress string `json:"ethaddress"`
	PubKey     string `json:"pubkey"`
	ETHPubKey  string `json:"ethpubkey"`
	Mnemonic   string `json:"mnemonic,omitempty"`
	Threshold  uint   `json:"threshold,omitempty"`
}

// NewKeyOutput creates a default KeyOutput instance without Mnemonic, Threshold and PubKeys
func NewKeyOutput(name, keyType, address, ethaddress, pubkey, ethpubkey string) KeyOutput {
	return KeyOutput{
		Name:       name,
		Type:       keyType,
		Address:    address,
		ETHAddress: ethaddress,
		PubKey:     pubkey,
		ETHPubKey:  ethpubkey,
	}
}

// Bech32KeysOutput returns a slice of KeyOutput objects, each with the "acc"
// Bech32 prefixes, given a slice of Info objects. It returns an error if any
// call to Bech32KeyOutput fails.
func Bech32KeysOutput(infos []cosmosKeys.Info) ([]KeyOutput, error) {
	kos := make([]KeyOutput, len(infos))
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
func Bech32ConsKeyOutput(keyInfo cosmosKeys.Info) (KeyOutput, error) {
	address := keyInfo.GetPubKey().Address()

	bechPubKey, err := sdk.Bech32ifyConsPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return NewKeyOutput(
		keyInfo.GetName(),
		keyInfo.GetType().String(),
		sdk.ConsAddress(address.Bytes()).String(),
		getEthAddress(keyInfo),
		bechPubKey,
		hex.EncodeToString(keyInfo.GetPubKey().Bytes()),
	), nil
}

// Bech32ValKeyOutput create a KeyOutput in with "val" Bech32 prefixes.
func Bech32ValKeyOutput(keyInfo cosmosKeys.Info) (KeyOutput, error) {
	address := keyInfo.GetPubKey().Address()

	bechPubKey, err := sdk.Bech32ifyValPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return NewKeyOutput(
		keyInfo.GetName(),
		keyInfo.GetType().String(),
		sdk.ValAddress(address.Bytes()).String(),
		getEthAddress(keyInfo),
		bechPubKey,
		hex.EncodeToString(keyInfo.GetPubKey().Bytes()),
	), nil
}

// Bech32KeyOutput create a KeyOutput in with "acc" Bech32 prefixes.
func Bech32KeyOutput(keyInfo cosmosKeys.Info) (KeyOutput, error) {
	address := keyInfo.GetPubKey().Address()

	bechPubKey, err := sdk.Bech32ifyAccPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return NewKeyOutput(
		keyInfo.GetName(),
		keyInfo.GetType().String(),
		sdk.AccAddress(address.Bytes()).String(),
		getEthAddress(keyInfo),
		bechPubKey,
		hex.EncodeToString(keyInfo.GetPubKey().Bytes()),
	), nil
}

func getEthAddress(info cosmosKeys.Info) string {
	return info.GetPubKey().Address().String()
}

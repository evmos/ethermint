package encoding

import (
	emintcrypto "github.com/cosmos/ethermint/crypto"
	amino "github.com/tendermint/go-amino"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

var cdc = amino.NewCodec()

func init() {
	RegisterAmino(cdc)
}

// RegisterAmino registers all crypto related types in the given (amino) codec.
func RegisterAmino(cdc *amino.Codec) {
	// These are all written here instead of
	cdc.RegisterInterface((*tmcrypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(emintcrypto.PubKeySecp256k1{}, emintcrypto.PubKeyAminoName, nil)

	cdc.RegisterInterface((*tmcrypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(emintcrypto.PrivKeySecp256k1{}, emintcrypto.PrivKeyAminoName, nil)
}

// PrivKeyFromBytes unmarshalls emint private key from encoded bytes
func PrivKeyFromBytes(privKeyBytes []byte) (privKey tmcrypto.PrivKey, err error) {
	err = cdc.UnmarshalBinaryBare(privKeyBytes, &privKey)
	return
}

// PubKeyFromBytes unmarshalls emint public key from encoded bytes
func PubKeyFromBytes(pubKeyBytes []byte) (pubKey tmcrypto.PubKey, err error) {
	err = cdc.UnmarshalBinaryBare(pubKeyBytes, &pubKey)
	return
}

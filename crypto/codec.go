package crypto

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cryptoCodec = codec.New()

const (
	// Amino encoding names
	PrivKeyAminoName = "crypto/PrivKeySecp256k1"
	PubKeyAminoName  = "crypto/PubKeySecp256k1"
)

func init() {
	RegisterCodec(cryptoCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PubKeySecp256k1{}, PubKeyAminoName, nil)
	cdc.RegisterConcrete(PrivKeySecp256k1{}, PrivKeyAminoName, nil)
}

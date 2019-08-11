package crypto

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cryptoCodec = codec.New()

func init() {
	RegisterCodec(cryptoCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PubKeySecp256k1{}, "crypto/PubKeySecp256k1", nil)
	cdc.RegisterConcrete(PrivKeySecp256k1{}, "crypto/PrivKeySecp256k1", nil)
}

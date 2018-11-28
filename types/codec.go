package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var typesCodec = codec.New()

func init() {
	RegisterCodec(typesCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&Account{}, "types/Account", nil)
}

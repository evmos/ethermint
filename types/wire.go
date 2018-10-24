package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var typesCodec = codec.New()

func init() {
	RegisterCodec(typesCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(&Transaction{}, "types/Transaction", nil)
	cdc.RegisterConcrete(&Account{}, "types/Account", nil)
}

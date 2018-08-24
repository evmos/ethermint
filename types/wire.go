package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

var typesCodec = wire.NewCodec()

func init() {
	RegisterWire(typesCodec)
}

// RegisterWire registers all the necessary types with amino for the given
// codec.
func RegisterWire(codec *wire.Codec) {
	sdk.RegisterWire(codec)
	codec.RegisterConcrete(&Account{}, "types/Account", nil)
	codec.RegisterConcrete(&EthSignature{}, "types/EthSignature", nil)
	codec.RegisterConcrete(TxData{}, "types/TxData", nil)
	codec.RegisterConcrete(Transaction{}, "types/Transaction", nil)
}

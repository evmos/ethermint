package types

import "github.com/cosmos/cosmos-sdk/codec"

var msgCodec = codec.New()

func init() {
	cdc := codec.New()

	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	msgCodec = cdc.Seal()
}

// Register concrete types and interfaces on the given codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgEthereumTx{}, "ethermint/MsgEthereumTx", nil)
}

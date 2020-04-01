package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc defines the codec to be used by evm module
var ModuleCdc = codec.New()

func init() {
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types and interfaces on the given codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgEthereumTx{}, "ethermint/MsgEthereumTx", nil)
	cdc.RegisterConcrete(MsgEthermint{}, "ethermint/MsgEthermint", nil)
	cdc.RegisterConcrete(EncodableTxData{}, "ethermint/EncodableTxData", nil)
}

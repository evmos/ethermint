package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// ModuleCdc defines the evm module's codec
var ModuleCdc = codec.New()

// RegisterCodec registers all the necessary types and interfaces for the
// evm module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgEthereumTx{}, "ethermint/MsgEthereumTx", nil)
	cdc.RegisterConcrete(MsgEthermint{}, "ethermint/MsgEthermint", nil)
	cdc.RegisterConcrete(EncodableTxData{}, "ethermint/EncodableTxData", nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}

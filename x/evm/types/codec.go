package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/ethermint/crypto"
)

var ModuleCdc = codec.New()

func init() {
	cdc := codec.New()

	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	ModuleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types and interfaces on the given codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&EthereumTxMsg{}, "ethermint/MsgEthereumTx", nil)
	crypto.RegisterCodec(cdc)
}

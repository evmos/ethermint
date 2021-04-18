package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	ExtensionOptionsEthereumTxI interface{}
	ExtensionOptionsWeb3TxI     interface{}
)

// RegisterInterfaces registers the client interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgEthereumTx{},
	)

	registry.RegisterInterface("injective.evm.v1beta1.ExtensionOptionsEthereumTx", (*ExtensionOptionsEthereumTxI)(nil))
	registry.RegisterImplementations(
		(*ExtensionOptionsEthereumTxI)(nil),
		&ExtensionOptionsEthereumTx{},
	)

	registry.RegisterInterface("injective.evm.v1beta1.ExtensionOptionsWeb3Tx", (*ExtensionOptionsWeb3TxI)(nil))
	registry.RegisterImplementations(
		(*ExtensionOptionsWeb3TxI)(nil),
		&ExtensionOptionsWeb3Tx{},
	)
}

var (
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

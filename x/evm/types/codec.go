package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	proto "github.com/gogo/protobuf/proto"
)

var (
	ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
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
	registry.RegisterInterface(
		"ethermint.evm.v1alpha1.ExtensionOptionsEthereumTx",
		(*ExtensionOptionsEthereumTxI)(nil),
		&ExtensionOptionsEthereumTx{},
	)
	registry.RegisterInterface(
		"ethermint.evm.v1alpha1.ExtensionOptionsWeb3Tx",
		(*ExtensionOptionsWeb3TxI)(nil),
		&ExtensionOptionsWeb3Tx{},
	)
	registry.RegisterInterface(
		"ethermint.evm.v1alpha1.TxData",
		(*TxData)(nil),
		&AccessListTx{},
		&LegacyTx{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// PackClientState constructs a new Any packed with the given tx data value. It returns
// an error if the client state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackTxData(txData TxData) (*codectypes.Any, error) {
	msg, ok := txData.(proto.Message)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", txData)
	}

	anyTxData, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, err.Error())
	}

	return anyTxData, nil
}

// UnpackTxData unpacks an Any into a TxData. It returns an error if the
// client state can't be unpacked into a TxData.
func UnpackTxData(any *codectypes.Any) (TxData, error) {
	if any == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	txData, ok := any.GetCachedValue().(TxData)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnpackAny, "cannot unpack Any into TxData %T", any)
	}

	return txData, nil
}

package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (m QueryTraceTxRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Msg.UnpackInterfaces(unpacker)
}

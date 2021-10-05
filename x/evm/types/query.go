package types

import (
	"math/big"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func (cr EthCallRequest) GetBaseFee() *big.Int {
	if cr.BaseFee == nil {
		return nil
	}

	return cr.BaseFee.BigInt()
}

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (m QueryTraceTxRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return m.Msg.UnpackInterfaces(unpacker)
}

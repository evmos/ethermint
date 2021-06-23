package encoding

import (
	"github.com/cosmos/cosmos-sdk/client"
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	enccodec "github.com/tharsis/ethermint/encoding/codec"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeConfig(mb module.BasicManager) params.EncodingConfig {
	cdc := amino.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          NewTxConfig(marshaler),
		Amino:             cdc,
	}

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

type txConfig struct {
	cdc amino.ProtoCodecMarshaler
	client.TxConfig
}

// NewTxConfig returns a new protobuf TxConfig using the provided ProtoCodec and sign modes. The
// first enabled sign mode will become the default sign mode.
func NewTxConfig(marshaler amino.ProtoCodecMarshaler) client.TxConfig {
	return &txConfig{
		marshaler,
		tx.NewTxConfig(marshaler, tx.DefaultSignModes),
	}
}

// TxEncoder overwrites sdk.TxEncoder to support MsgEthereumTx
func (g txConfig) TxEncoder() sdk.TxEncoder {
	return func(tx sdk.Tx) ([]byte, error) {
		msg, ok := tx.(*evmtypes.MsgEthereumTx)
		if ok {
			return msg.AsTransaction().MarshalBinary()
		}
		return g.TxConfig.TxEncoder()(tx)
	}
}

// TxDecoder overwrites sdk.TxDecoder to support MsgEthereumTx
func (g txConfig) TxDecoder() sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		tx := &ethtypes.Transaction{}

		err := tx.UnmarshalBinary(txBytes)
		if err == nil {
			msg := &evmtypes.MsgEthereumTx{}
			msg.FromEthereumTx(tx)
			return msg, nil
		}

		return g.TxConfig.TxDecoder()(txBytes)
	}
}

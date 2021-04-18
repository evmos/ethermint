package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/cosmos/ethermint/codec"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	cdc := amino.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          NewTxConfig(marshaler),
		Amino:             cdc,
	}

	codec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	codec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
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
		ethtx, ok := tx.(*evmtypes.MsgEthereumTx)
		if ok {
			return g.cdc.MarshalBinaryBare(ethtx)
		}
		return g.TxConfig.TxEncoder()(tx)
	}
}

// TxDecoder overwrites sdk.TxDecoder to support MsgEthereumTx
func (g txConfig) TxDecoder() sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var ethtx evmtypes.MsgEthereumTx

		err := g.cdc.UnmarshalBinaryBare(txBytes, &ethtx)
		if err == nil {
			return &ethtx, nil
		}

		return g.TxConfig.TxDecoder()(txBytes)
	}
}

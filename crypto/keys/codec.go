package keys

import (
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/codec"
	cosmosKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	emintCrypto "github.com/cosmos/ethermint/crypto"
)

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	cryptoAmino.RegisterAmino(cdc)
	cdc.RegisterInterface((*cosmosKeys.Info)(nil), nil)
	emintCrypto.RegisterCodec(cdc)
	cdc.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	cdc.RegisterConcrete(localInfo{}, "crypto/keys/localInfo", nil)
	cdc.RegisterConcrete(ledgerInfo{}, "crypto/keys/ledgerInfo", nil)
	cdc.RegisterConcrete(offlineInfo{}, "crypto/keys/offlineInfo", nil)
	// cdc.RegisterConcrete(multiInfo{}, "crypto/keys/multiInfo", nil)
	cdc.Seal()
}

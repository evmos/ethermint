package crypto

import (
	cryptoamino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

// CryptoCodec is the default amino codec used by ethermint
var CryptoCodec = codec.New()

// Amino encoding names
const (
	PrivKeyAminoName = "ethermint/PrivKeySecp256k1"
	PubKeyAminoName  = "ethermint/PubKeySecp256k1"
)

func init() {
	// replace the keyring codec with the ethermint crypto codec to prevent
	// amino panics because of unregistered Priv/PubKey
	keyring.CryptoCdc = CryptoCodec
	keyring.RegisterCodec(CryptoCodec)
	cryptoamino.RegisterAmino(CryptoCodec)
	RegisterCodec(CryptoCodec)
}

// RegisterCodec registers all the necessary types with amino for the given
// codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(PubKeySecp256k1{}, PubKeyAminoName, nil)
	cdc.RegisterConcrete(PrivKeySecp256k1{}, PrivKeyAminoName, nil)
}

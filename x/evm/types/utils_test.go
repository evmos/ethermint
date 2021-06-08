package types

import (
	"testing"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// GenerateEthAddress generates an Ethereum address.
func GenerateEthAddress() ethcmn.Address {
	priv, err := ethsecp256k1.GenerateKey()
	if err != nil {
		panic(err)
	}

	return ethcrypto.PubkeyToAddress(priv.ToECDSA().PublicKey)
}

func TestEvmDataEncoding(t *testing.T) {
	ret := []byte{0x5, 0x8}

	data := &MsgEthereumTxResponse{
		Logs: []*Log{{
			Data:        []byte{1, 2, 3, 4},
			BlockNumber: 17,
		}},
		Ret: ret,
	}

	enc, err := EncodeTxResponse(data)
	require.NoError(t, err)

	res, err := DecodeTxResponse(enc)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, data.Logs, res.Logs)
	require.Equal(t, ret, res.Ret)
}

package types

import (
	"testing"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"

	"github.com/stretchr/testify/require"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
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
	addr := "0x5dE8a020088a2D6d0a23c204FFbeD02790466B49"
	bloom := ethtypes.BytesToBloom([]byte{0x1, 0x3})
	ret := []byte{0x5, 0x8}

	data := &MsgEthereumTxResponse{
		ContractAddress: addr,
		Bloom:           bloom.Bytes(),
		TxLogs: TransactionLogs{
			Hash: ethcmn.BytesToHash([]byte{1, 2, 3, 4}).String(),
			Logs: []*Log{{
				Data:        []byte{1, 2, 3, 4},
				BlockNumber: 17,
			}},
		},
		Ret: ret,
	}

	enc, err := EncodeTxResponse(data)
	require.NoError(t, err)

	res, err := DecodeTxResponse(enc)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, addr, res.ContractAddress)
	require.Equal(t, bloom.Bytes(), res.Bloom)
	require.Equal(t, data.TxLogs, res.TxLogs)
	require.Equal(t, ret, res.Ret)
}

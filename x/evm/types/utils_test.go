package types

import (
	"testing"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestEvmDataEncoding(t *testing.T) {
	addr := ethcmn.HexToAddress("0x12345")
	bloom := ethtypes.BytesToBloom([]byte{0x1, 0x3})
	ret := []byte{0x5, 0x8}

	encoded := EncodeReturnData(addr, bloom, ret)

	decAddr, decBloom, decRet, err := DecodeReturnData(encoded)

	require.NoError(t, err)
	require.Equal(t, addr, decAddr)
	require.Equal(t, bloom, decBloom)
	require.Equal(t, ret, decRet)
}

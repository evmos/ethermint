package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnmarshalBlockNumberOrHash(t *testing.T) {
	bnh := new(BlockNumberOrHash)
	jsonInput := []byte("{\"blockHash\": \"0x579917054e325746fda5c3ee431d73d26255bc4e10b51163862368629ae19739\"}")
	err := bnh.UnmarshalJSON(jsonInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockHash, common.HexToHash("0x579917054e325746fda5c3ee431d73d26255bc4e10b51163862368629ae19739"))
	require.Nil(t, bnh.BlockNumber)

	bnh = new(BlockNumberOrHash)
	jsonInput = []byte("{\"blockNumber\": \"0x35\"}")
	err = bnh.UnmarshalJSON(jsonInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockNumber, BlockNumber(0x35))
	require.Nil(t, bnh.BlockHash)

	bnh = new(BlockNumberOrHash)
	jsonInput = []byte("{\"blockNumber\": \"latest\"}")
	err = bnh.UnmarshalJSON(jsonInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockNumber, EthLatestBlockNumber)
	require.Nil(t, bnh.BlockHash)

	bnh = new(BlockNumberOrHash)
	jsonInput = []byte("{\"blockHash\": \"0x579917054e325746fda5c3ee431d73d26255bc4e10b51163862368629ae19739\", \"blockNumber\": \"0x35\"}")
	err = bnh.UnmarshalJSON(jsonInput)
	require.Error(t, err)

	bnh = new(BlockNumberOrHash)
	stringInput := []byte("\"0x579917054e325746fda5c3ee431d73d26255bc4e10b51163862368629ae19739\"")
	err = bnh.UnmarshalJSON(stringInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockHash, common.HexToHash("0x579917054e325746fda5c3ee431d73d26255bc4e10b51163862368629ae19739"))
	require.Nil(t, bnh.BlockNumber)

	bnh = new(BlockNumberOrHash)
	stringInput = []byte("\"0x35\"")
	err = bnh.UnmarshalJSON(stringInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockNumber, BlockNumber(0x35))
	require.Nil(t, bnh.BlockHash)

	bnh = new(BlockNumberOrHash)
	stringInput = []byte("\"latest\"")
	err = bnh.UnmarshalJSON(stringInput)
	require.NoError(t, err)
	require.Equal(t, *bnh.BlockNumber, EthLatestBlockNumber)
	require.Nil(t, bnh.BlockHash)

	bnh = new(BlockNumberOrHash)
	stringInput = []byte("\"0xffffffffffffffffffffffffffffffffffffff\"")
	err = bnh.UnmarshalJSON(stringInput)
	require.Error(t, err)
}


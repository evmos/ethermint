package backend // Client defines a mocked object that implements the tendermint rpc CLient

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/evmos/ethermint/rpc/backend/mocks"
	rpc "github.com/evmos/ethermint/rpc/types"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
) // interface. It's used on tests to test the JSON-RPC without running a
// tendermint rpc client server. E.g. JSON-PRC-CLIENT -> BACKEND -> Mock GRPC
// CLIENT -> APP
var _ tmrpcclient.Client = &mocks.Client{}

func TestClient(t *testing.T) {
	client := mocks.NewClient(t)

	// Register the queries and their respective responses, so that they can be
	// called in tests using the client
	height := rpc.BlockNumber(1).Int64()
	RegisterBlockQueries(client, height)

	block := types.Block{Header: types.Header{Height: height}}
	res, err := client.Block(rpc.ContextWithHeight(height), &height)
	require.Equal(t, res, &tmrpctypes.ResultBlock{Block: &block})
	require.NoError(t, err)
}

func RegisterBlockQueries(client *mocks.Client, height int64) {
	block := types.Block{Header: types.Header{Height: height}}
	client.On("Block", rpc.ContextWithHeight(height), mock.AnythingOfType("*int64")).
		Return(&tmrpctypes.ResultBlock{Block: &block}, nil)
}

func RegisterBlockQueriesError(client *mocks.Client, height int64) {
	client.On("Block", rpc.ContextWithHeight(height), mock.AnythingOfType("*int64")).
		Return(nil, sdkerrors.ErrInvalidRequest)
}
func RegisterBlockQueriesNotFound(client *mocks.Client, height int64) {
	client.On("Block", rpc.ContextWithHeight(height), mock.AnythingOfType("*int64")).
		Return(&tmrpctypes.ResultBlock{Block: nil}, nil)
}

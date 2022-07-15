package mock

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/ethermint/rpc/backend"
	"github.com/evmos/ethermint/rpc/types"
	"github.com/stretchr/testify/mock"
)

var _ backend.EVMBackend = MockEVMBackend{}

type MockEVMBackend struct {
	mock.Mock
}

// BlockNumber returns the current block number in abci app state.
// Because abci app state could lag behind from tendermint latest block, it's more stable
// for the client to use the latest block number in abci app state than tendermint rpc.
func (m *MockEVMBackend) BlockNumber() (hexutil.Uint64, error) {
	args := m.Called()
	return hexutil.Uint64(args.Int(0)), args.Error(1)
}

// BloomStatus returns the BloomBitsBlocks and the number of processed sections maintained
// by the chain indexer.
func (m MockEVMBackend) BloomStatus() (uint64, uint64) {
	args := m.Called()
	return uint64(args.Int(0)), uint64(args.Int(1))
}

// GetBlockByNumber returns the block identified by number.
func (m MockEVMBackend) GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	args := m.Called()

	var r0 map[string]interface{}
	// Check for nil args with a type assertion
	if rf, ok := args.Get(0).(func(types.BlockNumber) map[string]interface{}); ok {
		r0 = rf(blockNum)
	} else {
		if args.Get(0) != nil {
			r0 = args.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := args.Get(1).(func(types.BlockNumber) error); ok {
		r1 = rf(blockNum)
	} else {
		r1 = args.Error(1)
	}

	return r0, r1
}

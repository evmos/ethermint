package miner

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx    *server.Context
	logger log.Logger
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
) *API {
	return &API{
		ctx:    ctx,
		logger: ctx.Logger.With("module", "miner"),
	}
}

// SetEtherbase sets the etherbase of the miner
func (api *API) SetEtherbase(etherbase common.Address) bool {
	//api.e.SetEtherbase(etherbase)
	return true
}

// SetGasPrice sets the minimum accepted gas price for the miner.
func (api *API) SetGasPrice(gasPrice hexutil.Big) bool {
	// api.e.lock.Lock()
	// api.e.gasPrice = (*big.Int)(&gasPrice)
	// api.e.lock.Unlock()

	// api.e.txPool.SetGasPrice((*big.Int)(&gasPrice))
	return false
}

package rpc

import (
	"github.com/cosmos/ethermint/ethereum/rpc/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/xlab/suplog"
)

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non-confidential.
// NOTE: For more info about the current status of this endpoints see https://github.com/tharsis/ethermint/issues/124
type PublicTxPoolAPI struct {
	logger log.Logger
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI() *PublicTxPoolAPI {
	return &PublicTxPoolAPI{
		logger: log.WithField("module", "txpool"),
	}
}

// Content returns the transactions contained within the transaction pool
func (api *PublicTxPoolAPI) Content() (map[string]map[string]map[string]*types.RPCTransaction, error) {
	api.logger.Debug("txpool_content")
	content := map[string]map[string]map[string]*types.RPCTransaction{
		"pending": make(map[string]map[string]*types.RPCTransaction),
		"queued":  make(map[string]map[string]*types.RPCTransaction),
	}
	return content, nil
}

// Inspect returns the content of the transaction pool and flattens it into an
func (api *PublicTxPoolAPI) Inspect() (map[string]map[string]map[string]string, error) {
	api.logger.Debug("txpool_inspect")
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	return content, nil
}

// Status returns the number of pending and queued transaction in the pool.
func (api *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	api.logger.Debug("txpool_status")
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(0),
		"queued":  hexutil.Uint(0),
	}
}

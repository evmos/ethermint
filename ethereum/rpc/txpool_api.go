package rpc

import (
	"github.com/cosmos/ethermint/ethereum/rpc/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/xlab/suplog"
)

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non-confidential.
// NOTE: For more info about the current status of this endpoints see https://github.com/tharsis/ethermint/issues/124
type PublicTxPoolAPI struct {
	logger  log.Logger
	backend Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(backend Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{
		logger:  log.WithField("module", "txpool"),
		backend: backend,
	}
}

// Content returns the transactions contained within the transaction pool
func (api *PublicTxPoolAPI) Content() (map[string]map[string]map[string]*types.RPCTransaction, error) {
	api.logger.Debug("txpool_content")
	return api.backend.TxPoolContent()
}

// Inspect returns the content of the transaction pool and flattens it into an
func (api *PublicTxPoolAPI) Inspect() (map[string]map[string]map[string]string, error) {
	api.logger.Debug("txpool_inspect")
	return api.backend.TxPoolInspect()
}

// Status returns the number of pending and queued transaction in the pool.
func (api *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	api.logger.Debug("txpool_status")
	return api.backend.TxPoolStatus()
}

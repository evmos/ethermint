package tracers

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the collection of tracing APIs exposed over the private debugging endpoint.
type API struct {
	ctx    *server.Context
	logger log.Logger
}

// NewAPI creates a new API definition for the tracing methods of the Ethereum service.
func NewAPI(
	ctx *server.Context,
) *API {
	return &API{
		ctx:    ctx,
		logger: ctx.Logger.With("module", "debug"),
	}
}

// TraceConfig holds extra parameters to trace functions.
type TraceConfig struct {
	*vm.LogConfig
	Tracer  *string
	Timeout *string
	Reexec  *uint64
}

// Context contains some contextual infos for a transaction execution that is not
// available from within the EVM object.
type Context struct {
	BlockHash common.Hash // Hash of the block the tx is contained within (zero if dangling tx or call)
	TxIndex   int         // Index of the transaction within a block (zero if dangling tx or call)
	TxHash    common.Hash // Hash of the transaction being traced (zero if dangling call)
}

func (api *API) TraceTransaction(hash common.Hash, config *TraceConfig) (interface{}, error) {
	api.logger.Debug("debug_traceTransaction", "hash", hash)
	//Get transaction by hash

	//Get block by number or hash
	//Find state at the transaction time
	return api.traceTx()
}

// traceTx configures a new tracer according to the provided configuration, and
// executes the given message in the provided environment. The return value will
// be tracer dependent.
func (api *API) traceTx() (interface{}, error) {
	// Assemble the structured logger or the JavaScript tracer
	// If custom javascript tracer is passed set configurations for it
	// Run the transaction with tracing enabled.
	// Depending on the tracer type, format and return the output.
	return nil, nil
}

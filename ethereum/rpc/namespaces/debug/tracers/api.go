package tracers

import (
	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
)

// API is the collection of tracing APIs exposed over the private debugging endpoint.
type API struct {
	ctx         *server.Context
	logger      log.Logger
	backend     backend.Backend
	clientCtx   client.Context
	queryClient *rpctypes.QueryClient
}

// NewAPI creates a new API definition for the tracing methods of the Ethereum service.
func NewAPI(
	ctx *server.Context,
	backend backend.Backend,
	clientCtx client.Context,
) *API {
	return &API{
		ctx:         ctx,
		logger:      ctx.Logger.With("module", "debug"),
		backend:     backend,
		clientCtx:   clientCtx,
		queryClient: rpctypes.NewQueryClient(clientCtx),
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

//
//func (api *API) TraceTransaction(hash common.Hash, _ *TraceConfig) (interface{}, error) {
//	api.logger.Debug("debug_traceTransaction", "hash", hash)
//	//Get transaction by hash
//	transaction, err := api.backend.GetTxByEthHash(hash)
//	if err != nil {
//		api.logger.Debug("tx not found", "hash", hash)
//		return nil, err
//	}
//
//	//check if block number is 0
//	if transaction.Height == 0 {
//		return nil, errors.New("genesis is not traceable")
//	}
//
//	tx, err := api.clientCtx.TxConfig.TxDecoder()(transaction.Tx)
//	if err != nil {
//		api.logger.Debug("tx not found", "hash", hash)
//		return nil, err
//	}
//
//	//TODO Check if there is more than one tx
//
//	ethMessage, ok := tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
//	if !ok {
//		api.logger.Debug("invalid transaction type", "type", fmt.Sprintf("%T", tx))
//		return nil, fmt.Errorf("invalid transaction type %T", tx)
//	}
//
//	//Get block by number or hash
//	//block, err := api.backend.GetBlockByNumber(types.BlockNumber(transaction.Height), false)
//	//if err != nil {
//	//	api.logger.Debug("block number not found", "block", transaction.Height, "hash", hash)
//	//	return nil, err
//	//}
//
//	return api.queryClient.TraceTx(rpctypes.ContextWithHeight(transaction.Height), &evmtypes.QueryTraceTxRequest{
//		Msg:   ethMessage,
//		Index: transaction.Index,
//	})
//}

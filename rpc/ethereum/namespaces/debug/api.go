package debug

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/cosmos/cosmos-sdk/client"
	stderrors "github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/server"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/rpc/ethereum/backend"
	rpctypes "github.com/tharsis/ethermint/rpc/ethereum/types"
)

// HandlerT keeps track of the cpu profiler and trace execution
type HandlerT struct {
	cpuFilename   string
	cpuFile       io.WriteCloser
	mu            sync.Mutex
	traceFilename string
	traceFile     io.WriteCloser
}

// API is the collection of tracing APIs exposed over the private debugging endpoint.
type API struct {
	ctx         *server.Context
	logger      log.Logger
	backend     backend.Backend
	clientCtx   client.Context
	queryClient *rpctypes.QueryClient
	handler     *HandlerT
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
		handler:     new(HandlerT),
	}
}

// TraceTransaction returns the structured logs created during the execution of EVM
// and returns them as a JSON object.
func (a *API) TraceTransaction(hash common.Hash, config *evmtypes.TraceConfig) (interface{}, error) {
	a.logger.Debug("debug_traceTransaction", "hash", hash)
	// Get transaction by hash
	transaction, err := a.backend.GetTxByEthHash(hash)
	if err != nil {
		a.logger.Debug("tx not found", "hash", hash)
		return nil, err
	}

	// check if block number is 0
	if transaction.Height == 0 {
		return nil, errors.New("genesis is not traceable")
	}

	blk, err := a.backend.GetTendermintBlockByNumber(rpctypes.BlockNumber(transaction.Height))
	if err != nil {
		a.logger.Debug("block not found", "height", transaction.Height)
		return nil, err
	}

	msgIndex, _ := rpctypes.FindTxAttributes(transaction.TxResult.Events, hash.Hex())
	if msgIndex < 0 {
		return nil, fmt.Errorf("ethereum tx not found in msgs: %s", hash.Hex())
	}

	// check tx index is not out of bound
	if uint32(len(blk.Block.Txs)) < transaction.Index {
		a.logger.Debug("tx index out of bounds", "index", transaction.Index, "hash", hash.String(), "height", blk.Block.Height)
		return nil, fmt.Errorf("transaction not included in block %v", blk.Block.Height)
	}

	var predecessors []*evmtypes.MsgEthereumTx
	for _, txBz := range blk.Block.Txs[:transaction.Index] {
		tx, err := a.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			a.logger.Debug("failed to decode transaction in block", "height", blk.Block.Height, "error", err.Error())
			continue
		}
		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				continue
			}

			predecessors = append(predecessors, ethMsg)
		}
	}

	tx, err := a.clientCtx.TxConfig.TxDecoder()(transaction.Tx)
	if err != nil {
		a.logger.Debug("tx not found", "hash", hash)
		return nil, err
	}

	// add predecessor messages in current cosmos tx
	for i := 0; i < msgIndex; i++ {
		ethMsg, ok := tx.GetMsgs()[i].(*evmtypes.MsgEthereumTx)
		if !ok {
			continue
		}
		predecessors = append(predecessors, ethMsg)
	}

	ethMessage, ok := tx.GetMsgs()[msgIndex].(*evmtypes.MsgEthereumTx)
	if !ok {
		a.logger.Debug("invalid transaction type", "type", fmt.Sprintf("%T", tx))
		return nil, fmt.Errorf("invalid transaction type %T", tx)
	}

	traceTxRequest := evmtypes.QueryTraceTxRequest{
		Msg:          ethMessage,
		TxIndex:      uint64(transaction.Index),
		Predecessors: predecessors,
		BlockNumber:  blk.Block.Height,
		BlockTime:    blk.Block.Time,
		BlockHash:    common.Bytes2Hex(blk.BlockID.Hash),
	}

	if config != nil {
		traceTxRequest.TraceConfig = config
	}

	// minus one to get the context of block beginning
	contextHeight := transaction.Height - 1
	if contextHeight < 1 {
		// 0 is a special value in `ContextWithHeight`
		contextHeight = 1
	}
	traceResult, err := a.queryClient.TraceTx(rpctypes.ContextWithHeight(contextHeight), &traceTxRequest)
	if err != nil {
		return nil, err
	}

	// Response format is unknown due to custom tracer config param
	// More information can be found here https://geth.ethereum.org/docs/dapp/tracing-filtered
	var decodedResult interface{}
	err = json.Unmarshal(traceResult.Data, &decodedResult)
	if err != nil {
		return nil, err
	}

	return decodedResult, nil
}

// TraceBlockByNumber returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (a *API) TraceBlockByNumber(height rpctypes.BlockNumber, config *evmtypes.TraceConfig) ([]*evmtypes.TxTraceResult, error) {
	a.logger.Debug("debug_traceBlockByNumber", "height", height)
	if height == 0 {
		return nil, errors.New("genesis is not traceable")
	}
	// Get Tendermint Block
	resBlock, err := a.backend.GetTendermintBlockByNumber(height)
	if err != nil {
		a.logger.Debug("get block failed", "height", height, "error", err.Error())
		return nil, err
	}

	return a.traceBlock(height, config, resBlock)
}

// TraceBlockByHash returns the structured logs created during the execution of
// EVM and returns them as a JSON object.
func (a *API) TraceBlockByHash(hash common.Hash, config *evmtypes.TraceConfig) ([]*evmtypes.TxTraceResult, error) {
	a.logger.Debug("debug_traceBlockByHash", "hash", hash)
	// Get Tendermint Block
	resBlock, err := a.backend.GetTendermintBlockByHash(hash)
	if err != nil {
		a.logger.Debug("get block failed", "hash", hash.Hex(), "error", err.Error())
		return nil, err
	}

	if resBlock == nil || resBlock.Block == nil {
		a.logger.Debug("block not found", "hash", hash.Hex())
		return nil, errors.New("block not found")
	}

	return a.traceBlock(rpctypes.BlockNumber(resBlock.Block.Height), config, resBlock)
}

// traceBlock configures a new tracer according to the provided configuration, and
// executes all the transactions contained within. The return value will be one item
// per transaction, dependent on the requested tracer.
func (a *API) traceBlock(height rpctypes.BlockNumber, config *evmtypes.TraceConfig, block *tmrpctypes.ResultBlock) ([]*evmtypes.TxTraceResult, error) {
	txs := block.Block.Txs
	txsLength := len(txs)

	if txsLength == 0 {
		// If there are no transactions return empty array
		return []*evmtypes.TxTraceResult{}, nil
	}

	txDecoder := a.clientCtx.TxConfig.TxDecoder()

	var txsMessages []*evmtypes.MsgEthereumTx
	for i, tx := range txs {
		decodedTx, err := txDecoder(tx)
		if err != nil {
			a.logger.Error("failed to decode transaction", "hash", txs[i].Hash(), "error", err.Error())
			continue
		}

		for _, msg := range decodedTx.GetMsgs() {
			ethMessage, ok := msg.(*evmtypes.MsgEthereumTx)
			if !ok {
				// Just considers Ethereum transactions
				continue
			}
			txsMessages = append(txsMessages, ethMessage)
		}
	}

	// minus one to get the context at the beginning of the block
	contextHeight := height - 1
	if contextHeight < 1 {
		// 0 is a special value for `ContextWithHeight`.
		contextHeight = 1
	}
	ctxWithHeight := rpctypes.ContextWithHeight(int64(contextHeight))

	traceBlockRequest := &evmtypes.QueryTraceBlockRequest{
		Txs:         txsMessages,
		TraceConfig: config,
		BlockNumber: block.Block.Height,
		BlockTime:   block.Block.Time,
		BlockHash:   common.Bytes2Hex(block.BlockID.Hash),
	}

	res, err := a.queryClient.TraceBlock(ctxWithHeight, traceBlockRequest)
	if err != nil {
		return nil, err
	}

	decodedResults := make([]*evmtypes.TxTraceResult, txsLength)
	if err := json.Unmarshal(res.Data, &decodedResults); err != nil {
		return nil, err
	}

	return decodedResults, nil
}

// BlockProfile turns on goroutine profiling for nsec seconds and writes profile data to
// file. It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *API) BlockProfile(file string, nsec uint) error {
	a.logger.Debug("debug_blockProfile", "file", file, "nsec", nsec)
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	time.Sleep(time.Duration(nsec) * time.Second)
	return writeProfile("block", file, a.logger)
}

// CpuProfile turns on CPU profiling for nsec seconds and writes
// profile data to file.
func (a *API) CpuProfile(file string, nsec uint) error { // nolint: golint, stylecheck, revive
	a.logger.Debug("debug_cpuProfile", "file", file, "nsec", nsec)
	if err := a.StartCPUProfile(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	return a.StopCPUProfile()
}

// GcStats returns GC statistics.
func (a *API) GcStats() *debug.GCStats {
	a.logger.Debug("debug_gcStats")
	s := new(debug.GCStats)
	debug.ReadGCStats(s)
	return s
}

// GoTrace turns on tracing for nsec seconds and writes
// trace data to file.
func (a *API) GoTrace(file string, nsec uint) error {
	a.logger.Debug("debug_goTrace", "file", file, "nsec", nsec)
	if err := a.StartGoTrace(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	return a.StopGoTrace()
}

// MemStats returns detailed runtime memory statistics.
func (a *API) MemStats() *runtime.MemStats {
	a.logger.Debug("debug_memStats")
	s := new(runtime.MemStats)
	runtime.ReadMemStats(s)
	return s
}

// SetBlockProfileRate sets the rate of goroutine block profile data collection.
// rate 0 disables block profiling.
func (a *API) SetBlockProfileRate(rate int) {
	a.logger.Debug("debug_setBlockProfileRate", "rate", rate)
	runtime.SetBlockProfileRate(rate)
}

// Stacks returns a printed representation of the stacks of all goroutines.
func (a *API) Stacks() string {
	a.logger.Debug("debug_stacks")
	buf := new(bytes.Buffer)
	err := pprof.Lookup("goroutine").WriteTo(buf, 2)
	if err != nil {
		a.logger.Error("Failed to create stacks", "error", err.Error())
	}
	return buf.String()
}

// StartCPUProfile turns on CPU profiling, writing to the given file.
func (a *API) StartCPUProfile(file string) error {
	a.logger.Debug("debug_startCPUProfile", "file", file)
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	switch {
	case isCPUProfileConfigurationActivated(a.ctx):
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	case a.handler.cpuFile != nil:
		a.logger.Debug("CPU profiling already in progress")
		return errors.New("CPU profiling already in progress")
	default:
		fp, err := ExpandHome(file)
		if err != nil {
			a.logger.Debug("failed to get filepath for the CPU profile file", "error", err.Error())
			return err
		}
		f, err := os.Create(fp)
		if err != nil {
			a.logger.Debug("failed to create CPU profile file", "error", err.Error())
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			a.logger.Debug("cpu profiling already in use", "error", err.Error())
			if err := f.Close(); err != nil {
				a.logger.Debug("failed to close cpu profile file")
				return stderrors.Wrap(err, "failed to close cpu profile file")
			}
			return err
		}

		a.logger.Info("CPU profiling started", "profile", file)
		a.handler.cpuFile = f
		a.handler.cpuFilename = file
		return nil
	}
}

// StopCPUProfile stops an ongoing CPU profile.
func (a *API) StopCPUProfile() error {
	a.logger.Debug("debug_stopCPUProfile")
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	switch {
	case isCPUProfileConfigurationActivated(a.ctx):
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	case a.handler.cpuFile != nil:
		a.logger.Info("Done writing CPU profile", "profile", a.handler.cpuFilename)
		pprof.StopCPUProfile()
		if err := a.handler.cpuFile.Close(); err != nil {
			a.logger.Debug("failed to close cpu file")
			return stderrors.Wrap(err, "failed to close cpu file")
		}
		a.handler.cpuFile = nil
		a.handler.cpuFilename = ""
		return nil
	default:
		a.logger.Debug("CPU profiling not in progress")
		return errors.New("CPU profiling not in progress")
	}
}

// WriteBlockProfile writes a goroutine blocking profile to the given file.
func (a *API) WriteBlockProfile(file string) error {
	a.logger.Debug("debug_writeBlockProfile", "file", file)
	return writeProfile("block", file, a.logger)
}

// WriteMemProfile writes an allocation profile to the given file.
// Note that the profiling rate cannot be set through the API,
// it must be set on the command line.
func (a *API) WriteMemProfile(file string) error {
	a.logger.Debug("debug_writeMemProfile", "file", file)
	return writeProfile("heap", file, a.logger)
}

// MutexProfile turns on mutex profiling for nsec seconds and writes profile data to file.
// It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *API) MutexProfile(file string, nsec uint) error {
	a.logger.Debug("debug_mutexProfile", "file", file, "nsec", nsec)
	runtime.SetMutexProfileFraction(1)
	time.Sleep(time.Duration(nsec) * time.Second)
	defer runtime.SetMutexProfileFraction(0)
	return writeProfile("mutex", file, a.logger)
}

// SetMutexProfileFraction sets the rate of mutex profiling.
func (a *API) SetMutexProfileFraction(rate int) {
	a.logger.Debug("debug_setMutexProfileFraction", "rate", rate)
	runtime.SetMutexProfileFraction(rate)
}

// WriteMutexProfile writes a goroutine blocking profile to the given file.
func (a *API) WriteMutexProfile(file string) error {
	a.logger.Debug("debug_writeMutexProfile", "file", file)
	return writeProfile("mutex", file, a.logger)
}

// FreeOSMemory forces a garbage collection.
func (a *API) FreeOSMemory() {
	a.logger.Debug("debug_freeOSMemory")
	debug.FreeOSMemory()
}

// SetGCPercent sets the garbage collection target percentage. It returns the previous
// setting. A negative value disables GC.
func (a *API) SetGCPercent(v int) int {
	a.logger.Debug("debug_setGCPercent", "percent", v)
	return debug.SetGCPercent(v)
}

// GetHeaderRlp retrieves the RLP encoded for of a single header.
func (a *API) GetHeaderRlp(number uint64) (hexutil.Bytes, error) {
	header, err := a.backend.HeaderByNumber(rpctypes.BlockNumber(number))
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(header)
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (a *API) GetBlockRlp(number uint64) (hexutil.Bytes, error) {
	block, err := a.backend.BlockByNumber(rpctypes.BlockNumber(number))
	if err != nil {
		return nil, err
	}

	return rlp.EncodeToBytes(block)
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (a *API) PrintBlock(number uint64) (string, error) {
	block, err := a.backend.BlockByNumber(rpctypes.BlockNumber(number))
	if err != nil {
		return "", err
	}

	return spew.Sdump(block), nil
}

// SeedHash retrieves the seed hash of a block.
func (a *API) SeedHash(number uint64) (string, error) {
	_, err := a.backend.HeaderByNumber(rpctypes.BlockNumber(number))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", ethash.SeedHash(number)), nil
}

// IntermediateRoots executes a block, and returns a list
// of intermediate roots: the stateroot after each transaction.
func (a *API) IntermediateRoots(hash common.Hash, _ *evmtypes.TraceConfig) ([]common.Hash, error) {
	a.logger.Debug("debug_intermediateRoots", "hash", hash)
	return ([]common.Hash)(nil), nil
}

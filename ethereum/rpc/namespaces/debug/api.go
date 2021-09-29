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

	"github.com/tendermint/tendermint/types"

	evmtypes "github.com/tharsis/ethermint/x/evm/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/ethereum/go-ethereum/common"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
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

	tx, err := a.clientCtx.TxConfig.TxDecoder()(transaction.Tx)
	if err != nil {
		a.logger.Debug("tx not found", "hash", hash)
		return nil, err
	}

	ethMessage, ok := tx.GetMsgs()[0].(*evmtypes.MsgEthereumTx)
	if !ok {
		a.logger.Debug("invalid transaction type", "type", fmt.Sprintf("%T", tx))
		return nil, fmt.Errorf("invalid transaction type %T", tx)
	}

	traceTxRequest := evmtypes.QueryTraceTxRequest{
		Msg:     ethMessage,
		TxIndex: uint64(transaction.Index),
	}

	if config != nil {
		traceTxRequest.TraceConfig = config
	}

	traceResult, err := a.queryClient.TraceTx(rpctypes.ContextWithHeight(transaction.Height), &traceTxRequest)
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
		return nil, err
	}

	return a.traceBlock(height, config, resBlock.Block.Txs)
}

// traceBlock configures a new tracer according to the provided configuration, and
// executes all the transactions contained within. The return value will be one item
// per transaction, dependent on the requested tracer.
func (a API) traceBlock(height rpctypes.BlockNumber, config *evmtypes.TraceConfig, txs types.Txs) ([]*evmtypes.TxTraceResult, error) {
	txsLength := len(txs)

	if txsLength == 0 {
		// If there are no transactions return empty array
		return []*evmtypes.TxTraceResult{}, nil
	}

	var (
		results = make([]*evmtypes.TxTraceResult, txsLength)
		wg      = new(sync.WaitGroup)
		jobs    = make(chan *evmtypes.TxTraceTask, txsLength)
	)

	threads := runtime.NumCPU()
	if threads > txsLength {
		threads = txsLength
	}

	ctxWithHeight := rpctypes.ContextWithHeight(int64(height))

	wg.Add(threads)
	for th := 0; th < threads; th++ {
		go func() {
			defer wg.Done()
			txDecoder := a.clientCtx.TxConfig.TxDecoder()
			// Fetch and execute the next transaction trace tasks
			for task := range jobs {
				tx, err := txDecoder(txs[task.Index])
				if err != nil {
					a.logger.Error("failed to decode transaction", "hash", txs[task.Index].Hash(), "error", err.Error())
					continue
				}

				messages := tx.GetMsgs()
				if len(messages) == 0 {
					continue
				}
				ethMessage, ok := messages[0].(*evmtypes.MsgEthereumTx)
				if !ok {
					// Just considers Ethereum transactions
					continue
				}

				traceTxRequest := &evmtypes.QueryTraceTxRequest{
					Msg:         ethMessage,
					TxIndex:     uint64(task.Index),
					TraceConfig: config,
				}

				res, err := a.queryClient.TraceTx(ctxWithHeight, traceTxRequest)
				if err != nil {
					results[task.Index] = &evmtypes.TxTraceResult{Error: err.Error()}
					continue
				}
				// Response format is unknown due to custom tracer config param
				// More information can be found here https://geth.ethereum.org/docs/dapp/tracing-filtered
				var decodedResult interface{}
				if err := json.Unmarshal(res.Data, &decodedResult); err != nil {
					results[task.Index] = &evmtypes.TxTraceResult{Error: err.Error()}
					continue
				}
				results[task.Index] = &evmtypes.TxTraceResult{Result: decodedResult}
			}
		}()
	}

	for i := range txs {
		jobs <- &evmtypes.TxTraceTask{Index: i}
	}

	close(jobs)
	wg.Wait()

	return results, nil
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
func (a *API) CpuProfile(file string, nsec uint) error { // nolint: golint, stylecheck
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
		f, err := os.Create(ExpandHome(file))
		if err != nil {
			a.logger.Debug("failed to create CPU profile file", "error", err.Error())
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			a.logger.Debug("cpu profiling already in use", "error", err.Error())
			f.Close()
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
		a.handler.cpuFile.Close()
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

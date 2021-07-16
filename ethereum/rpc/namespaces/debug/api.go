package debug

import (
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"
)

// HandlerT keeps track of the cpu profiler and trace execution
type HandlerT struct {
	cpuFilename   string
	cpuFile       io.WriteCloser
	mu            sync.Mutex
	traceFilename string
	traceFile     io.WriteCloser
}

func isCPUProfileConfigurationActivated(ctx *server.Context) bool {
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := ctx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return true
	}
	return false
}

// expands home directory in file paths.
// ~someuser/tmp will not be expanded.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home := os.Getenv("HOME")
		if home == "" {
			if usr, err := user.Current(); err == nil {
				home = usr.HomeDir
			}
		}
		if home != "" {
			p = home + p[1:]
		}
	}
	return filepath.Clean(p)
}

// DebugAPI is the debug_ prefixed set of APIs in the Debug JSON-RPC spec.
type DebugAPI struct {
	ctx     *server.Context
	logger  log.Logger
	handler *HandlerT
}

// NewPublicAPI creates an instance of the Debug API.
func NewDebugAPI(
	ctx *server.Context,
) *DebugAPI {

	return &DebugAPI{
		ctx:     ctx,
		logger:  ctx.Logger.With("module", "debug"),
		handler: new(HandlerT),
	}
}

func (a *DebugAPI) BacktraceAt() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) BlockProfile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) CpuProfile(file string, nsec uint) error {
	if err := a.StartCPUProfile(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	a.StopCPUProfile()
	return nil
}

func (a *DebugAPI) DumpBlock() error {
	return errors.New("Currently not supported.")
}

// GcStats returns GC statistics.
func (a *DebugAPI) GcStats() *debug.GCStats {
	s := new(debug.GCStats)
	debug.ReadGCStats(s)
	return s
}

func (a *DebugAPI) GetBlockRlp() error {
	return errors.New("Currently not supported.")
}

// GoTrace turns on tracing for nsec seconds and writes
// trace data to file.
func (a *DebugAPI) GoTrace(file string, nsec uint) error {
	if err := a.StartGoTrace(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	a.StopGoTrace()
	return nil
}

// MemStats returns detailed runtime memory statistics.
func (a *DebugAPI) MemStats() *runtime.MemStats {
	s := new(runtime.MemStats)
	runtime.ReadMemStats(s)
	return s
}

func (a *DebugAPI) SeedHash() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) SetHead() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) SetBlockProfileRate() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) Stacks() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StartCPUProfile(file string) error {
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.handler.cpuFile != nil {
		return errors.New("CPU profiling already in progress")
	} else {
		f, err := os.Create(expandHome(file))
		if err != nil {
			a.logger.Error("failed to create CPU profile file", "error", err.Error())
			return err
		}

		if err := pprof.StartCPUProfile(f); err != nil {
			a.logger.Error("cpu profiling already in use", "error", err.Error())
			f.Close()
			return err
		}

		a.logger.Info("CPU profiling started", "profile", file)
		a.handler.cpuFile = f
		a.handler.cpuFilename = file
		return nil
	}
}

func (a *DebugAPI) StopCPUProfile() error {
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.handler.cpuFile != nil {
		a.logger.Info("Done writing CPU profile", "profile", a.handler.cpuFilename)
		pprof.StopCPUProfile()
		a.handler.cpuFile.Close()
		a.handler.cpuFile = nil
		a.handler.cpuFilename = ""
		return nil
	} else {
		return errors.New("CPU profiling not in progress")
	}
}

func (a *DebugAPI) TraceBlock() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockByNumber() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockByHash() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockFromFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StandardTraceBlockToFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StandardTraceBadBlockToFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceTransaction(hashHex string) error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceCall() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) Verbosity() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) Vmodule() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) WriteBlockProfile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) WriteMemProfile() error {
	return errors.New("Currently not supported.")
}

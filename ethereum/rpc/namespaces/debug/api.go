package debug

import (
	"bytes"
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

// writeProfile writes the data to a file
func writeProfile(name, file string, log log.Logger) error {
	p := pprof.Lookup(name)
	log.Info("Writing profile records", "count", p.Count(), "type", name, "dump", file)
	f, err := os.Create(expandHome(file))
	if err != nil {
		return err
	}
	defer f.Close()
	return p.WriteTo(f, 0)
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

// BlockProfile turns on goroutine profiling for nsec seconds and writes profile data to
// file. It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *DebugAPI) BlockProfile(file string, nsec uint) error {
	a.logger.Debug("debug_blockProfile", "file", file, "nsec", nsec)

	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	time.Sleep(time.Duration(nsec) * time.Second)
	return writeProfile("block", file, a.logger)
}

// CpuProfile turns on CPU profiling for nsec seconds and writes
// profile data to file.
func (a *DebugAPI) CpuProfile(file string, nsec uint) error {
	a.logger.Debug("debug_cpuProfile", "file", file, "nsec", nsec)
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
	a.logger.Debug("debug_gcStats")
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
	a.logger.Debug("debug_goTrace", "file", file, "nsec", nsec)
	if err := a.StartGoTrace(file); err != nil {
		return err
	}
	time.Sleep(time.Duration(nsec) * time.Second)
	a.StopGoTrace()
	return nil
}

// MemStats returns detailed runtime memory statistics.
func (a *DebugAPI) MemStats() *runtime.MemStats {
	a.logger.Debug("debug_memStats")
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

// SetBlockProfileRate sets the rate of goroutine block profile data collection.
// rate 0 disables block profiling.
func (a *DebugAPI) SetBlockProfileRate(rate int) {
	a.logger.Debug("debug_setBlockProfileRate", "rate", rate)
	runtime.SetBlockProfileRate(rate)
}

// Stacks returns a printed representation of the stacks of all goroutines.
func (a *DebugAPI) Stacks() string {
	a.logger.Debug("debug_stacks")
	buf := new(bytes.Buffer)
	pprof.Lookup("goroutine").WriteTo(buf, 2)
	return buf.String()
}

// StartCPUProfile turns on CPU profiling, writing to the given file.
func (a *DebugAPI) StartCPUProfile(file string) error {
	a.logger.Debug("debug_startCPUProfile", "file", file)
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.handler.cpuFile != nil {
		a.logger.Debug("CPU profiling already in progress")
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

// StopCPUProfile stops an ongoing CPU profile.
func (a *DebugAPI) StopCPUProfile() error {
	a.logger.Debug("debug_stopCPUProfile")
	a.handler.mu.Lock()
	defer a.handler.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		a.logger.Debug("CPU profiling already in progress using the configuration file")
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.handler.cpuFile != nil {
		a.logger.Info("Done writing CPU profile", "profile", a.handler.cpuFilename)
		pprof.StopCPUProfile()
		a.handler.cpuFile.Close()
		a.handler.cpuFile = nil
		a.handler.cpuFilename = ""
		return nil
	} else {
		a.logger.Debug("CPU profiling not in progress")
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

// WriteBlockProfile writes a goroutine blocking profile to the given file.
func (a *DebugAPI) WriteBlockProfile(file string) error {
	a.logger.Debug("debug_writeBlockProfile", "file", file)
	return writeProfile("block", file, a.logger)
}

// WriteMemProfile writes an allocation profile to the given file.
// Note that the profiling rate cannot be set through the API,
// it must be set on the command line.
func (a *DebugAPI) WriteMemProfile(file string) error {
	a.logger.Debug("debug_writeMemProfile", "file", file)
	return writeProfile("heap", file, a.logger)
}

// MutexProfile turns on mutex profiling for nsec seconds and writes profile data to file.
// It uses a profile rate of 1 for most accurate information. If a different rate is
// desired, set the rate and write the profile manually.
func (a *DebugAPI) MutexProfile(file string, nsec uint) error {
	a.logger.Debug("debug_mutexProfile", "file", file, "nsec", nsec)
	runtime.SetMutexProfileFraction(1)
	time.Sleep(time.Duration(nsec) * time.Second)
	defer runtime.SetMutexProfileFraction(0)
	return writeProfile("mutex", file, a.logger)
}

// SetMutexProfileFraction sets the rate of mutex profiling.
func (a *DebugAPI) SetMutexProfileFraction(rate int) {
	a.logger.Debug("debug_setMutexProfileFraction", "rate", rate)
	runtime.SetMutexProfileFraction(rate)
}

// WriteMutexProfile writes a goroutine blocking profile to the given file.
func (a *DebugAPI) WriteMutexProfile(file string) error {
	a.logger.Debug("debug_writeMutexProfile", "file", file)
	return writeProfile("mutex", file, a.logger)
}

// FreeOSMemory forces a garbage collection.
func (a *DebugAPI) FreeOSMemory() {
	a.logger.Debug("debug_freeOSMemory")
	debug.FreeOSMemory()
}

// SetGCPercent sets the garbage collection target percentage. It returns the previous
// setting. A negative value disables GC.
func (a *DebugAPI) SetGCPercent(v int) int {
	a.logger.Debug("debug_setGCPercent", "percent", v)
	return debug.SetGCPercent(v)
}

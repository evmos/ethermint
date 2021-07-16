package debug

import (
	"errors"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"
)

// CPUProfileData keeps track of the cpu profiler execution
type CPUProfileData struct {
	fileName  string
	file      os.File
	activated bool
	mu        sync.Mutex
}

// NewCPUProfileData creates an instance of the CpuProfileData
func NewCPUProfileData() *CPUProfileData {
	return &CPUProfileData{
		activated: false,
	}
}

func isCPUProfileConfigurationActivated(ctx *server.Context) bool {
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := ctx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return true
	}
	return false
}

// DebugAPI is the debug_ prefixed set of APIs in the Debug JSON-RPC spec.
type DebugAPI struct {
	ctx        *server.Context
	logger     log.Logger
	cpuProfile *CPUProfileData
}

// NewPublicAPI creates an instance of the Debug API.
func NewDebugAPI(
	ctx *server.Context,
) *DebugAPI {

	return &DebugAPI{
		ctx:        ctx,
		logger:     ctx.Logger.With("module", "debug"),
		cpuProfile: NewCPUProfileData(),
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

func (a *DebugAPI) GoTrace() error {
	return errors.New("Currently not supported.")
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
	a.cpuProfile.mu.Lock()
	defer a.cpuProfile.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.cpuProfile.activated {
		return errors.New("CPU profiling already in progress")
	} else {
		f, err := os.Create(file)
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
		a.cpuProfile.file = *f
		a.cpuProfile.fileName = file
		a.cpuProfile.activated = true
		return nil
	}
}

func (a *DebugAPI) StartGoTrace() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StopCPUProfile() error {
	a.cpuProfile.mu.Lock()
	defer a.cpuProfile.mu.Unlock()

	if isCPUProfileConfigurationActivated(a.ctx) {
		return errors.New("CPU profiling already in progress using the configuration file")
	} else if a.cpuProfile.activated == true {
		a.logger.Info("Done writing CPU profile", "profile", a.cpuProfile.fileName)
		pprof.StopCPUProfile()
		a.cpuProfile.file.Close()
		a.cpuProfile.activated = false
		a.cpuProfile.fileName = ""
		return nil
	} else {
		return errors.New("CPU profiling not in progress")
	}
}

func (a *DebugAPI) StopGoTrace() error {
	return errors.New("Currently not supported.")
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

package types

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

const (
	TracerAccessList = "access_list"
	TracerJSON       = "json"
	TracerStruct     = "struct"
	TracerMarkdown   = "markdown"
)

// NewTracer creates a new Logger tracer to collect execution traces from an
// EVM transaction.
func NewTracer(tracer string, msg core.Message, cfg *params.ChainConfig, height int64, debug bool) vm.Tracer {
	// TODO: enable additional log configuration
	logCfg := &vm.LogConfig{
		Debug: debug,
	}

	switch tracer {
	case TracerAccessList:
		precompiles := vm.ActivePrecompiles(cfg.Rules(big.NewInt(height)))
		return vm.NewAccessListTracer(msg.AccessList(), msg.From(), *msg.To(), precompiles)
	case TracerJSON:
		return vm.NewJSONLogger(logCfg, os.Stderr)
	case TracerMarkdown:
		return vm.NewMarkdownLogger(logCfg, os.Stdout) // TODO: Stderr ?
	case TracerStruct:
		return vm.NewStructLogger(logCfg)
	default:
		return nil
	}
}

// TxTraceTask represents a single transaction trace task when an entire block
// is being traced.
type TxTraceTask struct {
	Index int // Transaction offset in the block
}

// TxTraceResult is the result of a single transaction trace during a block trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode. Taken from go-ethereum
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   string             `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// FormatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.ErrorString(),
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", stackValue)
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

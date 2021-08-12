package types

import (
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

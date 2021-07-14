package debug

import (
	"github.com/tendermint/tendermint/libs/log"
)

// DebugAPI is the debug_ prefixed set of APIs in the Debug JSON-RPC spec.
type DebugAPI struct {
	logger log.Logger
}

// NewPublicAPI creates an instance of the Web3 API.
func NewDebugAPI(logger log.Logger) *DebugAPI {
	return &DebugAPI{
		logger: logger.With("module", "debug"),
	}
}

// Return hello world as string
// Example call $ curl -X POST --data '{"jsonrpc":"2.0","method":"debug_test","params":[],"id":67}' -H "Content-Type: application/json" http://localhost:8545
func (a *DebugAPI) Test() string {
	a.logger.Debug("Hello world debug")
	return "Hello World"
}

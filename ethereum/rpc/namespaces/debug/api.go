package debug

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/crypto/hd"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"

	sdk "github.com/cosmos/cosmos-sdk/types"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
	ethermint "github.com/tharsis/ethermint/types"
)

// DebugAPI is the debug_ prefixed set of APIs in the Debug JSON-RPC spec.
type DebugAPI struct {
	ctx         context.Context
	queryClient *rpctypes.QueryClient
	backend     backend.Backend
	logger      log.Logger
}

// NewPublicAPI creates an instance of the Web3 API.
func NewDebugAPI(
	logger log.Logger,
	clientCtx client.Context,
	backend backend.Backend,
) *DebugAPI {

	_, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	algos, _ := clientCtx.Keyring.SupportedAlgorithms()

	if !algos.Contains(hd.EthSecp256k1) {
		kr, err := keyring.New(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			clientCtx.KeyringDir,
			clientCtx.Input,
			hd.EthSecp256k1Option(),
		)

		if err != nil {
			panic(err)
		}

		clientCtx = clientCtx.WithKeyring(kr)
	}

	return &DebugAPI{
		ctx:         context.Background(),
		queryClient: rpctypes.NewQueryClient(clientCtx),
		logger:      logger.With("module", "debug"),
		backend:     backend,
	}
}

// Return hello world as string
// Example call $ curl -X POST --data '{"jsonrpc":"2.0","method":"debug_test","params":[],"id":67}' -H "Content-Type: application/json" http://localhost:8545
func (a *DebugAPI) Test() string {
	a.logger.Debug("Hello world debug")
	return "Hello World"
}

func (a *DebugAPI) BacktraceAt() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) BlockProfile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) CpuProfile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) DumpBlock() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) GcStats() string {
	// s := new(debug.GCStats)
	// debug.ReadGCStats(s)
	// return s

	return "TO BE IMPLEMENTED"
}

// func (ec *Client) GCStats(ctx context.Context) (*debug.GCStats, error) {
// 	var result debug.GCStats
// 	err := ec.c.CallContext(ctx, &result, "debug_gcStats")
// 	return &result, err
// }

func (a *DebugAPI) GetBlockRlp() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) GoTrace() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) MemStats() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) SeedHash() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) SetHead() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) SetBlockProfileRate() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) Stacks() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StartCPUProfile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StartGoTrace() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StopCPUProfile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StopGoTrace() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) TraceBlock() string {
	return "TO BE IMPLEMENTED"
}

// We need this for etherscan
func (a *DebugAPI) TraceBlockByNumber() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) TraceBlockByHash() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) TraceBlockFromFile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StandardTraceBlockToFile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StandardTraceBadBlockToFile() string {
	return "TO BE IMPLEMENTED"
}

// We need this for etherscan
func (a *DebugAPI) TraceTransaction() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) TraceCall() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) Verbosity() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) Vmodule() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) WriteBlockProfile() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) WriteMemProfile() string {
	return "TO BE IMPLEMENTED"
}

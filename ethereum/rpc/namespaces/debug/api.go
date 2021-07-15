package debug

import (
	"context"
	"encoding/hex"
	"os"
	"runtime/pprof"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ethereum/go-ethereum/common"
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
	ctx         	context.Context
	clientCtx   	client.Context
	queryClient 	*rpctypes.QueryClient
	backend     	backend.Backend
	logger     		log.Logger
	cpuProfileFile 	os.File
	cpuProfileActivated bool
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
		clientCtx:   clientCtx,
		queryClient: rpctypes.NewQueryClient(clientCtx),
		logger:      logger.With("module", "debug"),
		backend:     backend,
		cpuProfileActivated: false,

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
	// This is blocked because we don't have go-ethereum/metrics/debug.go runing on the background
	return "TO BE IMPLEMENTED"
}

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
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := a.clientCtx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return "Already running using configuration file"
	} else if a.cpuProfileActivated {
		return "Already running using RPC call"
	} else {
		f, err := os.Create("/tmp/cpu-profile")
		if err != nil {
			a.logger.Error("failed to create CP profile", "error", err.Error())
			return "Failed to create cpu profile file."
		}
		a.cpuProfileFile = *f
		a.cpuProfileActivated = true

		a.logger.Info("starting CPU profiler", "profile", cpuProfile)
		if err := pprof.StartCPUProfile(f); err != nil {
			return "Failed to start cpu profile."
		}
		return "Started"
	}
}

func (a *DebugAPI) StartGoTrace() string {
	return "TO BE IMPLEMENTED"
}

func (a *DebugAPI) StopCPUProfile() string {
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := a.clientCtx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return "Already running using configuration file"
	} else if a.cpuProfileActivated == true {
		a.logger.Info("stopping CPU profiler", "profile", cpuProfile)
		pprof.StopCPUProfile()
		a.cpuProfileFile.Close()
		a.cpuProfileActivated = false
		return "Closed"
	} else {
		return "Already Closed"
	}
	
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
// $ curl -X POST --data '{"jsonrpc":"2.0","method":"debug_traceTransaction","params":["735C9268FCFBC944DE61376637FD522CFF447A221D64B744C6BBC72D67420372"],"id":67}' -H "Content-Type: application/json" http://localhost:8545
func (a *DebugAPI) TraceTransaction(hashHex string) (string, error) {

	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return "", err
	}
	node, err := a.clientCtx.GetNode()
	if err != nil {
		return "", err
	}

	tx, err := node.Tx(context.Background(), hash, false)
	if err != nil {
		return "", err
	}

	// Can either cache or just leave this out if not necessary
	block, err := node.Block(context.Background(), &tx.Height)
	if err != nil {
		return "", err
	}

	blockHash := common.BytesToHash(block.Block.Header.Hash())

	ethTx, err := rpctypes.RawTxToEthTx(a.clientCtx, tx.Tx)
	if err != nil {
		return "", err
	}

	// height := uint64(tx.Height)
	// rpcTx := rpctypes.NewTransaction(ethTx.AsTransaction(), blockHash, height, uint64(tx.Index))

	a.logger.Info(blockHash.Hex())

	return blockHash.Hex() + " " + ethTx.AsTransaction().Hash().Hex(), nil

	// return "TO BE IMPLEMENTED"
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

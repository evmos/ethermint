package backend

import (
	"context"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/server/config"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/tendermint/tendermint/libs/log"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// BackendI implements the Cosmos and EVM backend.
type BackendI interface { // nolint: revive
	CosmosBackend
	EVMBackend
}

// CosmosBackend implements the functionality shared within cosmos namespaces
// as defined by Wallet Connect V2: https://docs.walletconnect.com/2.0/json-rpc/cosmos.
// Implemented by Backend.
type CosmosBackend interface { // TODO: define
	// GetAccounts()
	// SignDirect()
	// SignAmino()
}

// EVMBackend implements the functionality shared within ethereum namespaces
// as defined by EIP-1474: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-1474.md
// Implemented by Backend.
type EVMBackend interface {
	// General Ethereum API
	RPCGasCap() uint64            // global gas cap for eth_call over rpc: DoS protection
	RPCEVMTimeout() time.Duration // global timeout for eth_call over rpc: DoS protection
	RPCTxFeeCap() float64         // RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for send-transaction variants. The unit is ether.
	UnprotectedAllowed() bool

	RPCMinGasPrice() int64
	SuggestGasTipCap(baseFee *big.Int) (*big.Int, error)

	// Blockchain API
	BlockNumber() (hexutil.Uint64, error)
	GetTendermintBlockByNumber(blockNum types.BlockNumber) (*tmrpctypes.ResultBlock, error)
	GetTendermintBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error)
	GetBlockByNumber(blockNum types.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	BlockByNumber(blockNum types.BlockNumber) (*ethtypes.Block, error)
	BlockByHash(blockHash common.Hash) (*ethtypes.Block, error)
	CurrentHeader() *ethtypes.Header
	HeaderByNumber(blockNum types.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	PendingTransactions() ([]*sdk.Tx, error)
	GetTransactionCount(address common.Address, blockNum types.BlockNumber) (*hexutil.Uint64, error)
	SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error)
	GetCoinbase() (sdk.AccAddress, error)
	GetTransactionByHash(txHash common.Hash) (*types.RPCTransaction, error)
	GetTxByEthHash(txHash common.Hash) (*tmrpctypes.ResultTx, error)
	GetTxByTxIndex(height int64, txIndex uint) (*tmrpctypes.ResultTx, error)
	EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *types.BlockNumber) (hexutil.Uint64, error)
	BaseFee(height int64) (*big.Int, error)
	GlobalMinGasPrice() (sdk.Dec, error)

	// Fee API
	FeeHistory(blockCount rpc.DecimalOrHex, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (*types.FeeHistoryResult, error)

	// Filter API
	BloomStatus() (uint64, uint64)
	GetLogs(hash common.Hash) ([][]*ethtypes.Log, error)
	GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error)
	ChainConfig() *params.ChainConfig
	SetTxDefaults(args evmtypes.TransactionArgs) (evmtypes.TransactionArgs, error)
	GetEthereumMsgsFromTendermintBlock(block *tmrpctypes.ResultBlock, blockRes *tmrpctypes.ResultBlockResults) []*evmtypes.MsgEthereumTx
}

var _ BackendI = (*Backend)(nil)

// Backend implements the BackendI interface
type Backend struct {
	ctx                 context.Context
	clientCtx           client.Context
	queryClient         *types.QueryClient // gRPC query client
	logger              log.Logger
	chainID             *big.Int
	cfg                 config.Config
	allowUnprotectedTxs bool
}

// NewBackend creates a new Backend instance for cosmos and ethereum namespaces
func NewBackend(ctx *server.Context, logger log.Logger, clientCtx client.Context, allowUnprotectedTxs bool) *Backend {
	chainID, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	appConf := config.GetConfig(ctx.Viper)

	return &Backend{
		ctx:                 context.Background(),
		clientCtx:           clientCtx,
		queryClient:         types.NewQueryClient(clientCtx),
		logger:              logger.With("module", "backend"),
		chainID:             chainID,
		cfg:                 appConf,
		allowUnprotectedTxs: allowUnprotectedTxs,
	}
}

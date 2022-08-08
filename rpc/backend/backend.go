package backend

import (
	"context"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/evmos/ethermint/crypto/hd"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	"github.com/evmos/ethermint/server/config"
	ethermint "github.com/evmos/ethermint/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	tmrpctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// BackendI implements the Cosmos and EVM backend.
type BackendI interface { //nolint: revive
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
	RPCMinGasPrice() int64
	SuggestGasTipCap(baseFee *big.Int) (*big.Int, error)

	// Node specific queries
	Accounts() ([]common.Address, error)
	Syncing() (interface{}, error)
	Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error)
	SendTransaction(args evmtypes.TransactionArgs) (common.Hash, error)
	SignTypedData(address common.Address, typedData apitypes.TypedData) (hexutil.Bytes, error)
	SetEtherbase(etherbase common.Address) bool
	ImportRawKey(privkey, password string) (common.Address, error)
	ListAccounts() ([]common.Address, error)
	NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (*keyring.Record, error)
	UnprotectedAllowed() bool

	// Blocks Info
	BlockNumber() (hexutil.Uint64, error)
	GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error)
	GetTendermintBlockByNumber(blockNum rpctypes.BlockNumber) (*tmrpctypes.ResultBlock, error)
	GetTendermintBlockResultByNumber(height *int64) (*tmrpctypes.ResultBlockResults, error)
	GetTendermintBlockByHash(blockHash common.Hash) (*tmrpctypes.ResultBlock, error)
	GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error)
	BlockByNumber(blockNum rpctypes.BlockNumber) (*ethtypes.Block, error)
	BlockByHash(blockHash common.Hash) (*ethtypes.Block, error)
	GetBlockNumberByHash(blockHash common.Hash) (*big.Int, error)
	GetBlockNumber(blockNrOrHash rpctypes.BlockNumberOrHash) (rpctypes.BlockNumber, error)
	GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint
	GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint
	BlockBloom(blockRes *tmrpctypes.ResultBlockResults) (ethtypes.Bloom, error)

	// Chain Info
	ChainID() (*hexutil.Big, error)
	CurrentHeader() *ethtypes.Header
	HeaderByNumber(blockNum rpctypes.BlockNumber) (*ethtypes.Header, error)
	HeaderByHash(blockHash common.Hash) (*ethtypes.Header, error)
	PendingTransactions() ([]*sdk.Tx, error)
	GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error)
	GetCoinbase() (sdk.AccAddress, error)
	GetTransactionByHash(txHash common.Hash) (*rpctypes.RPCTransaction, error)
	GetTxByEthHash(txHash common.Hash) (*tmrpctypes.ResultTx, error)
	GetTxByTxIndex(height int64, txIndex uint) (*tmrpctypes.ResultTx, error)
	EstimateGas(args evmtypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (hexutil.Uint64, error)
	BaseFee(blockRes *tmrpctypes.ResultBlockResults) (*big.Int, error)
	GlobalMinGasPrice() (sdk.Dec, error)
	GetTransactionByBlockAndIndex(block *tmrpctypes.ResultBlock, idx hexutil.Uint) (*rpctypes.RPCTransaction, error)
	DoCall(args evmtypes.TransactionArgs, blockNr rpctypes.BlockNumber) (*evmtypes.MsgEthereumTxResponse, error)
	GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error)
	GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error)
	GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error)
	GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error)
	GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error)
	GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error)
	GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error)
	ChainConfig() *params.ChainConfig
	GetEthereumMsgsFromTendermintBlock(block *tmrpctypes.ResultBlock, blockRes *tmrpctypes.ResultBlockResults) []*evmtypes.MsgEthereumTx

	// Send Transaction
	Resend(args evmtypes.TransactionArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error)
	SendRawTransaction(data hexutil.Bytes) (common.Hash, error)

	// Fee API
	FeeHistory(blockCount rpc.DecimalOrHex, lastBlock rpc.BlockNumber, rewardPercentiles []float64) (*rpctypes.FeeHistoryResult, error)

	// Filter API
	BloomStatus() (uint64, uint64)
	GetLogs(hash common.Hash) ([][]*ethtypes.Log, error)
	GetLogsByHeight(height *int64) ([][]*ethtypes.Log, error)
	SetTxDefaults(args evmtypes.TransactionArgs) (evmtypes.TransactionArgs, error)

	// Tracing
	TraceTransaction(hash common.Hash, config *evmtypes.TraceConfig) (interface{}, error)
	TraceBlock(height rpctypes.BlockNumber, config *evmtypes.TraceConfig, block *tmrpctypes.ResultBlock) ([]*evmtypes.TxTraceResult, error)
}

var _ BackendI = (*Backend)(nil)

// Backend implements the BackendI interface
type Backend struct {
	ctx                 context.Context
	clientCtx           client.Context
	queryClient         *rpctypes.QueryClient // gRPC query client
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

	algos, _ := clientCtx.Keyring.SupportedAlgorithms()
	if !algos.Contains(hd.EthSecp256k1) {
		kr, err := keyring.New(
			sdk.KeyringServiceName(),
			viper.GetString(flags.FlagKeyringBackend),
			clientCtx.KeyringDir,
			clientCtx.Input,
			clientCtx.Codec,
			hd.EthSecp256k1Option(),
		)
		if err != nil {
			panic(err)
		}

		clientCtx = clientCtx.WithKeyring(kr)
	}

	return &Backend{
		ctx:                 context.Background(),
		clientCtx:           clientCtx,
		queryClient:         rpctypes.NewQueryClient(clientCtx),
		logger:              logger.With("module", "backend"),
		chainID:             chainID,
		cfg:                 appConf,
		allowUnprotectedTxs: allowUnprotectedTxs,
	}
}

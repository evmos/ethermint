package eth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/viper"

	"github.com/cosmos/ethermint/crypto/ethsecp256k1"
	"github.com/cosmos/ethermint/crypto/hd"
	"github.com/cosmos/ethermint/rpc/backend"
	rpctypes "github.com/cosmos/ethermint/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// PublicEthereumAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthereumAPI struct {
	ctx          context.Context
	clientCtx    client.Context
	queryClient  *rpctypes.QueryClient // gRPC query client
	chainIDEpoch *big.Int
	logger       log.Logger
	backend      backend.Backend
	keys         []ethsecp256k1.PrivKey // unlocked keys
	nonceLock    *rpctypes.AddrLocker
	keyringLock  sync.Mutex
}

// NewAPI creates an instance of the public ETH Web3 API.
func NewAPI(
	clientCtx client.Context, backend backend.Backend, nonceLock *rpctypes.AddrLocker,
	keys ...ethsecp256k1.PrivKey,
) *PublicEthereumAPI {

	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	api := &PublicEthereumAPI{
		ctx:          context.Background(),
		clientCtx:    clientCtx,
		queryClient:  rpctypes.NewQueryClient(clientCtx),
		chainIDEpoch: epoch,
		logger:       log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "json-rpc", "namespace", "eth"),
		backend:      backend,
		keys:         keys,
		nonceLock:    nonceLock,
	}

	if err := api.GetKeyringInfo(); err != nil {
		api.logger.Error("failed to get keybase info", "error", err)
	}

	return api
}

// GetKeyringInfo checks if the keyring is present on the client context. If not, it creates a new
// instance and sets it to the client context for later usage.
func (api *PublicEthereumAPI) GetKeyringInfo() error {
	api.keyringLock.Lock()
	defer api.keyringLock.Unlock()

	/*api.clientCtx.keyring won't be nil because we init node with ethermintd keys add CLI
	we need to create a new keyring here with initialized account info merged by using the same api.clientCtx.KeyringDir
	we also need to add eth_secp256k1 key type here*/
	keybase, err := keyring.New(
		sdk.KeyringServiceName(),
		viper.GetString(flags.FlagKeyringBackend),
		api.clientCtx.KeyringDir,
		api.clientCtx.Input,
		hd.EthSecp256k1Option(),
	)
	if err != nil {
		return err
	}

	api.clientCtx.Keyring = keybase
	return nil
}

// ClientCtx returns the Cosmos SDK client context.
func (api *PublicEthereumAPI) ClientCtx() client.Context {
	return api.clientCtx
}

// GetKeys returns the Cosmos SDK client context.
func (api *PublicEthereumAPI) GetKeys() []ethsecp256k1.PrivKey {
	return api.keys
}

// SetKeys sets the given key slice to the set of private keys
func (api *PublicEthereumAPI) SetKeys(keys []ethsecp256k1.PrivKey) {
	api.keys = keys
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (api *PublicEthereumAPI) ProtocolVersion() hexutil.Uint {
	api.logger.Debug("eth_protocolVersion")
	return hexutil.Uint(ethermint.ProtocolVersion)
}

// ChainId returns the chain's identifier in hex format
func (api *PublicEthereumAPI) ChainId() (hexutil.Uint, error) { // nolint
	api.logger.Debug("eth_chainId")
	return hexutil.Uint(uint(api.chainIDEpoch.Uint64())), nil
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (api *PublicEthereumAPI) Syncing() (interface{}, error) {
	api.logger.Debug("eth_syncing")

	status, err := api.clientCtx.Client.Status(api.ctx)
	if err != nil {
		return false, err
	}

	if !status.SyncInfo.CatchingUp {
		return false, nil
	}

	return map[string]interface{}{
		// "startingBlock": nil, // NA
		"currentBlock": hexutil.Uint64(status.SyncInfo.LatestBlockHeight),
		// "highestBlock":  nil, // NA
		// "pulledStates":  nil, // NA
		// "knownStates":   nil, // NA
	}, nil
}

// Coinbase is the address that staking rewards will be send to (alias for Etherbase).
func (api *PublicEthereumAPI) Coinbase() (common.Address, error) {
	api.logger.Debug("eth_coinbase")

	node, err := api.clientCtx.GetNode()
	if err != nil {
		return common.Address{}, err
	}

	status, err := node.Status(api.ctx)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(status.ValidatorInfo.Address.Bytes()), nil
}

// Mining returns whether or not this node is currently mining. Always false.
func (api *PublicEthereumAPI) Mining() bool {
	api.logger.Debug("eth_mining")
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (api *PublicEthereumAPI) Hashrate() hexutil.Uint64 {
	api.logger.Debug("eth_hashrate")
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (api *PublicEthereumAPI) GasPrice() *hexutil.Big {
	api.logger.Debug("eth_gasPrice")
	out := big.NewInt(0)
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (api *PublicEthereumAPI) Accounts() ([]common.Address, error) {
	api.logger.Debug("eth_accounts")
	api.keyringLock.Lock()
	defer api.keyringLock.Unlock()

	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := api.clientCtx.Keyring.List()
	if err != nil {
		return addresses, err
	}

	for _, info := range infos {
		addressBytes := info.GetPubKey().Address().Bytes()
		addresses = append(addresses, common.BytesToAddress(addressBytes))
	}

	return addresses, nil
}

// BlockNumber returns the current block number.
func (api *PublicEthereumAPI) BlockNumber() (hexutil.Uint64, error) {
	api.logger.Debug("eth_blockNumber")
	return api.backend.BlockNumber()
}

// GetBalance returns the provided account's balance up to the provided block number.
//nolint:interfacer
func (api *PublicEthereumAPI) GetBalance(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Big, error) {
	api.logger.Debug("eth_getBalance", "address", address, "block number", blockNum)

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	ctx := api.ctx
	if !(blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber) {
		// wrap the context with the requested height
		ctx = rpctypes.ContextWithHeight(blockNum.Int64())
	}

	res, err := api.queryClient.Balance(ctx, req)
	if err != nil {
		return nil, err
	}

	balance := big.NewInt(0)
	err = balance.UnmarshalText([]byte(res.Balance))
	if err != nil {
		return nil, err
	}

	if blockNum != rpctypes.PendingBlockNumber {
		return (*hexutil.Big)(balance), nil
	}

	// update the address balance with the pending transactions value (if applicable)
	pendingTxs, err := api.backend.PendingTransactions()
	if err != nil {
		return nil, err
	}

	for _, tx := range pendingTxs {
		if tx == nil {
			continue
		}

		if tx.From == address {
			balance = new(big.Int).Sub(balance, tx.Value.ToInt())
		}
		if *tx.To == address {
			balance = new(big.Int).Add(balance, tx.Value.ToInt())
		}
	}

	return (*hexutil.Big)(balance), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
//nolint:interfacer
func (api *PublicEthereumAPI) GetStorageAt(address common.Address, key string, blockNum rpctypes.BlockNumber) (hexutil.Bytes, error) {
	api.logger.Debug("eth_getStorageAt", "address", address, "key", key, "block number", blockNum)

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := api.queryClient.Storage(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (api *PublicEthereumAPI) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	api.logger.Debug("eth_getTransactionCount", "address", address, "block number", blockNum)

	clientCtx := api.clientCtx
	pending := blockNum == rpctypes.PendingBlockNumber

	// pass the given block height to the context if the height is not pending or latest
	if !pending && blockNum != rpctypes.LatestBlockNumber {
		clientCtx = api.clientCtx.WithHeight(blockNum.Int64())
	}

	nonce, err := api.accountNonce(clientCtx, address, pending)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (api *PublicEthereumAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	api.logger.Debug("eth_getBlockTransactionCountByHash", "hash", hash)

	resBlock, err := api.clientCtx.Client.BlockByHash(api.ctx, hash.Bytes())
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by its height.
func (api *PublicEthereumAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	api.logger.Debug("eth_getBlockTransactionCountByNumber", "block number", blockNum)

	var (
		height  int64
		txCount hexutil.Uint
		txs     int
	)

	switch blockNum {
	case rpctypes.PendingBlockNumber:
		// NOTE: pending fetches the latest block
		resBlock, err := api.clientCtx.Client.Block(api.ctx, nil)
		if err != nil {
			return nil
		}

		var resBlockTxLength = 0
		if !resBlock.BlockID.IsZero() {
			resBlockTxLength = len(resBlock.Block.Txs)
		}

		// get the pending transaction count
		pendingTxs, err := api.backend.PendingTransactions()
		if err != nil {
			return nil
		}

		txs = resBlockTxLength + len(pendingTxs)
	case rpctypes.LatestBlockNumber:
		resBlock, err := api.clientCtx.Client.Block(api.ctx, nil)
		if err != nil {
			return nil
		}

		var resBlockTxLength = 0
		if !resBlock.BlockID.IsZero() {
			resBlockTxLength = len(resBlock.Block.Txs)
		}

		txs = resBlockTxLength
	default:
		height = blockNum.Int64()
		resBlock, err := api.clientCtx.Client.Block(api.ctx, &height)
		if err != nil {
			return nil
		}
		txs = len(resBlock.Block.Txs)
	}

	txCount = hexutil.Uint(txs)
	return &txCount
}

// GetUncleCountByBlockHash returns the number of uncles in the block identified by hash. Always zero.
func (api *PublicEthereumAPI) GetUncleCountByBlockHash(_ common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block identified by number. Always zero.
func (api *PublicEthereumAPI) GetUncleCountByBlockNumber(_ rpctypes.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
//nolint:interfacer
func (api *PublicEthereumAPI) GetCode(address common.Address, blockNumber rpctypes.BlockNumber) (hexutil.Bytes, error) {
	api.logger.Debug("eth_getCode", "address", address, "block number", blockNumber)

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	ctx := api.ctx
	if !(blockNumber == rpctypes.PendingBlockNumber || blockNumber == rpctypes.LatestBlockNumber) {
		// wrap the context with the requested height
		ctx = rpctypes.ContextWithHeight(blockNumber.Int64())
	}

	res, err := api.queryClient.Code(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (api *PublicEthereumAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	api.logger.Debug("eth_getTransactionLogs", "hash", txHash)
	return api.backend.GetTransactionLogs(txHash)
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (api *PublicEthereumAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	api.logger.Debug("eth_sign", "address", address, "data", data)
	// TODO: Change this functionality to find an unlocked account by address

	key, exist := rpctypes.GetKeyByAddress(api.keys, address)
	if !exist {
		return nil, keystore.ErrLocked
	}

	// Sign the requested hash with the wallet
	signature, err := key.Sign(data)
	if err != nil {
		return nil, err
	}

	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SendTransaction sends an Ethereum transaction.
func (api *PublicEthereumAPI) SendTransaction(args rpctypes.SendTxArgs) (common.Hash, error) {
	api.logger.Debug("eth_sendTransaction", "args", args)
	// TODO: Change this functionality to find an unlocked account by address

	key, exist := rpctypes.GetKeyByAddress(api.keys, args.From)
	if !exist {
		api.logger.Debug("failed to find key in keyring", "key", args.From)
		return common.Hash{}, keystore.ErrLocked
	}

	// Mutex lock the address' nonce to avoid assigning it to multiple requests
	if args.Nonce == nil {
		api.nonceLock.LockAddr(args.From)
		defer api.nonceLock.UnlockAddr(args.From)
	}

	// Assemble transaction from fields
	tx, err := api.generateFromArgs(args)
	if err != nil {
		api.logger.Debug("failed to generate tx", "error", err)
		return common.Hash{}, err
	}

	if err := tx.ValidateBasic(); err != nil {
		api.logger.Debug("tx failed basic validation", "error", err)
		return common.Hash{}, err
	}

	// Sign transaction
	if err := tx.Sign(api.chainIDEpoch, key.ToECDSA()); err != nil {
		api.logger.Debug("failed to sign tx", "error", err)
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txEncoder := api.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	res, err := api.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return common.Hash{}, err
	}

	if res.Code != abci.CodeTypeOK {
		return common.Hash{}, fmt.Errorf(res.RawLog)
	}
	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (api *PublicEthereumAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	api.logger.Debug("eth_sendRawTransaction", "data", data)
	tx := new(evmtypes.MsgEthereumTx)

	// RLP decode raw transaction bytes
	if err := rlp.DecodeBytes(data, tx); err != nil {
		// Return nil is for when gasLimit overflows uint64
		return common.Hash{}, nil
	}

	// Encode transaction by default Tx encoder
	txBytes, err := api.clientCtx.TxConfig.TxEncoder()(tx)
	if err != nil {
		return common.Hash{}, err
	}

	// TODO: Possibly log the contract creation address (if recipient address is nil) or tx data
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	res, err := api.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return common.Hash{}, err
	}

	if res.Code != abci.CodeTypeOK {
		return common.Hash{}, fmt.Errorf(res.RawLog)
	}
	// Return transaction hash
	return common.HexToHash(res.TxHash), nil
}

// Call performs a raw contract call.
func (api *PublicEthereumAPI) Call(args rpctypes.CallArgs, blockNr rpctypes.BlockNumber, _ *map[common.Address]rpctypes.Account) (hexutil.Bytes, error) {
	api.logger.Debug("eth_call", "args", args, "block number", blockNr)
	simRes, err := api.doCall(args, blockNr, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return []byte{}, err
	}

	data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
	if err != nil {
		return []byte{}, err
	}

	return (hexutil.Bytes)(data.Ret), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (api *PublicEthereumAPI) doCall(
	args rpctypes.CallArgs, blockNum rpctypes.BlockNumber, globalGasCap *big.Int,
) (*sdk.SimulationResponse, error) {

	var height int64
	// pass the given block height to the context if the height is not pending or latest
	if !(blockNum == rpctypes.PendingBlockNumber || blockNum == rpctypes.LatestBlockNumber) {
		height = blockNum.Int64()
	}

	var (
		addr common.Address
		err  error
	)

	// Set sender address or use a default if none specified
	if args.From == nil {
		addrs, err := api.Accounts()
		if err == nil && len(addrs) > 0 {
			addr = addrs[0]
		}
	} else {
		addr = *args.From
	}

	nonce, _ := api.accountNonce(api.clientCtx, addr, true)

	// Set default gas & gas price if none were set
	// TODO: Change this to uint64(math.MaxUint64 / 2) if gas cap can be configured
	gas := uint64(ethermint.DefaultRPCGasLimit)
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != nil && globalGasCap.Uint64() < gas {
		api.logger.Debug("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
		gas = globalGasCap.Uint64()
	}

	// Set gas price using default or parameter if passed in
	gasPrice := new(big.Int).SetUint64(ethermint.DefaultGasPrice)
	if args.GasPrice != nil {
		gasPrice = args.GasPrice.ToInt()
	}

	// Set value for transaction
	value := new(big.Int)
	if args.Value != nil {
		value = args.Value.ToInt()
	}

	// Set Data if provided
	var data []byte
	if args.Data != nil {
		data = []byte(*args.Data)
	}

	var accNum, seq uint64

	// Set destination address for call
	var fromAddr sdk.AccAddress
	if args.From != nil {
		fromAddr = sdk.AccAddress(args.From.Bytes())
		accNum, seq, err = api.clientCtx.AccountRetriever.GetAccountNumberSequence(api.clientCtx, fromAddr)
		if err != nil {
			return nil, err
		}
	}

	var msgs []sdk.Msg
	// Create new call message
	msg := evmtypes.NewMsgEthereumTx(seq, args.To, value, gas, gasPrice, data)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	msgs = append(msgs, msg)

	feeAmount := big.NewInt(0)

	// convert the pending transactions into ethermint msgs
	if blockNum == rpctypes.PendingBlockNumber {
		pendingMsgs, fee, err := api.pendingMsgs()
		if err != nil {
			return nil, err
		}
		feeAmount = new(big.Int).Add(feeAmount, fee)
		msgs = append(msgs, pendingMsgs...)
	}

	privKey, exists := rpctypes.GetKeyByAddress(api.keys, addr)
	if !exists {
		return nil, fmt.Errorf("account with address %s does not exist in keyring", addr.String())
	}

	// NOTE: we query the EVM denomination to allow other chains to use their custom denomination as
	// the fee token
	paramsRes, err := api.queryClient.Params(api.ctx, &evmtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	// create the fee coins with the amount equal to the sum of all msg fees
	fees := sdk.NewCoins(sdk.NewCoin(paramsRes.Params.EvmDenom, sdk.NewIntFromBigInt(feeAmount)))

	finalMsgs := make([]sdk.Msg, 0, len(msgs))
	for _, msg := range msgs {
		m, ok := msg.(*evmtypes.MsgEthereumTx)
		if !ok {
			return nil, fmt.Errorf("invalid message type assertion")
		}
		err = m.Sign(api.chainIDEpoch, privKey.ToECDSA())
		if err != nil {
			return nil, fmt.Errorf("sign msg err: %s", err.Error())
		}
		_, err = m.VerifySig(api.chainIDEpoch)
		if err != nil {
			return nil, fmt.Errorf("verify msg signature err %s", err.Error())
		}
		finalMsgs = append(finalMsgs, m)
	}

	txBytes, err := rpctypes.BuildEthereumTx(api.clientCtx, finalMsgs, accNum, seq, gas, fees, privKey)
	if err != nil {
		return nil, err
	}

	// simulate by calling ABCI Query
	query := abci.RequestQuery{
		Path:   "/app/simulate",
		Data:   txBytes,
		Height: height,
	}

	queryResult, err := api.clientCtx.QueryABCI(query)
	if err != nil {
		return nil, err
	}

	var simResponse sdk.SimulationResponse
	err = jsonpb.Unmarshal(strings.NewReader(string(queryResult.Value)), &simResponse)
	if err != nil {
		return nil, err
	}

	return &simResponse, nil
}

// EstimateGas returns an estimate of gas usage for the given smart contract call.
// It adds 1,000 gas to the returned value instead of using the gas adjustment
// param from the SDK.
func (api *PublicEthereumAPI) EstimateGas(args rpctypes.CallArgs) (hexutil.Uint64, error) {
	api.logger.Debug("eth_estimateGas", "args", args)
	simResponse, err := api.doCall(args, 0, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return 0, err
	}

	// TODO: change 1000 buffer for more accurate buffer (eg: SDK's gasAdjusted)
	estimatedGas := simResponse.GasInfo.GasUsed
	gas := estimatedGas + 1000

	return hexutil.Uint64(gas), nil
}

// GetBlockByHash returns the block identified by hash.
func (api *PublicEthereumAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	api.logger.Debug("eth_getBlockByHash", "hash", hash, "full", fullTx)
	return api.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (api *PublicEthereumAPI) GetBlockByNumber(blockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	api.logger.Debug("eth_getBlockByNumber", "number", blockNum, "full", fullTx)

	if blockNum != rpctypes.PendingBlockNumber {
		return api.backend.GetBlockByNumber(blockNum, fullTx)
	}

	// fetch latest block
	latestBlock, err := api.clientCtx.Client.Block(api.ctx, nil)
	if err != nil {
		return nil, err
	}

	// number of pending txs queried from the mempool
	limit := 1000
	unconfirmedTxs, err := api.clientCtx.Client.UnconfirmedTxs(api.ctx, &limit)
	if err != nil {
		return nil, err
	}

	pendingTxs, gasUsed, err := rpctypes.EthTransactionsFromTendermint(api.clientCtx, unconfirmedTxs.Txs)
	if err != nil {
		return nil, err
	}

	return rpctypes.FormatBlock(
		tmtypes.Header{
			Version:        latestBlock.Block.Version,
			ChainID:        api.clientCtx.ChainID,
			Height:         latestBlock.Block.Height + 1,
			Time:           time.Unix(0, 0),
			LastBlockID:    latestBlock.Block.LastBlockID,
			ValidatorsHash: latestBlock.Block.NextValidatorsHash,
		},
		0,
		latestBlock.Block.Hash(),
		0,
		gasUsed,
		pendingTxs,
		ethtypes.Bloom{},
	), nil

}

// GetTransactionByHash returns the transaction identified by hash.
func (api *PublicEthereumAPI) GetTransactionByHash(hash common.Hash) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByHash", "hash", hash)
	tx, err := api.clientCtx.Client.Tx(api.ctx, hash.Bytes(), false)
	if err != nil {
		// check if the tx is on the mempool
		pendingTxs, pendingErr := api.PendingTransactions()
		if pendingErr != nil {
			return nil, err
		}

		if len(pendingTxs) != 0 {
			for _, tx := range pendingTxs {
				if tx != nil && hash == tx.Hash {
					return tx, nil
				}
			}
		}

		// Return nil for transaction when not found
		return nil, nil
	}

	// Can either cache or just leave this out if not necessary
	block, err := api.clientCtx.Client.Block(api.ctx, &tx.Height)
	if err != nil {
		return nil, err
	}

	blockHash := common.BytesToHash(block.Block.Hash())

	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	height := uint64(tx.Height)
	return rpctypes.NewTransaction(ethTx, common.BytesToHash(tx.Tx.Hash()), blockHash, height, uint64(tx.Index))
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (api *PublicEthereumAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByHashAndIndex", "hash", hash, "index", idx)

	resBlock, err := api.clientCtx.Client.BlockByHash(api.ctx, hash.Bytes())
	if err != nil {
		return nil, err
	}
	return api.getTransactionByBlockAndIndex(resBlock.Block, idx)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (api *PublicEthereumAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	api.logger.Debug("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)
	var (
		height *int64
		err    error
	)

	switch blockNum {
	case rpctypes.PendingBlockNumber:
		// get all the EVM pending txs
		pendingTxs, err := api.backend.PendingTransactions()
		if err != nil {
			return nil, err
		}

		// return if index out of bounds
		if uint64(idx) >= uint64(len(pendingTxs)) {
			return nil, nil
		}

		// change back to pendingTxs[idx] once pending queue is fixed.
		return pendingTxs[int(idx)], nil

	case rpctypes.LatestBlockNumber:
		// nil fetches the latest block
		height = nil

	default:
		h := blockNum.Int64()
		height = &h
	}

	resBlock, err := api.clientCtx.Client.Block(api.ctx, height)
	if err != nil {
		return nil, err
	}

	return api.getTransactionByBlockAndIndex(resBlock.Block, idx)
}

func (api *PublicEthereumAPI) getTransactionByBlockAndIndex(block *tmtypes.Block, idx hexutil.Uint) (*rpctypes.Transaction, error) {
	// return if index out of bounds
	if uint64(idx) >= uint64(len(block.Txs)) {
		return nil, nil
	}

	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, block.Txs[idx])
	if err != nil {
		// return nil error if the transaction is not a MsgEthereumTx
		return nil, nil
	}

	height := uint64(block.Height)
	txHash := common.BytesToHash(block.Txs[idx].Hash())
	blockHash := common.BytesToHash(block.Hash())
	return rpctypes.NewTransaction(ethTx, txHash, blockHash, height, uint64(idx))
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (api *PublicEthereumAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	api.logger.Debug("eth_getTransactionReceipt", "hash", hash)
	tx, err := api.clientCtx.Client.Tx(api.ctx, hash.Bytes(), false)
	if err != nil {
		// Return nil for transaction when not found
		return nil, nil
	}

	// Query block for consensus hash
	block, err := api.clientCtx.Client.Block(api.ctx, &tx.Height)
	if err != nil {
		return nil, err
	}

	blockHash := common.BytesToHash(block.Block.Hash())

	// Convert tx bytes to eth transaction
	ethTx, err := rpctypes.RawTxToEthTx(api.clientCtx, tx.Tx)
	if err != nil {
		return nil, err
	}

	from, err := ethTx.VerifySig(ethTx.ChainID())
	if err != nil {
		return nil, err
	}

	cumulativeGasUsed := uint64(tx.TxResult.GasUsed)
	if tx.Index != 0 {
		cumulativeGasUsed += rpctypes.GetBlockCumulativeGas(api.clientCtx, block.Block, int(tx.Index))
	}

	// Set status codes based on tx result
	var status hexutil.Uint
	if tx.TxResult.IsOK() {
		status = hexutil.Uint(1)
	} else {
		status = hexutil.Uint(0)
	}

	txData := tx.TxResult.GetData()

	data, err := evmtypes.DecodeTxResponse(txData)
	if err != nil {
		status = 0 // transaction failed
	}

	if len(data.TxLogs.Logs) == 0 {
		data.TxLogs.Logs = []*evmtypes.Log{}
	}

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         data.Bloom,
		"logs":              data.TxLogs.EthLogs(),

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": data.ContractAddress,
		"gasUsed":         hexutil.Uint64(tx.TxResult.GasUsed),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        blockHash,
		"blockNumber":      hexutil.Uint64(tx.Height),
		"transactionIndex": hexutil.Uint64(tx.Index),

		// sender and receiver (contract or EOA) addresses
		"from": from,
		"to":   ethTx.To(),
	}

	return receipt, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (api *PublicEthereumAPI) PendingTransactions() ([]*rpctypes.Transaction, error) {
	api.logger.Debug("eth_pendingTransactions")
	return api.backend.PendingTransactions()
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (api *PublicEthereumAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (api *PublicEthereumAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetProof returns an account object with proof and any storage proofs
func (api *PublicEthereumAPI) GetProof(address common.Address, storageKeys []string, block rpctypes.BlockNumber) (*rpctypes.AccountResult, error) {
	height := block.Int64()
	api.logger.Debug("eth_getProof", "address", address, "keys", storageKeys, "number", height)

	ctx := rpctypes.ContextWithHeight(height)
	clientCtx := api.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))
	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := api.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
		if err != nil {
			return nil, err
		}

		// check for proof
		var proofStr string
		if proof != nil {
			proofStr = proof.String()
		}

		storageProofs[i] = rpctypes.StorageResult{
			Key:   key,
			Value: (*hexutil.Big)(new(big.Int).SetBytes(valueBz)),
			Proof: []string{proofStr},
		}
	}

	// query EVM account
	req := &evmtypes.QueryAccountRequest{
		Address: address.String(),
	}

	res, err := api.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := authtypes.AddressStoreKey(sdk.AccAddress(address.Bytes()))
	_, proof, err := api.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	// check for proof
	var accProofStr string
	if proof != nil {
		accProofStr = proof.String()
	}

	balance := big.NewInt(0)
	err = balance.UnmarshalText([]byte(res.Balance))
	if err != nil {
		return nil, err
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: []string{accProofStr},
		Balance:      (*hexutil.Big)(balance),
		CodeHash:     common.BytesToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Ethermint doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// generateFromArgs populates tx message with args (used in RPC API)
func (api *PublicEthereumAPI) generateFromArgs(args rpctypes.SendTxArgs) (*evmtypes.MsgEthereumTx, error) {
	var (
		nonce, gasLimit uint64
		err             error
	)

	amount := (*big.Int)(args.Value)
	gasPrice := (*big.Int)(args.GasPrice)

	if args.GasPrice == nil {
		// Set default gas price
		// TODO: Change to min gas price from context once available through server/daemon
		gasPrice = big.NewInt(ethermint.DefaultGasPrice)
	}

	// get the nonce from the account retriever and the pending transactions
	nonce, err = api.accountNonce(api.clientCtx, args.From, true)
	if err != nil {
		return nil, err
	}
	if args.Nonce != nil {
		if nonce != (uint64)(*args.Nonce) {
			return nil, fmt.Errorf(fmt.Sprintf("invalid nonce; got %d, expected %d", (uint64)(*args.Nonce), nonce))
		}
	}

	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return nil, errors.New("both 'data' and 'input' are set and not equal. Please use 'input' to pass transaction call data")
	}

	// Sets input to either Input or Data, if both are set and not equal error above returns
	var input hexutil.Bytes
	if args.Input != nil {
		input = *args.Input
	} else if args.Data != nil {
		input = *args.Data
	}

	if args.To == nil && len(input) == 0 {
		// Contract creation
		return nil, fmt.Errorf("contract creation without any data provided")
	}

	if args.Gas == nil {
		callArgs := rpctypes.CallArgs{
			From:     &args.From,
			To:       args.To,
			Gas:      args.Gas,
			GasPrice: args.GasPrice,
			Value:    args.Value,
			Data:     &input,
		}
		gl, err := api.EstimateGas(callArgs)
		if err != nil {
			return nil, err
		}
		gasLimit = uint64(gl)
	} else {
		gasLimit = (uint64)(*args.Gas)
	}
	msg := evmtypes.NewMsgEthereumTx(nonce, args.To, amount, gasLimit, gasPrice, input)

	return msg, nil
}

// pendingMsgs constructs an array of sdk.Msg. This method will check pending transactions and convert
// those transactions into ethereum messages. Alonside with the msgs it returns the total fees for all the
// pending txs.
func (api *PublicEthereumAPI) pendingMsgs() ([]sdk.Msg, *big.Int, error) {
	// nolint: prealloc
	var msgs []sdk.Msg
	feeAmount := big.NewInt(0)

	pendingTxs, err := api.PendingTransactions()
	if err != nil {
		return nil, nil, err
	}

	for _, pendingTx := range pendingTxs {
		// NOTE: we have to construct the EVM transaction instead of just casting from the tendermint
		// transactions because PendingTransactions only checks for MsgEthereumTx messages.

		pendingGas, err := hexutil.DecodeUint64(pendingTx.Gas.String())
		if err != nil {
			return nil, nil, err
		}

		pendingValue := pendingTx.Value.ToInt()
		pendingGasPrice := new(big.Int).SetUint64(ethermint.DefaultGasPrice)
		if pendingTx.GasPrice != nil {
			pendingGasPrice = pendingTx.GasPrice.ToInt()
		}

		pendingData := pendingTx.Input
		nonce, _ := api.accountNonce(api.clientCtx, pendingTx.From, true)

		msg := evmtypes.NewMsgEthereumTx(
			0,
			pendingTx.To,
			pendingValue,
			pendingGas,
			pendingGasPrice,
			pendingData,
		)

		feeAmount = new(big.Int).Add(feeAmount, msg.Fee())
		msgs = append(msgs, msg)
	}
	return msgs, feeAmount, nil
}

// accountNonce returns looks up the transaction nonce count for a given address. If the pending boolean
// is set to true, it will add to the counter all the uncommitted EVM transactions sent from the address.
// NOTE: The function returns no error if the account doesn't exist.
func (api *PublicEthereumAPI) accountNonce(
	clientCtx client.Context, address common.Address, pending bool,
) (uint64, error) {
	// Get nonce (sequence) from sender account
	from := sdk.AccAddress(address.Bytes())

	// use a the given client context in case its wrapped with a custom height
	account, err := clientCtx.AccountRetriever.GetAccount(clientCtx, from)
	if err != nil || account == nil {
		// account doesn't exist yet, return 0
		return 0, nil
	}

	nonce := account.GetSequence()

	if !pending {
		return nonce, nil
	}

	// the account retriever doesn't include the uncommitted transactions on the nonce so we need to
	// to manually add them.
	pendingTxs, err := api.backend.PendingTransactions()
	if err != nil {
		return 0, err
	}

	if len(pendingTxs) == 0 {
		return nonce, nil
	}

	// add the uncommitted txs to the nonce counter
	for i := range pendingTxs {
		if pendingTxs[i] == nil {
			continue
		}
		if pendingTxs[i].From == address {
			nonce++
		}
	}

	return nonce, nil
}

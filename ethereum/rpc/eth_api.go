package rpc

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	log "github.com/xlab/suplog"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/ethermint/crypto/hd"
	rpctypes "github.com/cosmos/ethermint/ethereum/rpc/types"
	ethermint "github.com/cosmos/ethermint/types"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"
)

// PublicEthAPI is the eth_ prefixed set of APIs in the Web3 JSON-RPC spec.
type PublicEthAPI struct {
	ctx          context.Context
	clientCtx    client.Context
	queryClient  *rpctypes.QueryClient
	chainIDEpoch *big.Int
	logger       log.Logger
	backend      Backend
	nonceLock    *rpctypes.AddrLocker
}

// NewPublicEthAPI creates an instance of the public ETH Web3 API.
func NewPublicEthAPI(
	clientCtx client.Context,
	backend Backend,
	nonceLock *rpctypes.AddrLocker,
) *PublicEthAPI {
	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
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

	api := &PublicEthAPI{
		ctx:          context.Background(),
		clientCtx:    clientCtx,
		queryClient:  rpctypes.NewQueryClient(clientCtx),
		chainIDEpoch: epoch,
		logger:       log.WithField("module", "json-rpc"),
		backend:      backend,
		nonceLock:    nonceLock,
	}

	return api
}

// ClientCtx returns client context
func (e *PublicEthAPI) ClientCtx() client.Context {
	return e.clientCtx
}

// ProtocolVersion returns the supported Ethereum protocol version.
func (e *PublicEthAPI) ProtocolVersion() hexutil.Uint {
	e.logger.Debugln("eth_protocolVersion")
	return hexutil.Uint(evmtypes.ProtocolVersion)
}

// ChainId returns the chain's identifier in hex format
func (e *PublicEthAPI) ChainId() (hexutil.Uint, error) { // nolint
	e.logger.Debugln("eth_chainId")
	return hexutil.Uint(uint(e.chainIDEpoch.Uint64())), nil
}

// Syncing returns whether or not the current node is syncing with other peers. Returns false if not, or a struct
// outlining the state of the sync if it is.
func (e *PublicEthAPI) Syncing() (interface{}, error) {
	e.logger.Debugln("eth_syncing")

	status, err := e.clientCtx.Client.Status(e.ctx)
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
func (e *PublicEthAPI) Coinbase() (common.Address, error) {
	e.logger.Debugln("eth_coinbase")

	node, err := e.clientCtx.GetNode()
	if err != nil {
		return common.Address{}, err
	}

	status, err := node.Status(e.ctx)
	if err != nil {
		return common.Address{}, err
	}

	return common.BytesToAddress(status.ValidatorInfo.Address.Bytes()), nil
}

// Mining returns whether or not this node is currently mining. Always false.
func (e *PublicEthAPI) Mining() bool {
	e.logger.Debugln("eth_mining")
	return false
}

// Hashrate returns the current node's hashrate. Always 0.
func (e *PublicEthAPI) Hashrate() hexutil.Uint64 {
	e.logger.Debugln("eth_hashrate")
	return 0
}

// GasPrice returns the current gas price based on Ethermint's gas price oracle.
func (e *PublicEthAPI) GasPrice() *hexutil.Big {
	e.logger.Debugln("eth_gasPrice")
	out := big.NewInt(0)
	return (*hexutil.Big)(out)
}

// Accounts returns the list of accounts available to this node.
func (e *PublicEthAPI) Accounts() ([]common.Address, error) {
	e.logger.Debugln("eth_accounts")

	addresses := make([]common.Address, 0) // return [] instead of nil if empty

	infos, err := e.clientCtx.Keyring.List()
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
func (e *PublicEthAPI) BlockNumber() (hexutil.Uint64, error) {
	//e.logger.Debugln("eth_blockNumber")
	return e.backend.BlockNumber()
}

// GetBalance returns the provided account's balance up to the provided block number.
func (e *PublicEthAPI) GetBalance(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Big, error) { // nolint: interfacer
	e.logger.Debugln("eth_getBalance", "address", address.String(), "block number", blockNum)

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Balance(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, err := ethermint.UnmarshalBigInt(res.Balance)
	if err != nil {
		return nil, err
	}

	return (*hexutil.Big)(val), nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (e *PublicEthAPI) GetStorageAt(address common.Address, key string, blockNum rpctypes.BlockNumber) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debugln("eth_getStorageAt", "address", address.Hex(), "key", key, "block number", blockNum)

	req := &evmtypes.QueryStorageRequest{
		Address: address.String(),
		Key:     key,
	}

	res, err := e.queryClient.Storage(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	value := common.HexToHash(res.Value)
	return value.Bytes(), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (e *PublicEthAPI) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	e.logger.Debugln("eth_getTransactionCount", "address", address.Hex(), "block number", blockNum)

	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := e.clientCtx.AccountRetriever

	err := accRet.EnsureExists(e.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		n := hexutil.Uint64(0)
		return &n, nil
	}

	_, nonce, err := accRet.GetAccountNumberSequence(e.clientCtx, from)
	if err != nil {
		return nil, err
	}

	n := hexutil.Uint64(nonce)
	return &n, nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block identified by hash.
func (e *PublicEthAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	e.logger.Debugln("eth_getBlockTransactionCountByHash", "hash", hash.Hex())

	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block identified by number.
func (e *PublicEthAPI) GetBlockTransactionCountByNumber(blockNum rpctypes.BlockNumber) *hexutil.Uint {
	e.logger.Debugln("eth_getBlockTransactionCountByNumber", "block number", blockNum)
	resBlock, err := e.clientCtx.Client.Block(e.ctx, blockNum.TmHeight())
	if err != nil {
		return nil
	}

	n := hexutil.Uint(len(resBlock.Block.Txs))
	return &n
}

// GetUncleCountByBlockHash returns the number of uncles in the block identified by hash. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

// GetUncleCountByBlockNumber returns the number of uncles in the block identified by number. Always zero.
func (e *PublicEthAPI) GetUncleCountByBlockNumber(blockNum rpctypes.BlockNumber) hexutil.Uint {
	return 0
}

// GetCode returns the contract code at the given address and block number.
func (e *PublicEthAPI) GetCode(address common.Address, blockNumber rpctypes.BlockNumber) (hexutil.Bytes, error) { // nolint: interfacer
	e.logger.Debugln("eth_getCode", "address", address.Hex(), "block number", blockNumber)

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := e.queryClient.Code(rpctypes.ContextWithHeight(blockNumber.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetTransactionLogs returns the logs given a transaction hash.
func (e *PublicEthAPI) GetTransactionLogs(txHash common.Hash) ([]*ethtypes.Log, error) {
	e.logger.Debugln("eth_getTransactionLogs", "hash", txHash)
	return e.backend.GetTransactionLogs(txHash)
}

// Sign signs the provided data using the private key of address via Geth's signature standard.
func (e *PublicEthAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	e.logger.Debugln("eth_sign", "address", address.Hex(), "data", common.Bytes2Hex(data))

	from := sdk.AccAddress(address.Bytes())

	_, err := e.clientCtx.Keyring.KeyByAddress(from)
	if err != nil {
		e.logger.Errorln("failed to find key in keyring", "address", address.String())
		return nil, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	// Sign the requested hash with the wallet
	signature, _, err := e.clientCtx.Keyring.SignByAddress(from, data)
	if err != nil {
		e.logger.Panicln("keyring.SignByAddress failed")
		return nil, err
	}

	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// SendTransaction sends an Ethereum transaction.
func (e *PublicEthAPI) SendTransaction(args rpctypes.SendTxArgs) (common.Hash, error) {
	e.logger.Debugln("eth_sendTransaction", "args", args)

	// Look up the wallet containing the requested signer
	_, err := e.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	if err != nil {
		e.logger.WithError(err).Errorln("failed to find key in keyring", "address", args.From)
		return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	}

	args, err = e.setTxDefaults(args)
	if err != nil {
		return common.Hash{}, err
	}

	tx := args.ToTransaction()

	// Assemble transaction from fields
	builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		e.logger.WithError(err).Panicln("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		e.logger.WithError(err).Panicln("codectypes.NewAnyWithValue failed to pack an obvious value")
		return common.Hash{}, err
	}

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(tx.GetMsgs()...)
	if err != nil {
		e.logger.WithError(err).Panicln("builder.SetMsgs failed")
	}

	fees := sdk.NewCoins(ethermint.NewPhotonCoin(sdk.NewIntFromBigInt(tx.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(tx.GetGas())

	if err := tx.ValidateBasic(); err != nil {
		e.logger.Debugln("tx failed basic validation", "error", err)
		return common.Hash{}, err
	}

	// creates a new EIP2929 signer
	// TODO: support legacy txs
	signer := ethtypes.LatestSignerForChainID(args.ChainID.ToInt())

	// Sign transaction
	if err := tx.Sign(signer, e.clientCtx.Keyring); err != nil {
		e.logger.Debugln("failed to sign tx", "error", err)
		return common.Hash{}, err
	}

	// Encode transaction by default Tx encoder
	txEncoder := e.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(tx)
	if err != nil {
		e.logger.WithError(err).Errorln("failed to encode eth tx using default encoder")
		return common.Hash{}, err
	}

	// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
	// https://github.com/tendermint/tendermint/issues/6539
	tmTx := tmtypes.Tx(txBytes)
	txHash := common.BytesToHash(tmTx.Hash())

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	asyncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastAsync)
	if _, err := asyncCtx.BroadcastTx(txBytes); err != nil {
		e.logger.WithError(err).Errorln("failed to broadcast Eth tx")
		return txHash, err
	}

	// Return transaction hash
	return txHash, nil
}

// SendRawTransaction send a raw Ethereum transaction.
func (e *PublicEthAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	e.logger.Debugln("eth_sendRawTransaction", "data_len", len(data))

	// RLP decode raw transaction bytes
	tx, err := e.clientCtx.TxConfig.TxDecoder()(data)
	if err != nil {
		e.logger.WithError(err).Errorln("transaction decoding failed")

		return common.Hash{}, err
	}

	ethereumTx, isEthTx := tx.(*evmtypes.MsgEthereumTx)
	if !isEthTx {
		e.logger.Debugln("invalid transaction type", "type", fmt.Sprintf("%T", tx))
		return common.Hash{}, fmt.Errorf("invalid transaction type %T", tx)
	}

	builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		e.logger.Panicln("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		e.logger.WithError(err).Panicln("codectypes.NewAnyWithValue failed to pack an obvious value")
	}

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(tx.GetMsgs()...)
	if err != nil {
		e.logger.WithError(err).Panicln("builder.SetMsgs failed")
	}

	fees := sdk.NewCoins(ethermint.NewPhotonCoin(sdk.NewIntFromBigInt(ethereumTx.Fee())))
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(ethereumTx.GetGas())

	// Encode transaction by default Tx encoder
	txBytes, err := e.clientCtx.TxConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		e.logger.WithError(err).Errorln("failed to encode eth tx using default encoder")
		return common.Hash{}, err
	}

	// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
	// https://github.com/tendermint/tendermint/issues/6539
	tmTx := tmtypes.Tx(txBytes)
	txHash := common.BytesToHash(tmTx.Hash())

	asyncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastAsync)

	if _, err := asyncCtx.BroadcastTx(txBytes); err != nil {
		e.logger.WithError(err).Errorln("failed to broadcast eth tx")
		return txHash, err
	}

	return txHash, nil
}

// Call performs a raw contract call.
func (e *PublicEthAPI) Call(args rpctypes.CallArgs, blockNr rpctypes.BlockNumber, _ *rpctypes.StateOverride) (hexutil.Bytes, error) {
	e.logger.Debugln("eth_call", "args", args, "block number", blockNr)
	simRes, err := e.doCall(args, blockNr, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return []byte{}, err
	}

	data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
	if err != nil {
		e.logger.WithError(err).Warningln("call result decoding failed")
		return []byte{}, err
	}

	if data.Reverted {
		return []byte{}, rpctypes.ErrRevertedWith(data.Ret)
	}

	return (hexutil.Bytes)(data.Ret), nil
}

// DoCall performs a simulated call operation through the evmtypes. It returns the
// estimated gas used on the operation or an error if fails.
func (e *PublicEthAPI) doCall(
	args rpctypes.CallArgs, blockNr rpctypes.BlockNumber, globalGasCap *big.Int,
) (*sdk.SimulationResponse, error) {
	// Set default gas & gas price if none were set
	// Change this to uint64(math.MaxUint64 / 2) if gas cap can be configured
	gas := uint64(ethermint.DefaultRPCGasLimit)
	if args.Gas != nil {
		gas = uint64(*args.Gas)
	}
	if globalGasCap != nil && globalGasCap.Uint64() < gas {
		e.logger.Debugln("Caller gas above allowance, capping", "requested", gas, "cap", globalGasCap)
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

	var accessList *ethtypes.AccessList
	if args.AccessList != nil {
		accessList = args.AccessList
	}

	// Set destination address for call
	var fromAddr sdk.AccAddress
	if args.From != nil {
		fromAddr = sdk.AccAddress(args.From.Bytes())
	} else {
		fromAddr = sdk.AccAddress(common.Address{}.Bytes())
		args.From = &common.Address{}
	}

	_, seq, err := e.clientCtx.AccountRetriever.GetAccountNumberSequence(e.clientCtx, fromAddr)
	if err != nil {
		return nil, err
	}

	// Create new call message
	msg := evmtypes.NewMsgEthereumTx(e.chainIDEpoch, seq, args.To, value, gas, gasPrice, data, accessList)
	msg.From = args.From.String()

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Create a TxBuilder
	txBuilder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		log.Panicln("clientCtx.TxConfig.NewTxBuilder returns unsupported builder")
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		log.Panicln("codectypes.NewAnyWithValue failed to pack an obvious value")
	}
	txBuilder.SetExtensionOptions(option)

	if err := txBuilder.SetMsgs(msg); err != nil {
		log.Panicln("builder.SetMsgs failed")
	}

	fees := sdk.NewCoins(ethermint.NewPhotonCoin(sdk.NewIntFromBigInt(msg.Fee())))
	txBuilder.SetFeeAmount(fees)
	txBuilder.SetGasLimit(gas)

	// doc about generate and transform tx into json, protobuf bytes
	// https://github.com/cosmos/cosmos-sdk/blob/master/docs/run-node/txs.md
	txBytes, err := e.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	// simulate by calling ABCI Query
	query := abci.RequestQuery{
		Path:   "/app/simulate",
		Data:   txBytes,
		Height: blockNr.Int64(),
	}

	queryResult, err := e.clientCtx.QueryABCI(query)
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
func (e *PublicEthAPI) EstimateGas(args rpctypes.CallArgs) (hexutil.Uint64, error) {
	e.logger.Debugln("eth_estimateGas")

	// From ContextWithHeight: if the provided height is 0,
	// it will return an empty context and the gRPC query will use
	// the latest block height for querying.
	simRes, err := e.doCall(args, 0, big.NewInt(ethermint.DefaultRPCGasLimit))
	if err != nil {
		return 0, err
	}

	data, err := evmtypes.DecodeTxResponse(simRes.Result.Data)
	if err != nil {
		e.logger.WithError(err).Warningln("call result decoding failed")
		return 0, err
	}

	if data.Reverted {
		return 0, rpctypes.ErrRevertedWith(data.Ret)
	}

	// TODO: Add Gas Info from state transition to MsgEthereumTxResponse fields and return that instead
	estimatedGas := simRes.GasInfo.GasUsed
	gas := estimatedGas + 200000

	return hexutil.Uint64(gas), nil
}

// GetBlockByHash returns the block identified by hash.
func (e *PublicEthAPI) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debugln("eth_getBlockByHash", "hash", hash.Hex(), "full", fullTx)
	return e.backend.GetBlockByHash(hash, fullTx)
}

// GetBlockByNumber returns the block identified by number.
func (e *PublicEthAPI) GetBlockByNumber(ethBlockNum rpctypes.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	e.logger.Debugln("eth_getBlockByNumber", "number", ethBlockNum, "full", fullTx)
	return e.backend.GetBlockByNumber(ethBlockNum, fullTx)
}

// GetTransactionByHash returns the transaction identified by hash.
func (e *PublicEthAPI) GetTransactionByHash(hash common.Hash) (*rpctypes.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByHash", "hash", hash.Hex())

	res, err := e.clientCtx.Client.Tx(e.ctx, hash.Bytes(), false)
	if err != nil {
		e.logger.WithError(err).Debugln("tx not found", "hash", hash.Hex())
		return nil, nil
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &res.Height)
	if err != nil {
		e.logger.WithError(err).Debugln("block not found", "height", res.Height)
		return nil, nil
	}

	tx, err := e.clientCtx.TxConfig.TxDecoder()(res.Tx)
	if err != nil {
		e.logger.WithError(err).Debugln("decoding failed")
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		e.logger.Debugln("invalid tx")
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}

	// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
	// https://github.com/tendermint/tendermint/issues/6539

	return rpctypes.NewTransactionFromData(
		msg.Data,
		common.HexToAddress(msg.From),
		hash,
		common.BytesToHash(resBlock.Block.Hash()),
		uint64(res.Height),
		uint64(res.Index),
	)
}

// GetTransactionByBlockHashAndIndex returns the transaction identified by hash and index.
func (e *PublicEthAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByBlockHashAndIndex", "hash", hash.Hex(), "index", idx)

	resBlock, err := e.clientCtx.Client.BlockByHash(e.ctx, hash.Bytes())
	if err != nil {
		e.logger.WithError(err).Debugln("block not found", "hash", hash.Hex())
		return nil, nil
	}

	i := int(idx)
	if i >= len(resBlock.Block.Txs) {
		e.logger.Debugln("block txs index out of bound", "index", i)
		return nil, nil
	}

	txBz := resBlock.Block.Txs[i]
	tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
	if err != nil {
		e.logger.WithError(err).Debugln("decoding failed")
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		e.logger.Debugln("invalid tx")
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}

	// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
	// https://github.com/tendermint/tendermint/issues/6539
	txHash := common.BytesToHash(txBz.Hash())

	return rpctypes.NewTransactionFromData(
		msg.Data,
		common.HexToAddress(msg.From),
		txHash,
		hash,
		uint64(resBlock.Block.Height),
		uint64(idx),
	)
}

// GetTransactionByBlockNumberAndIndex returns the transaction identified by number and index.
func (e *PublicEthAPI) GetTransactionByBlockNumberAndIndex(blockNum rpctypes.BlockNumber, idx hexutil.Uint) (*rpctypes.RPCTransaction, error) {
	e.logger.Debugln("eth_getTransactionByBlockNumberAndIndex", "number", blockNum, "index", idx)

	resBlock, err := e.clientCtx.Client.Block(e.ctx, blockNum.TmHeight())
	if err != nil {
		e.logger.WithError(err).Debugln("block not found", "height", blockNum.Int64())
		return nil, nil
	}

	i := int(idx)
	if i >= len(resBlock.Block.Txs) {
		e.logger.Debugln("block txs index out of bound", "index", i)
		return nil, nil
	}

	txBz := resBlock.Block.Txs[i]
	tx, err := e.clientCtx.TxConfig.TxDecoder()(txBz)
	if err != nil {
		e.logger.WithError(err).Debugln("decoding failed")
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		e.logger.Debugln("invalid tx")
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}

	// TODO: use msg.AsTransaction.Hash() for txHash once hashing is fixed on Tendermint
	// https://github.com/tendermint/tendermint/issues/6539
	txHash := common.BytesToHash(txBz.Hash())

	return rpctypes.NewTransactionFromData(
		msg.Data,
		common.HexToAddress(msg.From),
		txHash,
		common.BytesToHash(resBlock.Block.Hash()),
		uint64(resBlock.Block.Height),
		uint64(idx),
	)
}

// GetTransactionReceipt returns the transaction receipt identified by hash.
func (e *PublicEthAPI) GetTransactionReceipt(hash common.Hash) (map[string]interface{}, error) {
	e.logger.Debugln("eth_getTransactionReceipt", "hash", hash.Hex())

	res, err := e.clientCtx.Client.Tx(e.ctx, hash.Bytes(), false)
	if err != nil {
		e.logger.WithError(err).Debugln("tx not found", "hash", hash.Hex())
		return nil, nil
	}

	resBlock, err := e.clientCtx.Client.Block(e.ctx, &res.Height)
	if err != nil {
		e.logger.WithError(err).Debugln("block not found", "height", res.Height)
		return nil, nil
	}

	tx, err := e.clientCtx.TxConfig.TxDecoder()(res.Tx)
	if err != nil {
		e.logger.WithError(err).Debugln("decoding failed")
		return nil, fmt.Errorf("failed to decode tx: %w", err)
	}

	msg, ok := tx.(*evmtypes.MsgEthereumTx)
	if !ok {
		e.logger.Debugln("invalid tx")
		return nil, fmt.Errorf("invalid tx type: %T", tx)
	}

	cumulativeGasUsed := uint64(0)
	blockRes, err := e.clientCtx.Client.BlockResults(e.ctx, &res.Height)
	if err != nil {
		e.logger.WithError(err).Debugln("failed to retrieve block results", "height", res.Height)
		return nil, nil
	}

	for i := 0; i <= int(res.Index) && i < len(blockRes.TxsResults); i++ {
		cumulativeGasUsed += uint64(blockRes.TxsResults[i].GasUsed)
	}

	var status hexutil.Uint

	// NOTE: Response{Deliver/Check}Tx Code from Tendermint and the Ethereum receipt status are switched:
	// |         | Tendermint | Ethereum |
	// | ------- | ---------- | -------- |
	// | Success | 0          | 1        |
	// | Fail    | 1          | 0        |

	if res.TxResult.Code == 0 {
		status = hexutil.Uint(ethtypes.ReceiptStatusSuccessful)
	} else {
		status = hexutil.Uint(ethtypes.ReceiptStatusFailed)
	}

	from := common.HexToAddress(msg.From)

	resLogs, err := e.queryClient.TxLogs(e.ctx, &evmtypes.QueryTxLogsRequest{Hash: hash.Hex()})
	if err != nil {
		e.logger.WithError(err).Debugln("logs not found", "hash", hash.Hex())
		resLogs = &evmtypes.QueryTxLogsResponse{Logs: []*evmtypes.Log{}}
	}

	logs := evmtypes.LogsToEthereum(resLogs.Logs)

	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            status,
		"cumulativeGasUsed": hexutil.Uint64(cumulativeGasUsed),
		"logsBloom":         ethtypes.BytesToBloom(ethtypes.LogsBloom(logs)),
		"logs":              logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": hash,
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(res.TxResult.GasUsed),
		"type":            hexutil.Uint(ethtypes.AccessListTxType), // TODO: support legacy type

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        common.BytesToHash(resBlock.Block.Header.Hash()).Hex(),
		"blockNumber":      hexutil.Uint64(res.Height),
		"transactionIndex": hexutil.Uint64(res.Index),

		// sender and receiver (contract or EOA) addreses
		"from": from,
		"to":   msg.To(),
	}

	if logs == nil {
		receipt["logs"] = [][]*ethtypes.Log{}
	}

	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if msg.To() == nil {
		receipt["contractAddress"] = crypto.CreateAddress(from, msg.Data.Nonce)
	}

	return receipt, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (e *PublicEthAPI) PendingTransactions() ([]*rpctypes.RPCTransaction, error) {
	e.logger.Debugln("eth_getPendingTransactions")
	return e.backend.PendingTransactions()
}

// GetUncleByBlockHashAndIndex returns the uncle identified by hash and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetUncleByBlockNumberAndIndex returns the uncle identified by number and index. Always returns nil.
func (e *PublicEthAPI) GetUncleByBlockNumberAndIndex(number hexutil.Uint, idx hexutil.Uint) map[string]interface{} {
	return nil
}

// GetProof returns an account object with proof and any storage proofs
func (e *PublicEthAPI) GetProof(address common.Address, storageKeys []string, blockNumber rpctypes.BlockNumber) (*rpctypes.AccountResult, error) {
	height := blockNumber.Int64()
	e.logger.Debugln("eth_getProof", "address", address.Hex(), "keys", storageKeys, "number", height)

	ctx := rpctypes.ContextWithHeight(height)
	clientCtx := e.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))
	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := e.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
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

	res, err := e.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := authtypes.AddressStoreKey(sdk.AccAddress(address.Bytes()))
	_, proof, err := e.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	// check for proof
	var accProofStr string
	if proof != nil {
		accProofStr = proof.String()
	}

	balance, err := ethermint.UnmarshalBigInt(res.Balance)
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

// setTxDefaults populates tx message with default values in case they are not
// provided on the args
func (e *PublicEthAPI) setTxDefaults(args rpctypes.SendTxArgs) (rpctypes.SendTxArgs, error) {

	// Get nonce (sequence) from sender account
	from := sdk.AccAddress(args.From.Bytes())

	if args.GasPrice == nil {
		// TODO: Change to either:
		// - min gas price from context once available through server/daemon, or
		// - suggest a gas price based on the previous included txs
		args.GasPrice = (*hexutil.Big)(big.NewInt(ethermint.DefaultGasPrice))
	}

	if args.Nonce != nil {
		// get the nonce from the account retriever
		// ignore error in case tge account doesn't exist yet
		_, nonce, _ := e.clientCtx.AccountRetriever.GetAccountNumberSequence(e.clientCtx, from)
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}

	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return args, errors.New("both 'data' and 'input' are set and not equal. Please use 'input' to pass transaction call data")
	}

	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}

		if len(input) == 0 {
			return args, errors.New(`contract creation without any data provided`)
		}
	}

	if args.Gas == nil {
		// For backwards-compatibility reason, we try both input and data
		// but input is preferred.
		input := args.Input
		if input == nil {
			input = args.Data
		}

		callArgs := rpctypes.CallArgs{
			From:       &args.From, // From shouldn't be nil
			To:         args.To,
			Gas:        args.Gas,
			GasPrice:   args.GasPrice,
			Value:      args.Value,
			Data:       input,
			AccessList: args.AccessList,
		}
		estimated, err := e.EstimateGas(callArgs)
		if err != nil {
			return args, err
		}
		args.Gas = &estimated
		e.logger.Debugln("estimate gas usage automatically", "gas", args.Gas)
	}

	if args.ChainID == nil {
		args.ChainID = (*hexutil.Big)(e.chainIDEpoch)
	}

	return args, nil
}

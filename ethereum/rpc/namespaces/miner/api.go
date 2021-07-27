package miner

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
	evmtypes "github.com/tharsis/ethermint/x/evm/types"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx    *server.Context
	logger log.Logger
	ethAPI *eth.PublicAPI
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
	ethAPI *eth.PublicAPI,
) *API {
	return &API{
		ctx:    ctx,
		ethAPI: ethAPI,
		logger: ctx.Logger.With("module", "miner"),
	}
}

// SetEtherbase sets the etherbase of the miner
func (api *API) SetEtherbase(etherbase common.Address) bool {
	api.logger.Debug("miner_setEtherbase")

	list, err := api.ethAPI.ClientCtx().Keyring.List()
	if err != nil && len(list) > 0 {
		api.logger.Debug("Could not get list of addresses")
		return false
	}

	addr := common.BytesToAddress(list[0].GetPubKey().Address())

	api.logger.Info(addr.String())

	args := rpctypes.SendTxArgs{}
	args.From = addr
	// TODO: set this as the message info
	// delegatorAddress := addr
	// withdrawAddress := etherbase
	args.Data = &hexutil.Bytes{}
	args, err = api.ethAPI.SetTxDefaults(args)
	if err != nil {
		api.logger.Debug("Unable to parse transaction args", "error", err.Error())
		return false
	}

	msg := args.ToTransaction()

	if err := msg.ValidateBasic(); err != nil {
		api.logger.Debug("tx failed basic validation", "error", err.Error())
		return false
	}

	signer := ethtypes.LatestSignerForChainID(args.ChainID.ToInt())

	// Sign transaction
	if err := msg.Sign(signer, api.ethAPI.ClientCtx().Keyring); err != nil {
		api.logger.Debug("failed to sign tx", "error", err.Error())
		return false
	}

	// Assemble transaction from fields
	builder, ok := api.ethAPI.ClientCtx().TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		api.logger.Error("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
	}

	option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	if err != nil {
		api.logger.Error("codectypes.NewAnyWithValue failed to pack an obvious value", "error", err.Error())
		return false
	}

	builder.SetExtensionOptions(option)
	err = builder.SetMsgs(msg)
	if err != nil {
		api.logger.Error("builder.SetMsgs failed", "error", err.Error())
	}

	// Query params to use the EVM denomination
	res, err := api.ethAPI.QueryClient().QueryClient.Params(api.ethAPI.Ctx(), &evmtypes.QueryParamsRequest{})
	if err != nil {
		api.logger.Error("failed to query evm params", "error", err.Error())
		return false
	}

	txData, err := evmtypes.UnpackTxData(msg.Data)
	if err != nil {
		api.logger.Error("failed to unpack tx data", "error", err.Error())
		return false
	}

	fees := sdk.Coins{sdk.NewCoin(res.Params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(msg.GetGas())

	// Encode transaction by default Tx encoder
	txEncoder := api.ethAPI.ClientCtx().TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		api.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return false
	}

	txHash := msg.AsTransaction().Hash()

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := api.ethAPI.ClientCtx().WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if err != nil || rsp.Code != 0 {
		if err == nil {
			err = errors.New(rsp.RawLog)
		}
		api.logger.Error("failed to broadcast tx", "error", err.Error())
		return false
	}

	api.logger.Info("Broadcasting tx...", "hash", txHash)
	return true
}

// SetGasPrice sets the minimum accepted gas price for the miner.
func (api *API) SetGasPrice(gasPrice hexutil.Big) bool {
	api.logger.Info(api.ctx.Viper.ConfigFileUsed())
	appConf, err := config.ParseConfig(api.ctx.Viper)
	if err != nil {
		api.logger.Error("failed to parse file.", "file", api.ctx.Viper.ConfigFileUsed(), "error:", err.Error())
		return false
	}
	// NOTE: To allow values less that 1 aphoton, we need to divide the gasPrice here using some constant
	// If we want to work the same as go-eth we should just use the gasPrice as an int without converting it
	coinsValue := gasPrice.ToInt().String()
	unit := "aphoton"
	c, err := sdk.ParseDecCoins(coinsValue + unit)
	if err != nil {
		api.logger.Error("failed to parse coins", "coins", coinsValue, "error", err.Error())
		return false
	}
	appConf.SetMinGasPrices(c)
	config.WriteConfigFile(api.ctx.Viper.ConfigFileUsed(), appConf)
	api.logger.Info("Your configuration file was modified. Please RESTART your node.", "value", coinsValue+unit)
	return true
}

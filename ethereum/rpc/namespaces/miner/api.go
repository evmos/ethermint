package miner

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth"
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
	// api.logger.Debug("miner_setEtherbase")

	// // Look up the wallet containing the requested signer
	// _, err := api.clientCtx.Keyring.KeyByAddress(sdk.AccAddress(args.From.Bytes()))
	// if err != nil {
	// 	e.logger.Error("failed to find key in keyring", "address", args.From, "error", err.Error())
	// 	return common.Hash{}, fmt.Errorf("%s; %s", keystore.ErrNoMatch, err.Error())
	// }

	// args, err = e.setTxDefaults(args)
	// if err != nil {
	// 	return common.Hash{}, err
	// }

	// msg := args.ToTransaction()

	// if err := msg.ValidateBasic(); err != nil {
	// 	e.logger.Debug("tx failed basic validation", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// // TODO: get from chain config
	// signer := ethtypes.LatestSignerForChainID(args.ChainID.ToInt())

	// // Sign transaction
	// if err := msg.Sign(signer, e.clientCtx.Keyring); err != nil {
	// 	e.logger.Debug("failed to sign tx", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// // Assemble transaction from fields
	// builder, ok := e.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	// if !ok {
	// 	e.logger.Error("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
	// }

	// option, err := codectypes.NewAnyWithValue(&evmtypes.ExtensionOptionsEthereumTx{})
	// if err != nil {
	// 	e.logger.Error("codectypes.NewAnyWithValue failed to pack an obvious value", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// builder.SetExtensionOptions(option)
	// err = builder.SetMsgs(msg)
	// if err != nil {
	// 	e.logger.Error("builder.SetMsgs failed", "error", err.Error())
	// }

	// // Query params to use the EVM denomination
	// res, err := e.queryClient.QueryClient.Params(e.ctx, &evmtypes.QueryParamsRequest{})
	// if err != nil {
	// 	e.logger.Error("failed to query evm params", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// txData, err := evmtypes.UnpackTxData(msg.Data)
	// if err != nil {
	// 	e.logger.Error("failed to unpack tx data", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// fees := sdk.Coins{sdk.NewCoin(res.Params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}
	// builder.SetFeeAmount(fees)
	// builder.SetGasLimit(msg.GetGas())

	// // Encode transaction by default Tx encoder
	// txEncoder := e.clientCtx.TxConfig.TxEncoder()
	// txBytes, err := txEncoder(builder.GetTx())
	// if err != nil {
	// 	e.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
	// 	return common.Hash{}, err
	// }

	// txHash := msg.AsTransaction().Hash()

	// // Broadcast transaction in sync mode (default)
	// // NOTE: If error is encountered on the node, the broadcast will not return an error
	// syncCtx := e.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	// rsp, err := syncCtx.BroadcastTx(txBytes)
	// if err != nil || rsp.Code != 0 {
	// 	if err == nil {
	// 		err = errors.New(rsp.RawLog)
	// 	}
	// 	e.logger.Error("failed to broadcast tx", "error", err.Error())
	// 	return txHash, err
	// }

	// // Return transaction hash
	// return txHash, nil
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

package miner

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	"github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx     *server.Context
	logger  log.Logger
	ethAPI  *eth.PublicAPI
	backend backend.Backend
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
	ethAPI *eth.PublicAPI,
	backend backend.Backend,
) *API {
	return &API{
		ctx:     ctx,
		ethAPI:  ethAPI,
		logger:  ctx.Logger.With("module", "miner"),
		backend: backend,
	}
}

// SetEtherbase sets the etherbase of the miner
func (api *API) SetEtherbase(etherbase common.Address) bool {
	api.logger.Debug("miner_setEtherbase")

	addr, err := api.backend.GetCoinbase()
	if err != nil {
		api.logger.Debug("failed to get address")
		return false
	}

	api.logger.Info("Etherbase account ", addr.String())

	withdrawAddr := sdk.AccAddress(etherbase.Bytes())
	msg := distributiontypes.NewMsgSetWithdrawAddress(addr, withdrawAddr)

	if err := msg.ValidateBasic(); err != nil {
		api.logger.Debug("tx failed basic validation", "error", err.Error())
		return false
	}

	// Assemble transaction from fields
	builder, ok := api.ethAPI.ClientCtx().TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		api.logger.Error("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
	}

	err = builder.SetMsgs(msg)
	if err != nil {
		api.logger.Error("builder.SetMsgs failed", "error", err.Error())
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(api.ethAPI.ClientCtx().ChainID).
		WithKeybase(api.ethAPI.ClientCtx().Keyring).
		WithTxConfig(api.ethAPI.ClientCtx().TxConfig)

	keyInfo, err := api.ethAPI.ClientCtx().Keyring.KeyByAddress(addr)
	if err != nil {
		return false
	}

	if err := tx.Sign(txFactory, keyInfo.GetName(), builder, false); err != nil {
		return false
	}

	// fees := sdk.Coins{sdk.NewCoin(res.Params.EvmDenom, sdk.NewIntFromBigInt(txData.Fee()))}
	// builder.SetFeeAmount(fees)
	// builder.SetGasLimit(msg.GetGas())

	// Encode transaction by default Tx encoder
	txEncoder := api.ethAPI.ClientCtx().TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		api.logger.Error("failed to encode eth tx using default encoder", "error", err.Error())
		return false
	}

	tmHash := common.BytesToHash(tmtypes.Tx(txBytes).Hash())

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

	api.logger.Info("Broadcasting tx...", "hash", tmHash)
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

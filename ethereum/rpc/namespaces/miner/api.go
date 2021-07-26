package miner

import (
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx    *server.Context
	logger log.Logger
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
) *API {
	return &API{
		ctx:    ctx,
		logger: ctx.Logger.With("module", "miner"),
	}
}

// SetEtherbase sets the etherbase of the miner
func (api *API) SetEtherbase(etherbase common.Address) bool {
	//api.e.SetEtherbase(etherbase)
	return true
}

// SetGasPrice sets the minimum accepted gas price for the miner.
func (api *API) SetGasPrice(gasPrice hexutil.Big) bool {
	api.logger.Info(api.ctx.Viper.ConfigFileUsed())
	appConf, err := config.ParseConfig(api.ctx.Viper)
	if err != nil {
		// TODO: fix this error format
		api.logger.Error("failed to parse %s: %w", api.ctx.Viper.ConfigFileUsed(), err)
		return false
	}
	// TODO: should this value be wei?
	coinsValue := gasPrice.ToInt().String()
	unit := "aphoton"
	c, err := sdk.ParseDecCoins(coinsValue + unit)
	if err != nil {
		// TODO: fix this error format
		api.logger.Error("failed to parse coins %s, %s: %w", coinsValue, unit, err)
		return false
	}
	appConf.SetMinGasPrices(c)
	config.WriteConfigFile(api.ctx.Viper.ConfigFileUsed(), appConf)
	api.logger.Info("Your configuration file was modified. Please RESTART your node.", "value", coinsValue+unit)
	return true
}

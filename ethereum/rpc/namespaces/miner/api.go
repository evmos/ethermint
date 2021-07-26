package miner

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tendermint/tendermint/libs/log"
	ethapi "github.com/tharsis/ethermint/ethereum/rpc/namespaces/eth"
	ethermint "github.com/tharsis/ethermint/types"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx          *server.Context
	logger       log.Logger
	chainIDEpoch *big.Int
	clientCtx    client.Context
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
	clientCtx client.Context,
) *API {
	epoch, err := ethermint.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	return &API{
		ctx:          ctx,
		clientCtx:    *ethapi.AddKeyringToClientCtx(clientCtx),
		chainIDEpoch: epoch,
		logger:       ctx.Logger.With("module", "miner"),
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

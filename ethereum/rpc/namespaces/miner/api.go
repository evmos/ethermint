package miner

import (
	"errors"
	"math/big"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	sdkconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/tharsis/ethermint/ethereum/rpc/backend"
	rpctypes "github.com/tharsis/ethermint/ethereum/rpc/types"
	"github.com/tharsis/ethermint/server/config"
)

// API is the miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx       *server.Context
	logger    log.Logger
	clientCtx client.Context
	backend   backend.Backend
}

// NewMinerAPI creates an instance of the Miner API.
func NewMinerAPI(
	ctx *server.Context,
	clientCtx client.Context,
	backend backend.Backend,
) *API {
	return &API{
		ctx:       ctx,
		clientCtx: clientCtx,
		logger:    ctx.Logger.With("api", "miner"),
		backend:   backend,
	}
}

// SetEtherbase sets the etherbase of the miner
func (api *API) SetEtherbase(etherbase common.Address) bool {
	api.logger.Debug("miner_setEtherbase")

	delAddr, err := api.backend.GetCoinbase()
	if err != nil {
		api.logger.Debug("failed to get coinbase address", "error", err.Error())
		return false
	}

	withdrawAddr := sdk.AccAddress(etherbase.Bytes())
	msg := distributiontypes.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)

	if err := msg.ValidateBasic(); err != nil {
		api.logger.Debug("tx failed basic validation", "error", err.Error())
		return false
	}

	// Assemble transaction from fields
	builder, ok := api.clientCtx.TxConfig.NewTxBuilder().(authtx.ExtensionOptionsTxBuilder)
	if !ok {
		api.logger.Debug("clientCtx.TxConfig.NewTxBuilder returns unsupported builder", "error", err.Error())
		return false
	}

	err = builder.SetMsgs(msg)
	if err != nil {
		api.logger.Error("builder.SetMsgs failed", "error", err.Error())
		return false
	}

	// Fetch minimun gas price to calculate fees using the configuration.
	appConf := config.GetConfig(api.ctx.Viper)

	minGasPrices := appConf.GetMinGasPrices()
	if len(minGasPrices) == 0 || minGasPrices.Empty() {
		api.logger.Debug("the minimun fee is not set")
		return false
	}
	minGasPriceValue := minGasPrices[0].Amount
	denom := minGasPrices[0].Denom

	delCommonAddr := common.BytesToAddress(delAddr.Bytes())
	nonce, err := api.backend.GetTransactionCount(delCommonAddr, rpctypes.EthPendingBlockNumber)
	if err != nil {
		api.logger.Debug("failed to get nonce", "error", err.Error())
		return false
	}

	txFactory := tx.Factory{}
	txFactory = txFactory.
		WithChainID(api.clientCtx.ChainID).
		WithKeybase(api.clientCtx.Keyring).
		WithTxConfig(api.clientCtx.TxConfig).
		WithSequence(uint64(*nonce)).
		WithGasAdjustment(1.25)

	_, gas, err := tx.CalculateGas(api.clientCtx, txFactory, msg)
	if err != nil {
		api.logger.Debug("failed to calculate gas", "error", err.Error())
		return false
	}

	txFactory = txFactory.WithGas(gas)

	value := new(big.Int).SetUint64(gas * minGasPriceValue.Ceil().TruncateInt().Uint64())
	fees := sdk.Coins{sdk.NewCoin(denom, sdk.NewIntFromBigInt(value))}
	builder.SetFeeAmount(fees)
	builder.SetGasLimit(gas)

	keyInfo, err := api.clientCtx.Keyring.KeyByAddress(delAddr)
	if err != nil {
		api.logger.Debug("failed to get the wallet address using the keyring", "error", err.Error())
		return false
	}

	if err := tx.Sign(txFactory, keyInfo.GetName(), builder, false); err != nil {
		api.logger.Debug("failed to sign tx", "error", err.Error())
		return false
	}

	// Encode transaction by default Tx encoder
	txEncoder := api.clientCtx.TxConfig.TxEncoder()
	txBytes, err := txEncoder(builder.GetTx())
	if err != nil {
		api.logger.Debug("failed to encode eth tx using default encoder", "error", err.Error())
		return false
	}

	tmHash := common.BytesToHash(tmtypes.Tx(txBytes).Hash())

	// Broadcast transaction in sync mode (default)
	// NOTE: If error is encountered on the node, the broadcast will not return an error
	syncCtx := api.clientCtx.WithBroadcastMode(flags.BroadcastSync)
	rsp, err := syncCtx.BroadcastTx(txBytes)
	if err != nil || rsp.Code != 0 {
		if err == nil {
			err = errors.New(rsp.RawLog)
		}
		api.logger.Debug("failed to broadcast tx", "error", err.Error())
		return false
	}

	api.logger.Debug("broadcasted tx to set miner withdraw address (etherbase)", "hash", tmHash.String())
	return true
}

// SetGasPrice sets the minimum accepted gas price for the miner.
// NOTE: this function accepts only integers to have the same interface than go-eth
// to use float values, the gas prices must be configured using the configuration file
func (api *API) SetGasPrice(gasPrice hexutil.Big) bool {
	api.logger.Info(api.ctx.Viper.ConfigFileUsed())
	appConf := config.GetConfig(api.ctx.Viper)

	var unit string
	minGasPrices := appConf.GetMinGasPrices()

	// fetch the base denom from the sdk Config in case it's not currently defined on the node config
	if len(minGasPrices) == 0 || minGasPrices.Empty() {
		var err error
		unit, err = sdk.GetBaseDenom()
		if err != nil {
			api.logger.Debug("could not get the denom of smallest unit registered", "error", err.Error())
			return false
		}
	} else {
		unit = minGasPrices[0].Denom
	}

	c := sdk.NewDecCoin(unit, sdk.NewIntFromBigInt(gasPrice.ToInt()))

	appConf.SetMinGasPrices(sdk.DecCoins{c})
	sdkconfig.WriteConfigFile(api.ctx.Viper.ConfigFileUsed(), appConf)
	api.logger.Info("Your configuration file was modified. Please RESTART your node.", "gas-price", c.String())
	return true
}

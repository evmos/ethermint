package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/cosmos/ethermint/handlers"
	"github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"
	ethparams "github.com/ethereum/go-ethereum/params"

	tmcmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

const (
	appName = "Ethermint"
)

type (
	// EthermintApp implements an extended ABCI application. It is an application
	// that may process transactions through Ethereum's EVM running atop of
	// Tendermint consensus.
	EthermintApp struct {
		*bam.BaseApp

		codec  *wire.Codec
		sealed bool

		accountKey    *sdk.KVStoreKey
		accountMapper auth.AccountMapper
		// TODO: keys, stores, mappers, and keepers
	}

	// Options is a function signature that provides the ability to modify
	// options of an EthermintApp during initialization.
	Options func(*EthermintApp)
)

// NewEthermintApp returns a reference to a new initialized Ethermint
// application.
func NewEthermintApp(
	logger tmlog.Logger, db dbm.DB, ethChainCfg *ethparams.ChainConfig,
	sdkAddr ethcmn.Address, opts ...Options,
) *EthermintApp {

	codec := CreateCodec()
	app := &EthermintApp{
		BaseApp:    bam.NewBaseApp(appName, codec, logger, db),
		codec:      codec,
		accountKey: sdk.NewKVStoreKey("accounts"),
	}
	app.accountMapper = auth.NewAccountMapper(codec, app.accountKey, auth.ProtoBaseAccount)

	app.SetTxDecoder(types.TxDecoder(codec, sdkAddr))
	app.SetAnteHandler(handlers.AnteHandler(app.accountMapper))
	app.MountStoresIAVL(app.accountKey)

	for _, opt := range opts {
		opt(app)
	}

	err := app.LoadLatestVersion(app.accountKey)
	if err != nil {
		tmcmn.Exit(err.Error())
	}

	app.seal()
	return app
}

// seal seals the Ethermint application and prohibits any future modifications
// that change critical components.
func (app *EthermintApp) seal() {
	app.sealed = true
}

// CreateCodec creates a new amino wire codec and registers all the necessary
// structures and interfaces needed for the application.
func CreateCodec() *wire.Codec {
	codec := wire.NewCodec()

	// Register other modules, types, and messages...
	types.RegisterWire(codec)
	return codec
}

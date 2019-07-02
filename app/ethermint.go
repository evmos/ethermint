package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/cosmos/ethermint/crypto"
	evmtypes "github.com/cosmos/ethermint/x/evm/types"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

const appName = "Ethermint"

// application multi-store keys
var (
	storeKeyAccount     = sdk.NewKVStoreKey("acc")
	storeKeyStorage     = sdk.NewKVStoreKey("contract_storage")
	storeKeyMain        = sdk.NewKVStoreKey("main")
	storeKeyStake       = sdk.NewKVStoreKey("stake")
	storeKeySlashing    = sdk.NewKVStoreKey("slashing")
	storeKeyGov         = sdk.NewKVStoreKey("gov")
	storeKeyFeeColl     = sdk.NewKVStoreKey("fee")
	storeKeyParams      = sdk.NewKVStoreKey("params")
	storeKeyTransParams = sdk.NewTransientStoreKey("transient_params")
)

type (
	// EthermintApp implements an extended ABCI application. It is an application
	// that may process transactions through Ethereum's EVM running atop of
	// Tendermint consensus.
	EthermintApp struct {
		*bam.BaseApp

		cdc *codec.Codec

		accountKey  *sdk.KVStoreKey
		storageKey  *sdk.KVStoreKey
		mainKey     *sdk.KVStoreKey
		stakeKey    *sdk.KVStoreKey
		slashingKey *sdk.KVStoreKey
		govKey      *sdk.KVStoreKey
		feeCollKey  *sdk.KVStoreKey
		paramsKey   *sdk.KVStoreKey
		tParamsKey  *sdk.TransientStoreKey

		accountKeeper auth.AccountKeeper
		feeCollKeeper auth.FeeCollectionKeeper
		// coinKeeper     bank.Keeper
		stakeKeeper    stake.Keeper
		slashingKeeper slashing.Keeper
		govKeeper      gov.Keeper
		paramsKeeper   params.Keeper
	}
)

// NewEthermintApp returns a reference to a new initialized Ethermint
// application.
//
// TODO: Ethermint needs to support being bootstrapped as an application running
// in a sovereign zone and as an application running with a shared security model.
// For now, it will support only running as a sovereign application.
func NewEthermintApp(logger tmlog.Logger, db dbm.DB, baseAppOpts ...func(*bam.BaseApp)) *EthermintApp {
	cdc := CreateCodec()

	baseApp := bam.NewBaseApp(appName, logger, db, evmtypes.TxDecoder(cdc), baseAppOpts...)
	app := &EthermintApp{
		BaseApp:     baseApp,
		cdc:         cdc,
		accountKey:  storeKeyAccount,
		storageKey:  storeKeyStorage,
		mainKey:     storeKeyMain,
		stakeKey:    storeKeyStake,
		slashingKey: storeKeySlashing,
		govKey:      storeKeyGov,
		feeCollKey:  storeKeyFeeColl,
		paramsKey:   storeKeyParams,
		tParamsKey:  storeKeyTransParams,
	}

	app.paramsKeeper = params.NewKeeper(app.cdc, app.paramsKey, app.tParamsKey)
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.accountKey, auth.ProtoBaseAccount)
	app.feeCollKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.feeCollKey)

	// register message handlers
	app.Router().
		// TODO: add remaining routes
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper)).
		AddRoute("gov", gov.NewHandler(app.govKeeper))

	// initialize the underlying ABCI BaseApp
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(NewAnteHandler(app.accountKeeper, app.feeCollKeeper))

	app.MountStores(
		app.mainKey, app.accountKey, app.stakeKey, app.slashingKey,
		app.govKey, app.feeCollKey, app.paramsKey, app.storageKey,
	)
	app.MountStore(app.tParamsKey, sdk.StoreTypeTransient)

	if err := app.LoadLatestVersion(app.accountKey); err != nil {
		tmcmn.Exit(err.Error())
	}

	app.BaseApp.Seal()
	return app
}

// BeginBlocker signals the beginning of a block. It performs application
// updates on the start of every block.
func (app *EthermintApp) BeginBlocker(
	_ sdk.Context, _ abci.RequestBeginBlock,
) abci.ResponseBeginBlock {

	return abci.ResponseBeginBlock{}
}

// EndBlocker signals the end of a block. It performs application updates on
// the end of every block.
func (app *EthermintApp) EndBlocker(
	_ sdk.Context, _ abci.RequestEndBlock,
) abci.ResponseEndBlock {

	return abci.ResponseEndBlock{}
}

// initChainer initializes the application blockchain with validators and other
// state data from TendermintCore.
func (app *EthermintApp) initChainer(
	_ sdk.Context, req abci.RequestInitChain,
) abci.ResponseInitChain {

	var genesisState GenesisState
	stateJSON := req.AppStateBytes

	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse application genesis state"))
	}

	// TODO: load the genesis accounts

	return abci.ResponseInitChain{}
}

// CreateCodec creates a new amino wire codec and registers all the necessary
// concrete types and interfaces needed for the application.
func CreateCodec() *codec.Codec {
	cdc := codec.New()

	// TODO: Add remaining codec registrations:
	// bank, staking, distribution, slashing, and gov

	crypto.RegisterCodec(cdc)
	evmtypes.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

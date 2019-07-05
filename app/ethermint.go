package app

import (
	"os"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"

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
	// default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.emintcli")

	storeKeyAccount     = sdk.NewKVStoreKey("acc")
	storeKeyStorage     = sdk.NewKVStoreKey("contract_storage")
	storeKeyMain        = sdk.NewKVStoreKey("main")
	storeKeyStake       = sdk.NewKVStoreKey("stake")
	storeKeySlashing    = sdk.NewKVStoreKey("slashing")
	storeKeyGov         = sdk.NewKVStoreKey("gov")
	storeKeySupply      = sdk.NewKVStoreKey("supply")
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
		supplyKey   *sdk.KVStoreKey
		paramsKey   *sdk.KVStoreKey
		tParamsKey  *sdk.TransientStoreKey

		accountKeeper  auth.AccountKeeper
		supplyKeeper   supply.Keeper
		bankKeeper     bank.Keeper
		stakeKeeper    staking.Keeper
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
		supplyKey:   storeKeySupply,
		paramsKey:   storeKeyParams,
		tParamsKey:  storeKeyTransParams,
	}

	// Set params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(app.cdc, app.paramsKey, app.tParamsKey, params.DefaultCodespace)
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)
	bankSubspace := app.paramsKeeper.Subspace(bank.DefaultParamspace)

	// account permissions
	basicModuleAccs := []string{auth.FeeCollectorName, distr.ModuleName}
	minterModuleAccs := []string{mint.ModuleName}
	burnerModuleAccs := []string{staking.BondedPoolName, staking.NotBondedPoolName, gov.ModuleName}

	// Add keepers
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.accountKey, authSubspace, auth.ProtoBaseAccount)
	app.bankKeeper = bank.NewBaseKeeper(app.accountKeeper, bankSubspace, bank.DefaultCodespace)
	app.supplyKeeper = supply.NewKeeper(cdc, app.supplyKey, app.accountKeeper, app.bankKeeper, supply.DefaultCodespace, basicModuleAccs, minterModuleAccs, burnerModuleAccs)

	// register message handlers
	app.Router().
		// TODO: add remaining routes
		AddRoute("stake", staking.NewHandler(app.stakeKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper)).
		AddRoute("gov", gov.NewHandler(app.govKeeper))

	// initialize the underlying ABCI BaseApp
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(NewAnteHandler(app.accountKeeper, app.supplyKeeper))

	app.MountStores(
		app.mainKey, app.accountKey, app.stakeKey, app.slashingKey,
		app.govKey, app.supplyKey, app.paramsKey, app.storageKey,
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

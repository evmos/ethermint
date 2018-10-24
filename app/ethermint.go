package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/pkg/errors"

	"github.com/cosmos/ethermint/handlers"
	"github.com/cosmos/ethermint/types"

	ethcmn "github.com/ethereum/go-ethereum/common"

	abci "github.com/tendermint/tendermint/abci/types"
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

		accountKeeper  auth.AccountKeeper
		feeCollKeeper  auth.FeeCollectionKeeper
		coinKeeper     bank.Keeper
		stakeKeeper    stake.Keeper
		slashingKeeper slashing.Keeper
		govKeeper      gov.Keeper
		paramsKeeper   params.Keeper
	}
)

// NewEthermintApp returns a reference to a new initialized Ethermint
// application.
func NewEthermintApp(logger tmlog.Logger, db dbm.DB, sdkAddr ethcmn.Address) *EthermintApp {
	cdc := CreateCodec()
	cms := store.NewCommitMultiStore(db)

	baseAppOpts := []func(*bam.BaseApp){
		func(bApp *bam.BaseApp) { bApp.SetCMS(cms) },
	}
	baseApp := bam.NewBaseApp(appName, logger, db, types.TxDecoder(cdc, sdkAddr), baseAppOpts...)

	app := &EthermintApp{
		BaseApp:     baseApp,
		cdc:         cdc,
		accountKey:  types.StoreKeyAccount,
		storageKey:  types.StoreKeyStorage,
		mainKey:     types.StoreKeyMain,
		stakeKey:    types.StoreKeyStake,
		slashingKey: types.StoreKeySlashing,
		govKey:      types.StoreKeyGov,
		feeCollKey:  types.StoreKeyFeeColl,
		paramsKey:   types.StoreKeyParams,
		tParamsKey:  types.StoreKeyTransParams,
	}

	// set application keepers and mappers
	app.accountKeeper = auth.NewAccountKeeper(app.cdc, app.accountKey, auth.ProtoBaseAccount)
	app.paramsKeeper = params.NewKeeper(app.cdc, app.paramsKey, app.tParamsKey)
	app.feeCollKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.feeCollKey)

	// register message handlers
	app.Router().
		// TODO: Do we need to mount bank and IBC handlers? Should be handled
		// directly in the EVM.
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper)).
		AddRoute("gov", gov.NewHandler(app.govKeeper))

	// initialize the underlying ABCI BaseApp
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(handlers.AnteHandler(app.accountKeeper, app.feeCollKeeper))

	app.MountStoresIAVL(
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
func (app *EthermintApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return abci.ResponseBeginBlock{}
}

// EndBlocker signals the end of a block. It performs application updates on
// the end of every block.
func (app *EthermintApp) EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{}
}

// initChainer initializes the application blockchain with validators and other
// state data from TendermintCore.
func (app *EthermintApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
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

	types.RegisterCodec(cdc)
	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

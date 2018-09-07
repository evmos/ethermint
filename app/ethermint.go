package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/pkg/errors"

	"github.com/cosmos/ethermint/handlers"
	"github.com/cosmos/ethermint/state"
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

		codec *wire.Codec
		ethDB *state.Database

		accountKey  *sdk.KVStoreKey
		storageKey  *sdk.KVStoreKey
		mainKey     *sdk.KVStoreKey
		stakeKey    *sdk.KVStoreKey
		slashingKey *sdk.KVStoreKey
		govKey      *sdk.KVStoreKey
		feeCollKey  *sdk.KVStoreKey
		paramsKey   *sdk.KVStoreKey
		tParamsKey  *sdk.TransientStoreKey

		accountMapper  auth.AccountMapper
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
	codec := CreateCodec()
	cms := store.NewCommitMultiStore(db)

	baseAppOpts := []func(*bam.BaseApp){
		func(bApp *bam.BaseApp) { bApp.SetCMS(cms) },
	}
	baseApp := bam.NewBaseApp(appName, logger, db, types.TxDecoder(codec, sdkAddr), baseAppOpts...)

	app := &EthermintApp{
		BaseApp:     baseApp,
		codec:       codec,
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

	// create Ethereum state database
	ethDB, err := state.NewDatabase(cms, db, state.DefaultStoreCacheSize)
	if err != nil {
		tmcmn.Exit(err.Error())
	}

	app.ethDB = ethDB

	// set application keepers and mappers
	app.accountMapper = auth.NewAccountMapper(codec, app.accountKey, auth.ProtoBaseAccount)
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.paramsKeeper = params.NewKeeper(app.codec, app.paramsKey)
	app.feeCollKeeper = auth.NewFeeCollectionKeeper(app.codec, app.feeCollKey)
	app.stakeKeeper = stake.NewKeeper(
		app.codec, app.stakeKey, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace),
	)
	app.govKeeper = gov.NewKeeper(
		app.codec, app.govKey, app.paramsKeeper.Setter(), app.coinKeeper,
		app.stakeKeeper, app.RegisterCodespace(gov.DefaultCodespace),
	)
	app.slashingKeeper = slashing.NewKeeper(
		app.codec, app.slashingKey, app.stakeKeeper,
		app.paramsKeeper.Getter(), app.RegisterCodespace(slashing.DefaultCodespace),
	)

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
	app.SetAnteHandler(handlers.AnteHandler(app.accountMapper, app.feeCollKeeper))

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
	tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// EndBlocker signals the end of a block. It performs application updates on
// the end of every block.
func (app *EthermintApp) EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock) abci.ResponseEndBlock {
	tags := gov.EndBlocker(ctx, app.govKeeper)
	validatorUpdates := stake.EndBlocker(ctx, app.stakeKeeper)

	app.slashingKeeper.AddValidators(ctx, validatorUpdates)

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Tags:             tags,
	}
}

// initChainer initializes the application blockchain with validators and other
// state data from TendermintCore.
func (app *EthermintApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	stateJSON := req.AppStateBytes

	err := app.codec.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(errors.Wrap(err, "failed to parse application genesis state"))
	}

	// load the genesis accounts
	for _, genAcc := range genesisState.Accounts {
		acc := genAcc.ToAccount()
		acc.AccountNumber = app.accountMapper.GetNextAccountNumber(ctx)
		app.accountMapper.SetAccount(ctx, acc)
	}

	// load the genesis stake information
	validators, err := stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)
	if err != nil {
		panic(errors.Wrap(err, "failed to initialize genesis validators"))
	}

	slashing.InitGenesis(ctx, app.slashingKeeper, genesisState.StakeData)
	gov.InitGenesis(ctx, app.govKeeper, genesisState.GovData)

	return abci.ResponseInitChain{
		Validators: validators,
	}
}

// CreateCodec creates a new amino wire codec and registers all the necessary
// concrete types and interfaces needed for the application.
func CreateCodec() *wire.Codec {
	codec := wire.NewCodec()

	types.RegisterWire(codec)
	auth.RegisterWire(codec)
	gov.RegisterWire(codec)
	slashing.RegisterWire(codec)
	stake.RegisterWire(codec)
	wire.RegisterCrypto(codec)

	return codec
}

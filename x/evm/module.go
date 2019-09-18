package evm

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ethermint/x/evm/client/cli"
	"github.com/cosmos/ethermint/x/evm/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var _ module.AppModuleBasic = AppModuleBasic{}
var _ module.AppModule = AppModule{}

// app module Basics object
type AppModuleBasic struct{}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// Validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	// Once json successfully marshalled, passes along to genesis.go
	return ValidateGenesis(data)
}

// Register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	//rpc.RegisterRoutes(ctx, rtr, StoreKey)
}

// Get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.ModuleName, cdc)
}

// Get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(types.ModuleName, cdc)
}

type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// NewAppModule creates a new AppModule Object
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

func (AppModule) Name() string {
	return types.ModuleName
}

func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func (am AppModule) Route() string {
	return types.RouterKey
}

func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}
func (am AppModule) QuerierRoute() string {
	return types.ModuleName
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// TODO: Commit database here ?
	return []abci.ValidatorUpdate{}
}

func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

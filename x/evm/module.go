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

// AppModuleBasic struct
type AppModuleBasic struct{}

// Name for app module basic
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers types for module
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis is json default structure
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis is the validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	// Once json successfully marshalled, passes along to genesis.go
	return ValidateGenesis(data)
}

// RegisterRESTRoutes Registers rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	//rpc.RegisterRoutes(ctx, rtr, StoreKey)
}

// GetQueryCmd Gets the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.ModuleName, cdc)
}

// GetTxCmd Gets the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(types.ModuleName, cdc)
}

// AppModule is struct that defines variables used within module
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

// Name is module name
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants interface for registering invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route specifies path for transactions
func (am AppModule) Route() string {
	return types.RouterKey
}

// NewHandler sets up a new handler for module
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute sets up path for queries
func (am AppModule) QuerierRoute() string {
	return types.ModuleName
}

// NewQuerierHandler sets up new querier handler for module
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// BeginBlock function for module at start of each block
func (am AppModule) BeginBlock(ctx sdk.Context, bl abci.RequestBeginBlock) {
	// Consider removing this when using evm as module without web3 API
	am.keeper.SetBlockHashMapping(ctx, bl.Header.LastBlockId.GetHash(), bl.Header.GetHeight()-1)
	am.keeper.txCount.reset()
}

// EndBlock function for module at end of block
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	_, err := am.keeper.csdb.Commit(true)
	if err != nil {
		panic(err)
	}

	return []abci.ValidatorUpdate{}
}

// InitGenesis instantiates the genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, am.keeper, genesisState)
}

// ExportGenesis exports the genesis state to be used by daemon
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

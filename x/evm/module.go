package evm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/tharsis/ethermint/x/evm/client/cli"
	"github.com/tharsis/ethermint/x/evm/keeper"
	"github.com/tharsis/ethermint/x/evm/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the evm module.
type AppModuleBasic struct{}

// Name returns the evm module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec performs a no-op as the evm module doesn't support amino.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return 1
}

// DefaultGenesis returns default genesis state as raw bytes for the evm
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis is the validation check of the Genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genesisState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesisState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genesisState.Validate()
}

// RegisterRESTRoutes performs a no-op as the EVM module doesn't expose REST
// endpoints
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the evm module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the evm module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the evm module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// ____________________________________________________________________________

// AppModule implements an application module for the evm module.
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
	ak     types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k *keeper.Keeper, ak types.AccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		ak:             ak,
	}
}

// Name returns the evm module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants interface for registering invariants. Performs a no-op
// as the evm module doesn't expose invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// Invariats lead to performance degradation
	//
	// keeper.RegisterInvariants(ir, *am.keeper)
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// Route returns the message routing key for the evm module.
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper))
}

// QuerierRoute returns the evm module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// LegacyQuerierHandler returns nil as the evm module doesn't expose a legacy
// Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// BeginBlock returns the begin block for the evm module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	am.keeper.BeginBlock(ctx, req)
}

// EndBlock returns the end blocker for the evm module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return am.keeper.EndBlock(ctx, req)
}

// InitGenesis performs genesis initialization for the evm module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, am.ak, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the evm
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper, am.ak)
	return cdc.MustMarshalJSON(gs)
}

// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package feemarket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/mezo-org/mezod/x/feemarket/client/cli"
	"github.com/mezo-org/mezod/x/feemarket/keeper"
	"github.com/mezo-org/mezod/x/feemarket/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the fee market module.
type AppModuleBasic struct{}

// Name returns the fee market module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec performs a no-op as the fee market module doesn't support amino.
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// ConsensusVersion returns the consensus state-breaking version for the module.
func (AppModuleBasic) ConsensusVersion() uint64 {
	return 4
}

// DefaultGenesis returns default genesis state as raw bytes for the fee market
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
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {
}

func (b AppModuleBasic) RegisterGRPCGatewayRoutes(c client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(c)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the fee market module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command for the fee market module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces registers interfaces and implementations of the fee market module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// ____________________________________________________________________________

// AppModule implements an application module for the fee market module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	// legacySubspace is used solely for migration of x/params managed parameters
	legacySubspace types.Subspace
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.Keeper, ss types.Subspace) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		legacySubspace: ss,
	}
}

// Name returns the fee market module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants interface for registering invariants. Performs a no-op
// as the fee market module doesn't expose invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterServices registers the GRPC query service and migrator service to respond to the
// module-specific GRPC queries and handle the upgrade store migration for the module.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	types.RegisterMsgServer(cfg.MsgServer(), &am.keeper)

	m := keeper.NewMigrator(am.keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(types.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(err)
	}
}

// BeginBlock returns the begin block for the fee market module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return am.keeper.BeginBlock(ctx)
}

// EndBlock returns the end blocker for the fee market module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx context.Context) error {
	return am.keeper.EndBlock(ctx)
}

// InitGenesis performs genesis initialization for the fee market module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the fee market
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// RegisterStoreDecoder registers a decoder for fee market module's types
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

// GenerateGenesisState creates a randomized GenState of the fee market module.
func (AppModule) GenerateGenesisState(_ *module.SimulationState) {
}

// WeightedOperations returns the all the fee market module operations with their respective weights.
func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

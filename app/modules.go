package app

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparams "github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/skip-mev/connect/v2/x/marketmap"
	marketmapkeeper "github.com/skip-mev/connect/v2/x/marketmap/keeper"
	marketmaptypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/oracle"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"
	"google.golang.org/grpc"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// `AppModule`` wrappers provide a means to override sdk module methods

// ConsensusParams wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/consensus/module.go#L62
type WrappedConsensusParamsAppModule struct {
	consensusparams.AppModule

	keeper consensusparamskeeper.Keeper
}

func NewConsensusParamsAppModule(cdc codec.Codec, keeper consensusparamskeeper.Keeper) WrappedConsensusParamsAppModule {
	return WrappedConsensusParamsAppModule{
		consensusparams.NewAppModule(cdc, keeper),
		keeper,
	}
}

// ConsensusParams method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `consensusparamstypes.RegisterMsgServer()`
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/consensus/module.go#L76
func (am WrappedConsensusParamsAppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	consensusparamstypes.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// Auth wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/module.go#L86
type WrappedAuthAppModule struct {
	auth.AppModule

	accountKeeper     authkeeper.AccountKeeper
	randGenAccountsFn authtypes.RandomGenesisAccountsFn
	legacySubspace    paramstypes.Subspace
}

func NewAuthAppModule(cdc codec.Codec, ak authkeeper.AccountKeeper, randGenAccountsFn authtypes.RandomGenesisAccountsFn, ss paramstypes.Subspace) WrappedAuthAppModule {
	return WrappedAuthAppModule{
		auth.NewAppModule(cdc, ak, randGenAccountsFn, ss),
		ak,
		randGenAccountsFn,
		ss,
	}
}

// Auth method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `authtypes.RegisterMsgServer()`
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/module.go#L115
func (am WrappedAuthAppModule) RegisterServices(cfg module.Configurator) {
	authtypes.RegisterQueryServer(cfg.QueryServer(), authkeeper.NewQueryServer(am.accountKeeper))

	m := authkeeper.NewMigrator(am.accountKeeper, cfg.QueryServer(), am.legacySubspace)
	if err := cfg.RegisterMigration(authtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", authtypes.ModuleName, err))
	}

	if err := cfg.RegisterMigration(authtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", authtypes.ModuleName, err))
	}

	if err := cfg.RegisterMigration(authtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 3 to 4: %v", authtypes.ModuleName, err))
	}

	if err := cfg.RegisterMigration(authtypes.ModuleName, 4, m.Migrate4To5); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 4 to 5", authtypes.ModuleName))
	}
}

// Bank wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/module.go#L99
type WrappedBankAppModule struct {
	bank.AppModule

	keeper         bankkeeper.Keeper
	accountKeeper  banktypes.AccountKeeper
	legacySubspace paramstypes.Subspace
}

func NewBankAppModule(cdc codec.Codec, keeper bankkeeper.Keeper, ak banktypes.AccountKeeper, ss paramstypes.Subspace) WrappedBankAppModule {
	return WrappedBankAppModule{
		bank.NewAppModule(cdc, keeper, ak, ss),
		keeper,
		ak,
		ss,
	}
}

// Bank MsgServer wrapper
// Override all methods for clarity.
// Call original method if required, return error if not
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/keeper/msg_server.go
type RestrictedBankMsgServer struct {
	banktypes.MsgServer
}

// Required (btctoken), call original method
func (ms RestrictedBankMsgServer) UpdateParams(ctx context.Context, msg *banktypes.MsgUpdateParams) (*banktypes.MsgUpdateParamsResponse, error) {
	return ms.MsgServer.UpdateParams(ctx, msg)
}

// Required (btctoken), call original method
func (ms RestrictedBankMsgServer) Send(ctx context.Context, msg *banktypes.MsgSend) (*banktypes.MsgSendResponse, error) {
	return ms.MsgServer.Send(ctx, msg)
}

// disable
func (ms RestrictedBankMsgServer) MultiSend(_ context.Context, _ *banktypes.MsgMultiSend) (*banktypes.MsgMultiSendResponse, error) {
	return nil, fmt.Errorf("method is disabled")
}

// disable
func (ms RestrictedBankMsgServer) SetSendEnabled(_ context.Context, _ *banktypes.MsgSetSendEnabled) (*banktypes.MsgSetSendEnabledResponse, error) {
	return nil, fmt.Errorf("method is disabled")
}

// Bank method overrides
// Override the `RegisterServices` method, replicating the original, except
// with a wrapped MsgServer disabling unused methods
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/module.go#L116
func (am WrappedBankAppModule) RegisterServices(cfg module.Configurator) {
	msgServer := bankkeeper.NewMsgServerImpl(am.keeper)
	banktypes.RegisterMsgServer(cfg.MsgServer(), RestrictedBankMsgServer{msgServer})
	banktypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := bankkeeper.NewMigrator(am.keeper.(bankkeeper.BaseKeeper), am.legacySubspace)
	if err := cfg.RegisterMigration(banktypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(banktypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(banktypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 3 to 4: %v", err))
	}
}

// Crisis wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/crisis/module.go#L87
type WrappedCrisisAppModule struct {
	crisis.AppModule

	keeper                *crisiskeeper.Keeper
	legacySubspace        paramstypes.Subspace
	skipGenesisInvariants bool
}

func NewCrisisAppModule(keeper *crisiskeeper.Keeper, skipGenesisInvariants bool, ss paramstypes.Subspace) WrappedCrisisAppModule {
	return WrappedCrisisAppModule{
		crisis.NewAppModule(keeper, skipGenesisInvariants, ss),
		keeper,
		ss,
		skipGenesisInvariants,
	}
}

// Crisis method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `crisistypes.RegisterMsgServer()`
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/crisis/module.go#L127
func (am WrappedCrisisAppModule) RegisterServices(cfg module.Configurator) {
	m := crisiskeeper.NewMigrator(am.keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(crisistypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", crisistypes.ModuleName, err))
	}
}

// Upgrade wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/upgrade/module.go#L94
type WrappedUpgradeAppModule struct {
	upgrade.AppModule

	keeper *upgradekeeper.Keeper
}

func NewUpgradeAppModule(keeper *upgradekeeper.Keeper, ac address.Codec) WrappedUpgradeAppModule {
	return WrappedUpgradeAppModule{
		upgrade.NewAppModule(keeper, ac),
		keeper,
	}
}

// Upgrade method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `upgradetypes.RegisterMsgServer()`
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/upgrade/module.go#L114
func (am WrappedUpgradeAppModule) RegisterServices(cfg module.Configurator) {
	upgradetypes.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := upgradekeeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(upgradetypes.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", upgradetypes.ModuleName, err))
	}
}

// Params wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/params/module.go#L59
type WrappedParamsAppModule struct {
	params.AppModule

	keeper paramskeeper.Keeper
}

func NewParamsAppModule(keeper paramskeeper.Keeper) WrappedParamsAppModule {
	return WrappedParamsAppModule{
		params.NewAppModule(keeper),
		keeper,
	}
}

// Params method overrides
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/params/module.go#L83
func (am WrappedParamsAppModule) RegisterServices(cfg module.Configurator) {
	// Nothing to override (no MsgServer registration). Call original method
	am.AppModule.RegisterServices(cfg)
}

// Authz wrapper
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/authz/module/module.go#L102
type WrappedAuthzAppModule struct {
	authzmodule.AppModule

	keeper        authzkeeper.Keeper
	accountKeeper authz.AccountKeeper
	bankKeeper    authz.BankKeeper
	registry      cdctypes.InterfaceRegistry
}

func NewAuthzAppModule(cdc codec.Codec, keeper authzkeeper.Keeper, ak authz.AccountKeeper, bk authz.BankKeeper, registry cdctypes.InterfaceRegistry) WrappedAuthzAppModule {
	return WrappedAuthzAppModule{
		authzmodule.NewAppModule(cdc, keeper, ak, bk, registry),
		keeper,
		ak,
		bk,
		registry,
	}
}

// Authz method overrides
// Override the `RegisterServices` method, replicating the original
// TODO: determine what authz.MsgServer methods are required, disable the rest
// See: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/authz/module/module.go#L52
func (am WrappedAuthzAppModule) RegisterServices(cfg module.Configurator) {
	authz.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	authz.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	m := authzkeeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(authz.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", authz.ModuleName, err))
	}
}

// MarketMap Wrapper
// See: https://github.com/skip-mev/connect/blob/v2.1.2/x/marketmap/module.go#L105
type WrappedMarketMapAppModule struct {
	marketmap.AppModule
	k *marketmapkeeper.Keeper
}

func NewMarketMapAppModule(cdc codec.Codec, keeper *marketmapkeeper.Keeper) WrappedMarketMapAppModule {
	return WrappedMarketMapAppModule{
		marketmap.NewAppModule(cdc, keeper),
		keeper,
	}
}

// MarketMap method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `marketmaptypes.RegisterMsgServer()`
// See: https://github.com/skip-mev/connect/blob/v2.1.2/x/marketmap/module.go#L140
func (am WrappedMarketMapAppModule) RegisterServices(cfg module.Configurator) {
	marketmaptypes.RegisterQueryServer(cfg.QueryServer(), marketmapkeeper.NewQueryServer(am.k))
}

// Oracle Wrapper
// See: https://github.com/skip-mev/connect/blob/v2.1.2/x/oracle/module.go#L79
type WrappedOracleAppModule struct {
	oracle.AppModule

	k oraclekeeper.Keeper
}

func NewOracleAppModule(cdc codec.Codec, keeper oraclekeeper.Keeper) WrappedOracleAppModule {
	return WrappedOracleAppModule{
		oracle.NewAppModule(cdc, keeper),
		keeper,
	}
}

// Oracle method overrides
// Override the `RegisterServices` method, replicating the original, minus
// the call to `oracletypes.RegisterMsgServer()`
// See: https://github.com/skip-mev/connect/blob/v2.1.2/x/oracle/module.go#L115
func (am WrappedOracleAppModule) RegisterServices(cfg module.Configurator) {
	oracletypes.RegisterQueryServer(cfg.QueryServer(), oraclekeeper.NewQueryServer(am.k))
}

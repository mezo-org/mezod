package staking

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	upstream "github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/evmos/v12/x/staking/wrapper"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModule struct {
	upstream.AppModule
	upstream.AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

type AppModuleBasic struct {
	upstream.AppModuleBasic
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak types.AccountKeeper, bk types.BankKeeper) AppModule {
	am := upstream.NewAppModule(cdc, keeper, ak, bk)
	return AppModule{
		AppModule:      am,
		AppModuleBasic: am.AppModuleBasic,
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
	}
}

func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Use our msgServer wrapper
	ms := wrapper.NewWrappedMsgServer(keeper.NewMsgServerImpl(am.keeper))
	types.RegisterMsgServer(cfg.MsgServer(), ms)
	querier := keeper.Querier{Keeper: am.keeper}
	types.RegisterQueryServer(cfg.QueryServer(), querier)

	m := keeper.NewMigrator(am.keeper)
	cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3)
}

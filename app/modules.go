package app

import (
	"fmt"

	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	consensusparams "github.com/cosmos/cosmos-sdk/x/consensus"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/mezo-org/mezod/x/bridge"
	"github.com/mezo-org/mezod/x/evm"
	"github.com/mezo-org/mezod/x/feemarket"
	"github.com/mezo-org/mezod/x/poa"
	"github.com/skip-mev/connect/v2/x/marketmap"
	"github.com/skip-mev/connect/v2/x/oracle"
	"google.golang.org/grpc"
)

// The AppModule wrappers provide a means to override module methods

// SDK modules

// ConsensusParams wrapper
type WrappedConsensusParamsAppModule struct {
	consensusparams.AppModule
}

func WrapConsensusParamsAppModule(am consensusparams.AppModule) WrappedConsensusParamsAppModule {
	return WrappedConsensusParamsAppModule{
		am,
	}
}

// ConsensusParams method overrides
func (am WrappedConsensusParamsAppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	fmt.Printf("%s: Registering Services\n", am.Name())
	return am.AppModule.RegisterServices(registrar)
}

// Auth wrapper
type WrappedAuthAppModule struct {
	auth.AppModule
}

func WrapAuthAppModule(am auth.AppModule) WrappedAuthAppModule {
	return WrappedAuthAppModule{
		am,
	}
}

// Auth method overrides
func (am WrappedAuthAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// bank wrapper
type WrappedBankAppModule struct {
	bank.AppModule
}

func WrapBankAppModule(am bank.AppModule) WrappedBankAppModule {
	return WrappedBankAppModule{
		am,
	}
}

// bank method overrides
func (am WrappedBankAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Crisis wrapper
type WrappedCrisisAppModule struct {
	crisis.AppModule
}

func WrapCrisisAppModule(am crisis.AppModule) WrappedCrisisAppModule {
	return WrappedCrisisAppModule{
		am,
	}
}

// Crisis method overrides
func (am WrappedCrisisAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Upgrade wrapper
type WrappedUpgradeAppModule struct {
	upgrade.AppModule
}

func WrapUpgradeAppModule(am upgrade.AppModule) WrappedUpgradeAppModule {
	return WrappedUpgradeAppModule{
		am,
	}
}

// Upgrade method overrides
func (am WrappedUpgradeAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Params wrapper
type WrappedParamsAppModule struct {
	params.AppModule
}

func WrapParamsAppModule(am params.AppModule) WrappedParamsAppModule {
	return WrappedParamsAppModule{
		am,
	}
}

// Params method overrides
func (am WrappedParamsAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Authz wrapper
type WrappedAuthzAppModule struct {
	authz.AppModule
}

func WrapAuthzAppModule(am authz.AppModule) WrappedAuthzAppModule {
	return WrappedAuthzAppModule{
		am,
	}
}

// Authz method overrides
func (am WrappedAuthzAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Mezo modules
// Note: These will likely not be needed as we can make changes on the module side

// Poa Wrapper
type WrappedPoaAppModule struct {
	poa.AppModule
}

func WrapPoaAppModule(am poa.AppModule) WrappedPoaAppModule {
	return WrappedPoaAppModule{
		am,
	}
}

// Poa method overrides
func (am WrappedPoaAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Evm Wrapper
type WrappedEvmAppModule struct {
	evm.AppModule
}

func WrapEvmAppModule(am evm.AppModule) WrappedEvmAppModule {
	return WrappedEvmAppModule{
		am,
	}
}

// Evm method overrides
func (am WrappedEvmAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// FeeMarket Wrapper
type WrappedFeeMarketAppModule struct {
	feemarket.AppModule
}

func WrapFeeMarketAppModule(am feemarket.AppModule) WrappedFeeMarketAppModule {
	return WrappedFeeMarketAppModule{
		am,
	}
}

// FeeMarket method overrides
func (am WrappedFeeMarketAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Bridge Wrapper
type WrappedBridgeAppModule struct {
	bridge.AppModule
}

func WrapBridgeAppModule(am bridge.AppModule) WrappedBridgeAppModule {
	return WrappedBridgeAppModule{
		am,
	}
}

// Bridge method overrides
func (am WrappedBridgeAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// MarketMap Wrapper
type WrappedMarketMapAppModule struct {
	marketmap.AppModule
}

func WrapMarketMapAppModule(am marketmap.AppModule) WrappedMarketMapAppModule {
	return WrappedMarketMapAppModule{
		am,
	}
}

// MarketMap method overrides
func (am WrappedMarketMapAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

// Oracle Wrapper
type WrappedOracleAppModule struct {
	oracle.AppModule
}

func WrapOracleAppModule(am oracle.AppModule) WrappedOracleAppModule {
	return WrappedOracleAppModule{
		am,
	}
}

// Oracle method overrides
func (am WrappedOracleAppModule) RegisterServices(cfg module.Configurator) {
	fmt.Printf("%s: Registering Services\n", am.Name())
	am.AppModule.RegisterServices(cfg)
}

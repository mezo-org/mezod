package upgrades

import (
	store "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	feemarketkeeper "github.com/mezo-org/mezod/x/feemarket/keeper"
	poakeeper "github.com/mezo-org/mezod/x/poa/keeper"
	marketmapkeeper "github.com/skip-mev/connect/v2/x/marketmap/keeper"
	oraclekeeper "github.com/skip-mev/connect/v2/x/oracle/keeper"
)

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v2`.
	UpgradeName string
	// CreateUpgradeHandler defines the function that creates an upgrade handler.
	CreateUpgradeHandler func(*module.Manager, module.Configurator, *Keepers) upgradetypes.UpgradeHandler
	// StoreUpgrades, should be used for any new modules introduced,
	// new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}

// Fork defines a struct containing the requisite fields for hard-fork
// upgrade proposal. The one-time code that should be triggered at the start
// of the Fork, should be defined in `BeginForkLogic`.
type Fork struct {
	// Upgrade version name, for the upgrade handler, e.g. `v2`.
	UpgradeName string
	// UpgradeHeight the upgrade occurs at.
	UpgradeHeight func(chainID string) int64
	// BeginForkLogic runs some custom state transition code at the
	// beginning of a fork.
	BeginForkLogic func(sdk.Context, *Keepers)
}

// Keepers defines a set of keepers exposed by the app.
type Keepers struct {
	ConsensusParamsKeeper consensusparamskeeper.Keeper
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	PoaKeeper             poakeeper.Keeper
	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvmKeeper             *evmkeeper.Keeper
	FeeMarketKeeper       feemarketkeeper.Keeper
	BridgeKeeper          bridgekeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper
	MarketMapKeeper       marketmapkeeper.Keeper
}

package app

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/app/upgrades/v0_3"
	"github.com/mezo-org/mezod/app/upgrades/v0_4"
	"github.com/mezo-org/mezod/app/upgrades/v0_5"
	"github.com/mezo-org/mezod/app/upgrades/v0_6"
	"github.com/mezo-org/mezod/app/upgrades/v0_7"
)

var (
	Upgrades = []upgrades.Upgrade{v0_3.Upgrade}
	Forks    = []upgrades.Fork{v0_3.Fork, v0_4.Fork, v0_5.Fork, v0_6.Fork, v0_7.Fork}
)

// BeginBlockForks is intended to be run in a chain upgrade.
func (app *Mezo) beginBlockForks(ctx sdk.Context) {
	for _, fork := range Forks {
		if ctx.BlockHeight() == fork.UpgradeHeight(ctx.ChainID()) {
			fork.BeginForkLogic(ctx, app.GetKeepers())
			return
		}
	}
}

func (app *Mezo) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				app.GetKeepers(),
			),
		)
	}

	app.setupUpgradeStoreLoaders()
}

func (app *Mezo) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	currentHeight := app.CommitMultiStore().LastCommitID().Version

	if upgradeInfo.Height == currentHeight+1 {
		app.customPreUpgradeHandler(upgradeInfo)
	}

	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			storeUpgrades := upgrade.StoreUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
		}
	}
}

//nolint:all
func (app *Mezo) customPreUpgradeHandler(upgradeInfo upgradetypes.Plan) {
	switch upgradeInfo.Name {
	default:
		// no-op
		return
	}
}

func (app *Mezo) GetKeepers() *upgrades.Keepers {
	return &upgrades.Keepers{
		ConsensusParamsKeeper: app.ConsensusParamsKeeper,
		AccountKeeper:         app.AccountKeeper,
		BankKeeper:            app.BankKeeper,
		PoaKeeper:             app.PoaKeeper,
		CrisisKeeper:          app.CrisisKeeper,
		UpgradeKeeper:         app.UpgradeKeeper,
		ParamsKeeper:          app.ParamsKeeper,
		AuthzKeeper:           app.AuthzKeeper,
		EvmKeeper:             app.EvmKeeper,
		FeeMarketKeeper:       app.FeeMarketKeeper,
		BridgeKeeper:          app.BridgeKeeper,
		OracleKeeper:          app.OracleKeeper,
		MarketMapKeeper:       app.MarketMapKeeper,
	}
}

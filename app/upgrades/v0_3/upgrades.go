//nolint:revive,stylecheck
package v0_3

import (
	"context"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/types"
	"golang.org/x/exp/maps"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.Keepers,
) upgradetypes.UpgradeHandler {
	return func(
		ctx context.Context,
		_ upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		// Run module migrations to trigger initGenesis for x/marketmap and
		// x/oracle modules of the Connect oracle.
		migrations, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Add specific markets to the chain state.
		err = setupOracleMarkets(ctx, keepers)
		if err != nil {
			return nil, err
		}

		return migrations, nil
	}
}

func setupOracleMarkets(ctx context.Context, keepers *upgrades.Keepers) error {
	markets := types.MezoMarketMap.Markets

	// Ensure deterministic order of markets in the upgrade. This is a must
	// so all nodes get the same ID for the same currency pair.
	marketsKeys := maps.Keys(markets)
	slices.Sort(marketsKeys)

	for _, marketKey := range marketsKeys {
		// Create the market.
		market := markets[marketKey]
		err := keepers.MarketMapKeeper.CreateMarket(ctx, market)
		if err != nil {
			return err
		}

		// Invoke hooks to sync the market to x/oracle.
		err = keepers.MarketMapKeeper.Hooks().AfterMarketCreated(
			sdk.UnwrapSDKContext(ctx),
			market,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

//nolint:revive,stylecheck
package v2_0

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
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
		sdk.UnwrapSDKContext(ctx).Logger().Info("running v2.0.0 upgrade handler")

		// There are no particular state upgrades that must be executed globally during this upgrade.
		// The RunMigrations call may still execute module-specific state upgrades. This is not the
		// case for this upgrade but it's a Cosmos SDK requirement to call it here.

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

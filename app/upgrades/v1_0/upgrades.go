//nolint:revive,stylecheck
package v1_0

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
)

func CreateUpgradeHandlerRC0(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.Keepers,
) upgradetypes.UpgradeHandler {
	return func(
		ctx context.Context,
		_ upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		sdk.UnwrapSDKContext(ctx).Logger().Info("running v1.0.0-rc0 upgrade handler")

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func CreateUpgradeHandlerRC1(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *upgrades.Keepers,
) upgradetypes.UpgradeHandler {
	return func(
		ctx context.Context,
		_ upgradetypes.Plan,
		fromVM module.VersionMap,
	) (module.VersionMap, error) {
		sdk.UnwrapSDKContext(ctx).Logger().Info("running v1.0.0-rc1 upgrade handler")

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

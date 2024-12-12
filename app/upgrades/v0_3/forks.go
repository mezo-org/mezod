//nolint:revive,stylecheck
package v0_3

import (
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	plan := upgradetypes.Plan{
		Name:   UpgradeName,
		Height: ctx.BlockHeight(),
		Info:   UpgradeInfo,
	}

	err := keepers.UpgradeKeeper.ScheduleUpgrade(ctx, plan)
	if err != nil {
		panic(
			fmt.Errorf(
				"failed to schedule upgrade %s during fork at height %d: %w",
				plan.Name,
				ctx.BlockHeight(),
				err,
			),
		)
	}
}

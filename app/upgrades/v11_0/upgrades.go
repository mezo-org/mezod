//nolint:revive,stylecheck
package v11_0

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
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
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		sdkCtx.Logger().Info("running v11.0.0 upgrade handler")

		if err := setPragueAndOsakaTimes(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to set Prague/Osaka times: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// setPragueAndOsakaTimes activates the Prague and Osaka EVM forks at the
// upgrade block's timestamp. Both are flipped in the same handler because
// Mezo skipped the Cancun→Prague rollout and collapses Pectra and Fusaka
// into a single chain halt.
func setPragueAndOsakaTimes(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin Prague/Osaka time update",
		"pragueTime", params.ChainConfig.PragueTime,
		"osakaTime", params.ChainConfig.OsakaTime,
	)

	forkTime := sdkmath.NewIntFromUint64(uint64(ctx.BlockTime().Unix())) //nolint:gosec
	params.ChainConfig.PragueTime = &forkTime
	params.ChainConfig.OsakaTime = &forkTime

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	updated := evmKeeper.GetParams(ctx).ChainConfig
	ctx.Logger().Info(
		"Prague/Osaka times updated",
		"pragueTime", updated.PragueTime,
		"osakaTime", updated.OsakaTime,
	)

	return nil
}

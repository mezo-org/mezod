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

		if err := setPragueTime(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to set Prague time: %w", err)
		}

		if err := setOsakaTime(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to set Osaka time: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func setPragueTime(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin Prague time update",
		"pragueTime", params.ChainConfig.PragueTime,
	)

	pragueTime := sdkmath.NewIntFromUint64(uint64(ctx.BlockTime().Unix())) //nolint:gosec
	params.ChainConfig.PragueTime = &pragueTime

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info(
		"Prague time updated",
		"pragueTime", evmKeeper.GetParams(ctx).ChainConfig.PragueTime,
	)

	return nil
}

func setOsakaTime(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin Osaka time update",
		"osakaTime", params.ChainConfig.OsakaTime,
	)

	osakaTime := sdkmath.NewIntFromUint64(uint64(ctx.BlockTime().Unix())) //nolint:gosec
	params.ChainConfig.OsakaTime = &osakaTime

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info(
		"Osaka time updated",
		"osakaTime", evmKeeper.GetParams(ctx).ChainConfig.OsakaTime,
	)

	return nil
}

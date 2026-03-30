//nolint:revive,stylecheck
package v9_0

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
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

		sdkCtx.Logger().Info("running v9.0.0 upgrade handler")

		err := updateAssetsBridgePrecompileVersion(sdkCtx, keepers.EvmKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to update assets bridge precompile version: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func updateAssetsBridgePrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin assets bridge precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	assetsBridgeVersionInfoIndex := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			// Compare bytes just in case to avoid any potential issues
			// with string comparison.
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.AssetsBridgePrecompileAddress),
			)
		},
	)

	// 5 is the value of evmtypes.AssetsBridgePrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[assetsBridgeVersionInfoIndex].Version = 5

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		return err
	}

	ctx.Logger().Info(
		"assets bridge precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

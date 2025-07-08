//nolint:revive,stylecheck
package v3_0

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/utils"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	feemarketkeeper "github.com/mezo-org/mezod/x/feemarket/keeper"
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

		sdkCtx.Logger().Info("running v3.0.0 upgrade handler")

		err := updateMinGasPrice(sdkCtx, keepers.FeeMarketKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to update minimum gas price: %w", err)
		}

		err = updateMaintenancePrecompileVersion(sdkCtx, keepers.EvmKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to update maintenance precompile version: %w", err)
		}

		err = updateBridgePrecompileVersion(sdkCtx, keepers.EvmKeeper)
		if err != nil {
			return nil, fmt.Errorf("failed to update bridge precompile version: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func updateMinGasPrice(sdkCtx sdk.Context, feeMarketKeeper feemarketkeeper.Keeper) error {
	chainID := sdkCtx.ChainID()
	params := feeMarketKeeper.GetParams(sdkCtx)

	sdkCtx.Logger().Info(
		"begin minimum gas price update",
		"chainID", chainID,
		"minGasPrice", params.MinGasPrice,
	)

	var newMinGasPrice sdkmath.LegacyDec
	var err error

	//nolint:gocritic
	if utils.IsMainnet(chainID) {
		newMinGasPrice, err = sdkmath.LegacyNewDecFromStr("1300000.000000000000000000")
		if err != nil {
			return err
		}
	} else if utils.IsTestnet(chainID) {
		newMinGasPrice, err = sdkmath.LegacyNewDecFromStr("130.000000000000000000")
		if err != nil {
			return err
		}
	} else {
		panic(fmt.Sprintf("invalid chain ID: %s", chainID))
	}

	params.MinGasPrice = newMinGasPrice
	err = feeMarketKeeper.SetParams(sdkCtx, params)
	if err != nil {
		return err
	}

	sdkCtx.Logger().Info(
		"minimum gas price updated",
		"chainID", chainID,
		"minGasPrice", feeMarketKeeper.GetParams(sdkCtx).MinGasPrice,
	)

	return nil
}

func updateMaintenancePrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin maintenance precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	maintenanceVersionInfoIndex := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			// Compare bytes just in case to avoid any potential issues
			// with string comparison.
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.MaintenancePrecompileAddress),
			)
		},
	)

	// 4 is the value of evmtypes.MaintenancePrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[maintenanceVersionInfoIndex].Version = 4

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		return err
	}

	ctx.Logger().Info(
		"maintenance precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

func updateBridgePrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin bridge precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	bridgeVersionInfoIndex := slices.IndexFunc(
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

	// 3 is the value of evmtypes.AssetsBridgePrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly to
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[bridgeVersionInfoIndex].Version = 3

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		return err
	}

	ctx.Logger().Info(
		"bridge precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

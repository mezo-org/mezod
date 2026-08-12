//nolint:revive,stylecheck
package v13_0

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

		sdkCtx.Logger().Info("running v13.0.0 upgrade handler")

		// Disable the SELFDESTRUCT opcode by enabling the custom EIP.
		if err := enableExtraEIP(sdkCtx, keepers.EvmKeeper, evmtypes.SelfdestructDisableEIP); err != nil {
			return nil, fmt.Errorf("failed to enable EIP %d: %w", evmtypes.SelfdestructDisableEIP, err)
		}

		// Enable the maintenance precompile methods for the SELFDESTRUCT toggle.
		if err := updateMaintenancePrecompileVersion(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to update maintenance precompile version: %w", err)
		}

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}

// enableExtraEIP adds eip to the EVM params' ExtraEIPs if it is not already set.
func enableExtraEIP(ctx sdk.Context, evmKeeper *evmkeeper.Keeper, eip int64) error {
	params := evmKeeper.GetParams(ctx)

	if slices.Contains(params.ExtraEIPs, eip) {
		ctx.Logger().Info("EIP already enabled", "eip", eip)
		return nil
	}

	params.ExtraEIPs = append(params.ExtraEIPs, eip)

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info("EIP enabled", "eip", eip)

	return nil
}

// updateMaintenancePrecompileVersion enables the SELFDESTRUCT toggle methods.
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
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.MaintenancePrecompileAddress),
			)
		},
	)

	// Keep the version literal so future changes to the latest version do not
	// change this upgrade.
	params.PrecompilesVersions[maintenanceVersionInfoIndex].Version = 6

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info(
		"maintenance precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

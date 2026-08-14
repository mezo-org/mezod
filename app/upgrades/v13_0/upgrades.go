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
	bridgekeeper "github.com/mezo-org/mezod/x/bridge/keeper"
	evmkeeper "github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
	poakeeper "github.com/mezo-org/mezod/x/poa/keeper"
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

		// Enable the maintenance precompile methods for the SELFDESTRUCT toggle
		// and the emergency controls.
		if err := UpdateMaintenancePrecompileVersion(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to update maintenance precompile version: %w", err)
		}

		// Retire the pause methods of the assets bridge precompile.
		if err := UpdateAssetsBridgePrecompileVersion(sdkCtx, keepers.EvmKeeper); err != nil {
			return nil, fmt.Errorf("failed to update assets bridge precompile version: %w", err)
		}

		if err := MigrateEmergencyTeam(
			sdkCtx,
			keepers.PoaKeeper,
			keepers.BridgeKeeper,
		); err != nil {
			return nil, fmt.Errorf("failed to migrate emergency team: %w", err)
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

// UpdateMaintenancePrecompileVersion enables the SELFDESTRUCT toggle methods and
// the emergency controls.
func UpdateMaintenancePrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
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

// UpdateAssetsBridgePrecompileVersion retires the pauser and triparty pause
// methods of the assets bridge precompile.
func UpdateAssetsBridgePrecompileVersion(ctx sdk.Context, evmKeeper *evmkeeper.Keeper) error {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin assets bridge precompile version update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	assetsBridgeVersionInfoIndex := slices.IndexFunc(
		params.PrecompilesVersions,
		func(versionInfo *evmtypes.PrecompileVersionInfo) bool {
			return bytes.Equal(
				evmtypes.HexAddressToBytes(versionInfo.PrecompileAddress),
				evmtypes.HexAddressToBytes(evmtypes.AssetsBridgePrecompileAddress),
			)
		},
	)

	// We avoid using the constant directly as it will change in the future.
	params.PrecompilesVersions[assetsBridgeVersionInfoIndex].Version = 6

	if err := evmKeeper.SetParams(ctx, params); err != nil {
		return err
	}

	ctx.Logger().Info(
		"assets bridge precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)

	return nil
}

// MigrateEmergencyTeam grants the emergency team role to the bridge pauser and
// deletes the retired pause state. On mainnet the pauser is the quick technical
// multisig, so the pause capability stays continuous through the upgrade. A
// zero or absent pauser grants nothing. The retired triparty paused flag is
// dropped without a replacement; it is unset on all networks.
func MigrateEmergencyTeam(
	ctx sdk.Context,
	poaKeeper poakeeper.Keeper,
	bridgeKeeper bridgekeeper.Keeper,
) error {
	ctx.Logger().Info("begin emergency team migration")

	pauser := bridgeKeeper.DeleteRetiredPauseState(ctx)

	if !pauser.Empty() && !evmtypes.IsZeroHexAddress(evmtypes.BytesToHexAddress(pauser)) {
		poaKeeper.SetEmergencyTeamUnchecked(ctx, pauser)
		ctx.Logger().Info("emergency team granted", "emergencyTeam", pauser.String())
	} else {
		ctx.Logger().Info("no pauser was set; the emergency team stays unset")
	}

	ctx.Logger().Info("emergency team migration done")

	return nil
}

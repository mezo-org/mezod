//nolint:revive,stylecheck
package v0_7

import (
	"bytes"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	ctx.Logger().Info("running v0.7.0 fork logic")

	updateMaintenancePrecompileVersion(ctx, keepers.EvmKeeper)
	setMaxPrecompilesCallsPerExecution(ctx, keepers.EvmKeeper)

	ctx.Logger().Info("v0.7.0 fork logic applied successfully")
}

func updateMaintenancePrecompileVersion(ctx sdk.Context, evmKeeper *keeper.Keeper) {
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

	// 3 is the value of evmtypes.MaintenancePrecompileLatestVersion at
	// the time of this upgrade. We avoid using the constant directly
	// as it will change in the future so the actual value of the version
	// used during this upgrade would perish over time.
	params.PrecompilesVersions[maintenanceVersionInfoIndex].Version = 3

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the Maintenance precompile version update.
		ctx.Logger().Error(
			"failed to set maintenance precompile version; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"maintenance precompile version updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)
}

func setMaxPrecompilesCallsPerExecution(ctx sdk.Context, evmKeeper *keeper.Keeper) {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin setting max precompile calls per execution",
		"maxPrecompileCallsPerExecution",
		params.MaxPrecompilesCallsPerExecution,
	)

	params.MaxPrecompilesCallsPerExecution = uint32(evmtypes.DefaultMaxPrecompilesCallsPerExecution) //nolint:gosec

	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the max precompile calls per execution update.
		ctx.Logger().Error(
			"failed to set max precompile calls per execution; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"max precompile calls per execution updated",
		"maxPrecompileCallsPerExecution",
		evmKeeper.GetParams(ctx).MaxPrecompilesCallsPerExecution,
	)
}

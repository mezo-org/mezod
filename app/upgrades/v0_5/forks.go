//nolint:revive,stylecheck
package v0_5

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/app/upgrades"
	"github.com/mezo-org/mezod/x/evm/keeper"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

func RunForkLogic(ctx sdk.Context, keepers *upgrades.Keepers) {
	ctx.Logger().Info("running v0.5.0 fork logic")

	updatePrecompilesVersions(ctx, keepers.EvmKeeper)

	ctx.Logger().Info("v0.5.0 fork logic applied successfully")
}

func updatePrecompilesVersions(ctx sdk.Context, evmKeeper *keeper.Keeper) {
	params := evmKeeper.GetParams(ctx)

	ctx.Logger().Info(
		"begin precompiles versions update",
		"precompilesVersions",
		params.PrecompilesVersions,
	)

	params.PrecompilesVersions = evmtypes.DefaultPrecompilesVersions
	err := evmKeeper.SetParams(ctx, params)
	if err != nil {
		// Do not halt consensus in case of failure.
		// Just live without the precompiles versions.
		ctx.Logger().Error(
			"failed to set precompiles versions; abandoning",
			"error",
			err,
		)
		return
	}

	ctx.Logger().Info(
		"precompiles versions updated",
		"precompilesVersions",
		evmKeeper.GetParams(ctx).PrecompilesVersions,
	)
}
